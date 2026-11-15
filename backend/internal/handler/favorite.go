package handler

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/masaprog/go-samurai-store/backend/internal/database"
)

//お気に入り登録用の構造体
type FavoriteDataRequest struct {
	Product_ID       int      `json:"productId" binding:"required"`
}

func ListFavoritesHandler(c *gin.Context) {
	userClaims, exists := c.Get("user")
	if !exists {
		log.Println("GetUserMeHandler: コンテキストにユーザー情報が見つかりません")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証情報が見つかりません"})
		return
	}

	// 型アサーションでクレームを取得する
	claims, ok := userClaims.(*JWTCustomClaims)
	if !ok {
		log.Println("GetUserMeHandler: ユーザー情報の型アサーションに失敗しました")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}
	userID := claims.UserID

	// データベース接続を取得する
	db := database.GetDB()

		query := `
		SELECT
			p.id,
			p.name,
			p.price,
			p.image_url
		FROM favorites f
		JOIN products p ON f.product_id = p.id
		WHERE f.user_id = ?
		ORDER BY f.created_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
			log.Printf("新着商品取得エラー: %v", err)
			// newArrivalErr = err
			return
		}
	defer rows.Close()
	favorites := []ProductListItem{}

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
			&imageUrl,
		); err != nil {
			log.Printf("お気に入り商品データのスキャンエラー: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
			return
		}
		// 変数imageUrlがNULLでない場合にのみ、ProductListItem構造体のImageURLフィールドに値をセットする
		if imageUrl.Valid {
			p.ImageURL = &imageUrl.String
		}
		favorites = append(favorites, p)
	}
	if err = rows.Err(); err != nil {
		log.Printf("お気に入り一覧データ取得中の行エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}
	// JSONとしてレスポンスを返す
		c.JSON(http.StatusOK, favorites)

}

func AddFavoriteHandler(c *gin.Context) {
	userClaims, exists := c.Get("user")
		if !exists {
		log.Println("GetUserMeHandler: コンテキストにユーザー情報が見つかりません")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証情報が見つかりません"})
		return
	}

	// 型アサーションでクレームを取得する
	claims, ok := userClaims.(*JWTCustomClaims)
	if !ok {
		log.Println("GetUserMeHandler: ユーザー情報の型アサーションに失敗しました")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	var req FavoriteDataRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("HTTPリクエストボディのバインドエラー: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "入力内容が正しくありません"})
		return
	}

	db := database.GetDB()
	
	//同じ商品を２回登録してもエラーを出さない
	insertQuery := `
		INSERT IGNORE INTO favorites (user_id, product_id)
		VALUES (?, ?)
	`

		_, err := db.Exec(insertQuery, claims.UserID, req.Product_ID)
	if err != nil {
		log.Printf("お気に入り登録エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "お気に入り登録に失敗しました"})
		return
	}

	// 登録成功のレスポンスを返す
	c.JSON(http.StatusOK, gin.H{"message": "お気に入りに追加しました"})
}

func GetFavoriteStatusHandler(c *gin.Context) {
		userClaims, exists := c.Get("user")
		if !exists {
		log.Println("GetUserMeHandler: コンテキストにユーザー情報が見つかりません")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証情報が見つかりません"})
		return
	}

	idStr := c.Param("productId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("不正な商品ID形式です: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品IDが不正です"})
		return
	}

	// 型アサーションでクレームを取得する
	claims, ok := userClaims.(*JWTCustomClaims)
	if !ok {
		log.Println("GetUserMeHandler: ユーザー情報の型アサーションに失敗しました")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}
	
	db := database.GetDB()

	var count int
	err = db.QueryRow(
    	"SELECT COUNT(*) FROM favorites WHERE user_id = ? AND product_id = ?",
    claims.UserID, id,
	).Scan(&count)
	
	c.JSON(http.StatusOK, gin.H{"isFavorite": count > 0})
}

func RemoveFavoriteHandler(c *gin.Context) {
		userClaims, exists := c.Get("user")
		if !exists {
		log.Println("GetUserMeHandler: コンテキストにユーザー情報が見つかりません")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証情報が見つかりません"})
		return
	}

	// 型アサーションでクレームを取得する
	claims, ok := userClaims.(*JWTCustomClaims)
	if !ok {
		log.Println("GetUserMeHandler: ユーザー情報の型アサーションに失敗しました")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// パスパラメータ（.../products/:idのうち、:idの部分）から「商品ID」を取得する
	idStr := c.Param("productId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("不正な商品ID形式です: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品IDが不正です"})
		return
	}

	db := database.GetDB()

	deleteQuery := `
		DELETE FROM favorites WHERE user_id = ? and product_id = ?
	`

	_, err = db.Exec(deleteQuery, claims.UserID, id)
	if err != nil {
		log.Printf("お気に入り削除エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "お気に入り削除に失敗しました"})
		return
	}

	// 登録成功のレスポンスを返す
	c.JSON(http.StatusOK, gin.H{"message": "お気に入りの削除を完了しました"})

}