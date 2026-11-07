package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/masaprog/go-samurai-store/backend/internal/database"
)

// --- 1. 型定義（struct） ---

// レスポンス用の構造体
type InquiriesData struct {
	ID          int         `json:"id"`
	Name	    string      `json:"name"`
	Email       string      `json:"email"`
	Message     string      `json:"message"` 
	CreatedAt   time.Time 		`json:"created_at"`
}

//リクエスト登録用の構造体
type InquiriesDataRequest struct {
	Name	    string      `json:"name" binding:"required"`
	Email       string      `json:"email" binding:"required"`
	Message     string      `json:"message" binding:"required"`
}

func CreateInquiryHandler(c *gin.Context) {
	var req InquiriesDataRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("HTTPリクエストボディのバインドエラー: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "入力内容が正しくありません"})
		return
	}

	db := database.GetDB()

	insertQuery := `
		INSERT INTO inquiries (name, email, message)
		VALUES (?, ?, ?)
	`

		_, err := db.Exec(insertQuery, req.Name, req.Email, req.Message)
	if err != nil {
		log.Printf("お問い合わせ登録エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "お問い合わせに失敗しました"})
		return
	}

	// 登録成功のレスポンスを返す
	c.JSON(http.StatusOK, gin.H{"message": "お問い合わせの受付を完了しました"})
}

func ListInquiriesHandler(c *gin.Context) {
		db := database.GetDB()
		query := `
		SELECT
			id, name, email, message, created_at
		FROM inquiries
		ORDER BY created_at DESC
		`
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("問い合わせ一覧取得エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}
	defer rows.Close()

	inquiries := []InquiriesData{}

	for rows.Next(){
		var i InquiriesData
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Email,
			&i.Message,
			&i.CreatedAt,
		); err != nil {
			log.Printf("問い合わせ一覧の取得エラー: %v", err)
			return
		}
		inquiries = append(inquiries, i)
	}

	if err = rows.Err(); err != nil {
		log.Printf("問い合わせ一覧データ取得中のエラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーです"})
		return
	}
	// JSONとしてレスポンスを返す
	c.JSON(http.StatusOK, inquiries)

}