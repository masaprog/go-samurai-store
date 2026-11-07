package handler

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/masaprog/go-samurai-store/backend/internal/database"
)

// --- 1. 型定義（struct） ---
// フロントエンドが期待するJSONの「形」をGoの構造体で定義する
// `json:"..."`タグで、Goのフィールド名（大文字）とJSONのキー名（小文字）を対応させる

// 商品一覧用の構造体
type ProductListItem struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Stock	 int	`json:"stock"`
	ImageURL *string `json:"image_url"`
	ReviewAvg   float64   `json:"review_avg"`
	ReviewCount int       `json:"review_count"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 商品詳細用の構造体
type Product struct {
	ID		int		`json:"id"`
	Name	string	`json:"name"`
	Description *string   `json:"description"` // NULL許容
	Price       int       `json:"price"`
	Stock       int       `json:"stock"`
	ImageURL    *string   `json:"image_url"` // NULL許容
	SalesCount  int       `json:"sales_count"`
	IsFeatured  bool      `json:"is_featured"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ページネーション情報を表す構造体
type Pagination struct {
	CurrentPage int `json:"currentPage"`
	PerPage     int `json:"perPage"`
	TotalItems  int `json:"totalItems"`
	TotalPages  int `json:"totalPages"`
}

// 商品一覧ページ全体のレスポンスを表す構造体
type ProductsPageData struct {
	Products   []ProductListItem  `json:"products"`
	Pagination Pagination `json:"pagination"`
}

// トップページ用の構造体
type HomePageProduct struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    int     `json:"price"`
	ImageURL *string `json:"image_url"`
	ReviewAvg   float64 `json:"review_avg"`
	ReviewCount int     `json:"review_count"`
}

// トップページ全体のレスポンスを表す構造体
type HomePageData struct {
	PickUp     []HomePageProduct `json:"pickUp"`
	NewArrival []HomePageProduct `json:"newArrival"`
	HotItems   []HomePageProduct `json:"hotItems"`
}

// --- 2. ハンドラ定義 ---

// 商品一覧を返す関数
func GetProductsHandler(c *gin.Context) {
	// クエリパラメータ（?page=X）から「ページ番号」を取得する
	// クエリパラメータが存在しない、または不正な値の場合はデフォルトで1を設定する
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// クエリパラメータ（?perPage=X）から「1ページあたりの商品数」を取得する
	// クエリパラメータが存在しない、または不正な値の場合はデフォルトで16を設定する
	perPageStr := c.DefaultQuery("perPage", "16")
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage < 1 {
		perPage = 16
	}

		// クエリパラメータ（?sort=X）から「並べ替え条件」を取得する（デフォルトは新着順）
		sort := c.DefaultQuery("sort", "new")

		// 並べ替え条件（ORDER BY句）を構築する
		var orderByClause string
		switch sort {
		case "priceAsc": // 価格が安い順
			orderByClause = "ORDER BY p.price ASC"
		case "new": // 新着順
			orderByClause = "ORDER BY p.created_at DESC"
		default:
			// 上記以外はデフォルトの新着順にする
			orderByClause = "ORDER BY p.created_at DESC"
		}

	// クエリパラメータ（?keyword=X）から「検索キーワード」を取得する
	keyword := c.DefaultQuery("keyword", "")

	// 検索条件（WHERE句）を構築する
	var whereClause string
	var whereParams []any // WHERE句用のパラメータを格納するスライス
	if keyword != "" {
		whereClause = "WHERE (p.name LIKE ? OR p.description LIKE ?)"
		// パラメータ（%キーワード%）をスライスに追加する
		likeKeyword := "%" + keyword + "%"
		whereParams = append(whereParams, likeKeyword, likeKeyword)
	}

	// データベース接続を取得する
	db := database.GetDB()

	// 商品の総数を取得するSQL文を実行する
	var totalItems int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products AS p %s", whereClause)
	// whereParams...でスライスの内容を展開してQueryRow()メソッドに渡す
	err = db.QueryRow(countQuery, whereParams...).Scan(&totalItems)
	if err != nil {
		log.Printf("商品総数の取得エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// ページネーション情報を計算する
	totalPages := int(math.Ceil(float64(totalItems) / float64(perPage)))
	offset := (page - 1) * perPage

	// 商品一覧を取得するSQL文を準備する
	query := fmt.Sprintf(`
			SELECT
				p.id,
				p.name,
				p.price,
				p.stock,
				p.image_url,
				p.updated_at,
				COALESCE(ROUND(AVG(r.score), 1), 0.0) AS review_avg,
				COALESCE(COUNT(r.id), 0) AS review_count
			FROM products AS p
			LEFT JOIN reviews AS r ON p.id = r.product_id
			%s
			Group BY p.id, p.name, p.price, p.stock, p.image_url, p.updated_at
			%s
			LIMIT ? OFFSET ?
		`, whereClause, orderByClause)

	// SQL文に渡すパラメータを準備する
	queryParams := append(whereParams, perPage, offset)

	// SQL文を実行する
	rows, err := db.Query(query, queryParams...)
	if err != nil {
		log.Printf("商品一覧の取得エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}
	defer rows.Close()

	// 商品データを格納するスライス
	products := []ProductListItem{}

	// SQL文の実行結果をスキャンする
	for rows.Next() {
		var p ProductListItem
		// image_urlカラムがNULLの場合に対応するため、sql.NullString型の変数imageUrlを用意する
		var imageUrl sql.NullString
		// 取得したデータをProductListItem構造体のフィールドにマッピングする
		// image_urlカラムは変数imageUrlにスキャンする
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Price,
			&p.Stock,
			&imageUrl,
			&p.UpdatedAt,
			&p.ReviewAvg,
			&p.ReviewCount,
		); err != nil {
			log.Printf("商品データのスキャンエラー: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
			return
		}
		// 変数imageUrlがNULLでない場合にのみ、ProductListItem構造体のImageURLフィールドに値をセットする
		if imageUrl.Valid {
			p.ImageURL = &imageUrl.String
		}
		products = append(products, p)
	}
	if err = rows.Err(); err != nil {
		log.Printf("商品一覧データ取得中の行エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// ページネーション情報を作成する
	pagination := Pagination{
		CurrentPage: page,  // 現在のページ
		PerPage:     perPage, // 1ページあたりの商品数
		TotalItems:  totalItems,  // 全商品数
		TotalPages:  totalPages,  // 総ページ数
	}

	// 最終的なレスポンスの形にまとめる
	response := ProductsPageData{
		Products:   products,
		Pagination: pagination,
	}

	// JSONとしてレスポンスを返す
	c.JSON(http.StatusOK, response)
}

// 商品詳細を返す関数
func GetProductByIDHandler(c *gin.Context) {
	// パスパラメータ（.../products/:idのうち、:idの部分）から「商品ID」を取得する
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("不正な商品ID形式です: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品IDが不正です"})
		return
	}

	// データベース接続を取得する
	db := database.GetDB()

	// 商品詳細を取得するSQL文を準備する
	query := `
		SELECT
			id, name,
			description,
			price,
			stock,
			image_url,
			sales_count,
			is_featured,
			created_at,
			updated_at
		FROM products
		WHERE id = ?
	`
	var p Product
	var description sql.NullString
	var imageUrl sql.NullString

	// SQL文を実行して結果をスキャンする
	err = db.QueryRow(query, id).Scan(
		&p.ID,
		&p.Name,
		&description,
		&p.Price,
		&p.Stock,
		&imageUrl,
		&p.SalesCount,
		&p.IsFeatured,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("商品が見つかりません: ID=%d", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "商品が見つかりませんでした"})
		} else {
			log.Printf("商品取得エラー (ID=%d): %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		}
		return
	}

	// 変数descriptionと変数imageUrlがNULLでない場合にのみ、Product構造体の対応するフィールドに値をセットする
	if description.Valid {
		p.Description = &description.String
	}
	if imageUrl.Valid {
		p.ImageURL = &imageUrl.String
	}

	// JSONとしてレスポンスを返す
	c.JSON(http.StatusOK, p)
}

// トップページ用の商品一覧を返す関数
func GetHomePageProductsHandler(c *gin.Context) {
	// データベース接続を取得する
	db := database.GetDB()

	// 各セクションのデータを格納するスライス
	var pickUp, newArrival, hotItems []HomePageProduct
	// エラーハンドリング用の変数
	var pickUpErr, newArrivalErr, hotItemsErr error
	// Goルーチンの完了を待つためのWaitGroup
	var wg sync.WaitGroup

	// 3つの処理を並行実行するため、カウンタを3つ増やす
	wg.Add(3)

	// Goルーチン1: おすすめ商品（Pick Up）を取得する
	go func() {
		defer wg.Done() // 処理が終わったらカウンタを1つ減らす
		query := `
			SELECT
				id,
				name,
				price,
				image_url
			FROM products
			ORDER BY sales_count DESC
			LIMIT 3
		`
		rows, err := db.Query(query)
		if err != nil {
			log.Printf("おすすめ商品取得エラー: %v", err)
			pickUpErr = err
			return
		}
		defer rows.Close()

		for rows.Next() {
			var p HomePageProduct
			var imageUrl sql.NullString
			if err := rows.Scan(&p.ID, &p.Name, &p.Price, &imageUrl); err != nil {
				log.Printf("おすすめ商品スキャンエラー: %v", err)
				pickUpErr = err
				return
			}
			if imageUrl.Valid {
				p.ImageURL = &imageUrl.String
			}
			p.ReviewAvg = 0.0 // おすすめ商品（Pick Up）では平均評価は不要なので0.0で初期化
			p.ReviewCount = 0 // おすすめ商品（Pick Up）ではレビュー数は不要なので0で初期化
			pickUp = append(pickUp, p)
		}
		pickUpErr = rows.Err()
	}()

	// Goルーチン2: 新着商品（New Arrival）を取得する
	go func() {
		defer wg.Done()
		query := `
			SELECT
				p.id,
				p.name,
				p.price,
				p.image_url,
				COALESCE(ROUND(AVG(r.score), 1), 0.0) AS review_avg,
				COALESCE(COUNT(r.id), 0) AS review_count
			FROM products AS p
			LEFT JOIN reviews AS r ON p.id = r.product_id
			GROUP BY p.id, p.name, p.price, p.image_url, p.created_at
			ORDER BY p.created_at DESC
			LIMIT 4
		`
		rows, err := db.Query(query)
		if err != nil {
			log.Printf("新着商品取得エラー: %v", err)
			newArrivalErr = err
			return
		}
		defer rows.Close()

		for rows.Next() {
			var p HomePageProduct
			var imageUrl sql.NullString
			if err := rows.Scan(
				&p.ID,
				&p.Name,
				&p.Price,
				&imageUrl,
				&p.ReviewAvg,
				&p.ReviewCount,
			); err != nil {				log.Printf("新着商品スキャンエラー: %v", err)
				newArrivalErr = err
				return
			}
			if imageUrl.Valid {
				p.ImageURL = &imageUrl.String
			}
			newArrival = append(newArrival, p)
		}
		newArrivalErr = rows.Err()
	}()

	// Goルーチン3: 注目商品（Hot Items）を取得する
	go func() {
		defer wg.Done()
		// ORDER BY RAND()を使ってランダムに取得する
		query := `
			SELECT
				p.id,
				p.name,
				p.price,
				p.image_url,
				COALESCE(ROUND(AVG(r.score), 1), 0.0) AS review_avg,
				COALESCE(COUNT(r.id), 0) AS review_count
			FROM products AS p
			LEFT JOIN reviews AS r ON p.id = r.product_id
			WHERE p.is_featured = true
			GROUP BY p.id, p.name, p.price, p.image_url
			ORDER BY RAND()
			LIMIT 4
		`
		rows, err := db.Query(query)
		if err != nil {
			log.Printf("注目商品取得エラー: %v", err)
			hotItemsErr = err
			return
		}
		defer rows.Close()

		for rows.Next() {
			var p HomePageProduct
			var imageUrl sql.NullString
			if err := rows.Scan(&p.ID, &p.Name, &p.Price, &imageUrl); err != nil {
				log.Printf("注目商品スキャンエラー: %v", err)
				hotItemsErr = err
				return
			}
			if imageUrl.Valid {
				p.ImageURL = &imageUrl.String
			}
			hotItems = append(hotItems, p)
		}
		hotItemsErr = rows.Err()
	}()

	// すべてのGoルーチンの完了を待つ
	wg.Wait()

	// いずれかのGoルーチンでエラーが発生していた場合は500エラーを返す
	if pickUpErr != nil || newArrivalErr != nil || hotItemsErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// 最終的なレスポンスの形にまとめる
	response := HomePageData{
		PickUp:     pickUp,
		NewArrival: newArrival,
		HotItems:   hotItems,
	}

	// JSONとしてレスポンスを返す
	c.JSON(http.StatusOK, response)
}