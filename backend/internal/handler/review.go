package handler

import (
	"database/sql"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/masaprog/go-samurai-store/backend/internal/database"
)

// --- 1. 型定義（struct） ---

// レビュー表示用の構造体
type ReviewData struct {
	ID        int       `json:"id"`
	ProductID int       `json:"product_id"`
	UserID    int       `json:"user_id"`
	Score     int       `json:"score"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UserName  string    `json:"user_name"`
}

// レビュー一覧API (/api/products/:id/reviews) のレスポンスの構造体
type ReviewsPageData struct {
	Reviews    []ReviewData `json:"reviews"`
	ReviewAvg  float64      `json:"review_avg"` // 平均評価
	Pagination Pagination   `json:"pagination"` // ページネーション情報（product.goファイルで定義した型）
}

// レビュー投稿リクエストの構造体
type ReviewRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"` // 星の数（1-5）
	Content string `json:"content" binding:"required"`
}

// --- 2. ハンドラ定義 ---

// 指定された商品IDのレビュー一覧を取得する関数（GET /api/products/:id/reviews）
func GetReviewsHandler(c *gin.Context) {
	// パスパラメータ（:id）から「商品ID」を取得する
	productIDStr := c.Param("id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		log.Printf("不正な商品ID形式です (レビュー取得): %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品IDが不正です"})
		return
	}

	// クエリパラメータ（?page=X）からページネーション情報を取得する
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// 1ページあたりのレビュー数を設定する
	perPage := 10
	offset := (page - 1) * perPage

	// データベース接続を取得する
	db := database.GetDB()

	// 並行処理でレビュー一覧、レビュー総数、平均評価を取得する
	var reviews []ReviewData
	var totalItems int
	var reviewAvg sql.NullFloat64 // AVG（平均評価）はNULLになる可能性がある
	var reviewsErr, statsErr error

	// Goルーチンの完了を待つためのWaitGroup
	var wg sync.WaitGroup

	wg.Add(2)

	// Goルーチン1: レビュー一覧を取得する
	go func() {
		defer wg.Done()
		query := `
			SELECT
				r.id, r.product_id, r.user_id, r.score, r.content, r.created_at,
				u.name AS user_name
			FROM reviews AS r
			JOIN users AS u ON r.user_id = u.id
			WHERE r.product_id = ?
			ORDER BY r.created_at DESC
			LIMIT ? OFFSET ?
		`
		rows, err := db.Query(query, productID, perPage, offset)
		if err != nil {
			log.Printf("レビュー一覧取得エラー (ProductID=%d): %v", productID, err)
			reviewsErr = err
			return
		}
		defer rows.Close()

		for rows.Next() {
			var r ReviewData
			if err := rows.Scan(
				&r.ID, &r.ProductID, &r.UserID, &r.Score, &r.Content, &r.CreatedAt,
				&r.UserName,
			); err != nil {
				log.Printf("レビューデータスキャンエラー (ProductID=%d): %v", productID, err)
				reviewsErr = err
				return
			}
			reviews = append(reviews, r)
		}
		reviewsErr = rows.Err()
	}()

	// Goルーチン2: レビュー総数と平均評価を取得する
	go func() {
		defer wg.Done()
		query := `
			SELECT COUNT(*), COALESCE(AVG(score), 0.0)
			FROM reviews
			WHERE product_id = ?
		`
		err := db.QueryRow(query, productID).Scan(&totalItems, &reviewAvg)
		if err != nil {
			log.Printf("レビュー統計取得エラー (ProductID=%d): %v", productID, err)
			statsErr = err
		}
	}()

	// すべてのGoルーチンの完了を待つ
	wg.Wait()
	if reviewsErr != nil || statsErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// AVG（平均評価）の結果がNULLでなければ実際の値、NULLなら0.0を使う
	avg := 0.0
	if reviewAvg.Valid {
		avg = reviewAvg.Float64
	}
	// 小数点以下1桁に丸める
	avg = math.Round(avg*10) / 10

	// ページネーション情報を作成する
	totalPages := int(math.Ceil(float64(totalItems) / float64(perPage)))
	pagination := Pagination{ // product.goファイルで定義した型
		CurrentPage: page,
		PerPage:     perPage,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
	}

	// 最終的なレスポンスの形にまとめる
	response := ReviewsPageData{
		Reviews:    reviews,
		ReviewAvg:  avg,
		Pagination: pagination,
	}

	// JSONとしてレスポンスを返す
	c.JSON(http.StatusOK, response)
}

// レビューを新規登録する関数（POST /api/products/:id/reviews）
func CreateReviewHandler(c *gin.Context) {
	// パスパラメータ（:id）から「商品ID」を取得する
	productIDStr := c.Param("id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		log.Printf("不正な商品ID形式です (レビュー投稿): %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品IDが不正です"})
		return
	}

	// HTTPリクエストのコンテキストからユーザー情報を取得する
	userClaims, exists := c.Get("user")
	if !exists {
		log.Println("CreateReviewHandler: ユーザー情報がコンテキストにありません")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "レビューの投稿にはログインが必要です"})
		return
	}
	claims, ok := userClaims.(*JWTCustomClaims) // auth.goファイルで定義した型
	if !ok {
		log.Println("CreateReviewHandler: ユーザー情報の型アサーションに失敗しました")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}
	userID := claims.UserID

	// HTTPリクエストボディをReviewRequest構造体にバインドする
	// 必須項目や形式が異なる場合はエラーを返す
	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("レビュー投稿リクエストのバインドエラー: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "入力内容が正しくありません（評価は1-5, コメント必須）"})
		return
	}
	// 念のためContentフィールド（コメント）の空白チェックも行う
	if strings.TrimSpace(req.Content) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "コメントを入力してください"})
		return
	}

	// データベースにレビューを登録する
	db := database.GetDB()
	query := `
		INSERT INTO reviews (product_id, user_id, score, content)
		VALUES (?, ?, ?, ?)
	`
	_, err = db.Exec(query, productID, userID, req.Rating, req.Content)
	if err != nil {
		log.Printf("レビュー登録エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "レビューの投稿に失敗しました"})
		return
	}

	// 登録成功のレスポンスを返す
	c.JSON(http.StatusCreated, gin.H{"message": "レビューを登録しました"})
}