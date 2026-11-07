package handler

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/masaprog/go-samurai-store/backend/internal/database"
)

// --- ハンドラ定義 ---

// 商品を新規登録する関数
func AdminCreateProductHandler(c *gin.Context) {

	//======== 中略 ========

}

// 商品情報を編集する関数
func AdminUpdateProductHandler(c *gin.Context) {
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

	// 商品が存在するか確認し、既存の画像ファイル名を取得する
	var currentImageUrl sql.NullString
	checkQuery := "SELECT image_url FROM products WHERE id = ?"
	err = db.QueryRow(checkQuery, id).Scan(&currentImageUrl)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("更新対象の商品が見つかりません: ID=%d", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "商品が見つかりませんでした"})
		} else {
			log.Printf("既存商品チェックエラー (ID=%d): %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		}
		return
	}

	// フォームデータを取得する
	name := c.PostForm("name")
	description := c.PostForm("description")
	priceStr := c.PostForm("price")
	stockStr := c.PostForm("stock")
	isFeaturedStr := c.PostForm("isFeatured")

	// 画像ファイルを取得する
	fileHeader, err := c.FormFile("imageFile")
	if err != nil && err != http.ErrMissingFile {
		log.Printf("画像ファイルの取得中に予期せぬエラー: %v", err)
	}
	// 新しいファイルがアップロードされたかどうかのフラグ
	newFileUploaded := fileHeader != nil

	// 入力データのバリデーションを行う
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品名は必須です"})
		return
	}
	price, err := strconv.Atoi(priceStr)
	if err != nil || price < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "価格は0以上の整数で入力してください"})
		return
	}
	stock, err := strconv.Atoi(stockStr)
	if err != nil || stock < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "在庫数は0以上の整数で入力してください"})
		return
	}

	// チェックボックスがONであれば"on"、OFFであれば空文字が送信されてくる
	isFeatured := isFeaturedStr == "on"

	// descriptionが空であればデフォルト値を設定する
	if description == "" {
		description = "商品の説明がありません。"
	}

	// 画像ファイルの保存先パスやファイル名を準備する
	imageUrlToSave := currentImageUrl.String // DBに保存するファイル名（デフォルトは既存のファイル名）
	newFileName := ""                        // 新しく保存したファイル名（後で古いファイルを消すため）
	oldFileName := ""                        // 削除対象の古いファイル名
	if currentImageUrl.Valid {
		oldFileName = currentImageUrl.String
	}

	if newFileUploaded {
		// ファイル形式のバリデーションを行う
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
		if !allowedExts[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "対応していないファイル形式です (jpg, jpeg, pngのみ)"})
			return
		}

		// 新しいファイル名を生成する
		timestamp := time.Now().UnixNano() / int64(time.Millisecond)
		random := rand.Intn(10000)
		newFileName = fmt.Sprintf("%d_%d%s", timestamp, random, ext)
		savePath := filepath.Join("uploads", newFileName)

		// 新しいファイルを保存する
		if err := c.SaveUploadedFile(fileHeader, savePath); err != nil {
			log.Printf("新しい画像ファイルの保存エラー: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ファイルのアップロードに失敗しました"})
			return
		}
		log.Printf("新しい画像ファイルを保存しました: %s", savePath)

		// DBに保存するファイル名を新しいファイル名に更新する
		imageUrlToSave = newFileName
	}

	// データベースを更新する
	updateQuery := `
		UPDATE products SET
			name = ?, description = ?, price = ?, stock = ?, image_url = ?, is_featured = ?
		WHERE id = ?
	`
	_, err = db.Exec(
		updateQuery,
		name,
		description,
		price,
		stock,
		sql.NullString{String: imageUrlToSave, Valid: imageUrlToSave != ""},
		isFeatured,
		id,
	)
	if err != nil {
		log.Printf("商品更新エラー (ID=%d): %v", id, err)
		// 新しく保存したファイルを削除する
		if newFileName != "" {
			_ = os.Remove(filepath.Join("uploads", newFileName))
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// 古い画像ファイルを削除する（新しいファイルが保存され、かつ古いファイルが存在した場合）
	if newFileUploaded && oldFileName != "" && oldFileName != newFileName {
		oldFilePath := filepath.Join("uploads", oldFileName)
		log.Printf("古い画像ファイルを削除します: %s", oldFilePath)
		if err := os.Remove(oldFilePath); err != nil {
			log.Printf("古い画像ファイルの削除エラー: %v", err)
		}
	}

	// 更新成功のレスポンスを返す
	c.JSON(http.StatusOK, gin.H{"message": "商品を編集しました"})
}

// 商品を削除する関数
func AdminDeleteProductHandler(c *gin.Context) {
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

	// 商品が存在するか確認し、既存の画像ファイル名を取得する
	var imageUrlToDelete sql.NullString
	checkQuery := "SELECT image_url FROM products WHERE id = ?"
	err = db.QueryRow(checkQuery, id).Scan(&imageUrlToDelete)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("削除対象の商品が見つかりません: ID=%d", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "商品が見つかりませんでした"})
		} else {
			log.Printf("既存商品チェックエラー (ID=%d): %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		}
		return
	}

	// データベースから削除する
	deleteQuery := "DELETE FROM products WHERE id = ?"
	_, err = db.Exec(deleteQuery, id)
	if err != nil {
		log.Printf("商品削除エラー (ID=%d): %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// 画像ファイルを削除する（存在すれば）
	if imageUrlToDelete.Valid && imageUrlToDelete.String != "" {
		filePath := filepath.Join("uploads", imageUrlToDelete.String)
		log.Printf("関連する画像ファイルを削除します: %s", filePath)
		if err := os.Remove(filePath); err != nil {
			log.Printf("画像ファイルの削除エラー: %v", err)
		}
	}

	// 削除成功のレスポンスを返す
	c.JSON(http.StatusOK, gin.H{"message": "商品を削除しました"})
}