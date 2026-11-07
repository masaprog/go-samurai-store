package handler

import (
	"database/sql"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/masaprog/go-samurai-store/backend/internal/database"
)

// --- 1. 型定義（struct） ---

// 会員登録リクエストの構造体
type UserRegisterRequest struct {
	Name     string `json:"name" binding:"required"`        // binding:"required"で必須項目にする
	Email    string `json:"email" binding:"required,email"` // email形式もチェックする
	Password string `json:"password" binding:"required"`
}

// 会員情報レスポンスの構造体
type UserMeResponse struct {
	UserID  int    `json:"userId"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"isAdmin"`
}

// 会員編集リクエストの構造体
type UserUpdateRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// パスワード変更リクエストの構造体
type PasswordUpdateRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// --- 2. ハンドラ定義 ---

// 会員登録を行う関数
func RegisterUserHandler(c *gin.Context) {
	var req UserRegisterRequest
	// HTTPリクエストボディをUserRegisterRequest構造体にバインドする
	// 必須項目や形式が異なる場合はエラーを返す
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("HTTPリクエストボディのバインドエラー: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "入力内容が正しくありません"})
		return
	}

	// メールアドレスの形式をチェックする
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9.]+@[a-zA-Z0-9.]+$`)
	if !emailPattern.MatchString(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "正しいメールアドレスの形式で入力してください"})
		return
	}

	// パスワードの文字数をチェックする
	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "パスワードは8文字以上で入力してください"})
		return
	}

	// メールアドレスの重複をチェックする
	db := database.GetDB()
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", req.Email).Scan(&count)
	// QueryRow()メソッド自体のエラーもチェックする（DB接続エラーなど）
	if err != nil {
		log.Printf("メールアドレス重複チェックエラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "このメールアドレスは既に登録されています"})
		return
	}

	// パスワードをハッシュ化する
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("パスワードハッシュ化エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// ユーザーをデータベースに登録する
	insertQuery := `
		INSERT INTO users (name, email, password, is_admin, enabled)
		VALUES (?, ?, ?, false, true)
	`
	_, err = db.Exec(insertQuery, req.Name, req.Email, string(hashedPassword))
	if err != nil {
		log.Printf("ユーザー登録エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// 登録成功のレスポンスを返す
	c.JSON(http.StatusOK, gin.H{"message": "会員登録が完了しました"})
}

// 自分の会員情報を取得する関数
func GetUserMeHandler(c *gin.Context) {
	// ミドルウェア関数AuthMiddlewareによってコンテキストに保存されたクレームを取得する
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

	// 最終的なレスポンスの形にまとめる
	response := UserMeResponse{
		UserID:  claims.UserID,
		Name:    claims.Name,
		Email:   claims.Email,
		IsAdmin: claims.IsAdmin,
	}

	// JSONとしてレスポンスを返す
	c.JSON(http.StatusOK, response)
}

// 会員情報を編集する関数
func UpdateUserHandler(c *gin.Context) {
	// HTTPリクエストのコンテキストからユーザー情報を取得する
	userClaims, exists := c.Get("user")
	if !exists {
		log.Println("UpdateUserHandler: ユーザー情報がコンテキストにありません")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}
	claims, ok := userClaims.(*JWTCustomClaims) // auth.goファイルで定義した型
	if !ok {
		log.Println("UpdateUserHandler: ユーザー情報の型アサーションに失敗しました")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}
	userID := claims.UserID // 更新対象のユーザーID

	// HTTPリクエストボディをUserRegisterRequest構造体にバインドする
	// 必須項目や形式が異なる場合はエラーを返す
	var req UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("会員編集リクエストのバインドエラー: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "入力内容が正しくありません"})
		return
	}

	// メールアドレスの形式をチェックする
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9.]+@[a-zA-Z0-9.]+$`)
	if !emailPattern.MatchString(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "正しいメールアドレスの形式で入力してください"})
		return
	}

	// メールアドレスの重複をチェックする（自分自身のメールアドレスは除外する）
	db := database.GetDB()
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? AND id != ?", req.Email, userID).Scan(&count)
	if err != nil {
		log.Printf("メールアドレス重複チェックエラー (更新時): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "このメールアドレスは既に使用されています"})
		return
	}

	// データベースを更新する
	updateQuery := "UPDATE users SET name = ?, email = ? WHERE id = ?"
	_, err = db.Exec(updateQuery, req.Name, req.Email, userID)
	if err != nil {
		log.Printf("会員情報更新エラー (UserID=%d): %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// JWTトークンを更新後の情報で再生成する
	expirationTime := time.Now().Add(1 * time.Hour) // 有効期限はログイン時と同じ
	newClaims := &JWTCustomClaims{                  // auth.goファイルで定義した型
		UserID:  userID,
		Name:    req.Name,       // 更新後の氏名
		Email:   req.Email,      // 更新後のメールアドレス
		IsAdmin: claims.IsAdmin, // 管理者権限は元のJWTトークンから引き継ぐ
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()), // 発行日時を更新する
		},
	}
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)

	// JWTトークンに署名する
	newTokenString, err := newToken.SignedString(JWTSecret) // auth.goファイルで定義した変数
	if err != nil {
		// データベースの更新は成功しているので、ここではエラーログだけ残して処理を続ける
		log.Printf("JWT再署名エラー (更新時): %v", err)
	}

	// 署名に成功していれば、新しいJWTトークンをクッキーに保存する
	if newTokenString != "" {
		c.SetCookie(
			AuthTokenCookieName, // クッキー名（auth.goファイルで定義した定数）
			newTokenString,      // 新しいJWTトークンの文字列
			3600,                // 有効期限 (秒)
			"/",                 // パス
			"",                  // ドメイン
			false,               // 暗号化通信（HTTPS）のみを許可するかどうか（本番環境ではtrue）
			true,                // JavaScriptからのアクセスを禁止するHttpOnly属性を付与するかどうか
		)
	}

	// 更新成功のレスポンスを返す
	c.JSON(http.StatusOK, gin.H{"message": "会員情報を編集しました"})
}

// パスワードを変更する関数
func UpdatePasswordHandler(c *gin.Context) {
	// HTTPリクエストのコンテキストからユーザー情報を取得する
	userClaims, exists := c.Get("user")
	if !exists {
		log.Println("UpdatePasswordHandler: ユーザー情報がコンテキストにありません")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}
	claims, ok := userClaims.(*JWTCustomClaims) // auth.goファイルで定義した型
	if !ok {
		log.Println("UpdatePasswordHandler: ユーザー情報の型アサーションに失敗しました")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}
	userID := claims.UserID // 更新対象のユーザーID

	// HTTPリクエストボディをPasswordUpdateRequest構造体にバインドする
	// 必須項目や形式が異なる場合はエラーを返す
	var req PasswordUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("パスワード変更リクエストのバインドエラー: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "入力内容が正しくありません"})
		return
	}

	// 新しいパスワードの文字数をチェックする
	if len(req.NewPassword) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "パスワードは8文字以上で入力してください"})
		return
	}

	// 現在のパスワードをデータベースから取得する
	db := database.GetDB()
	var currentPasswordHash string
	err := db.QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&currentPasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("パスワード変更エラー: ユーザーが見つかりません（UserID=%d）", userID)
			c.JSON(http.StatusNotFound, gin.H{"error": "ユーザー情報が見つかりません"})
		} else {
			log.Printf("現在のパスワード取得エラー（UserID=%d）: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		}
		return
	}

	// 現在のパスワードが正しいか検証する
	err = bcrypt.CompareHashAndPassword([]byte(currentPasswordHash), []byte(req.OldPassword))
	if err != nil {
		log.Printf("パスワード変更失敗（現在のパスワード不一致）: UserID=%d", userID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "現在のパスワードが正しくありません"})
		return
	}

	// 新しいパスワードをハッシュ化する
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("新しいパスワードのハッシュ化エラー（UserID=%d）: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// データベースのパスワードを更新する
	updateQuery := "UPDATE users SET password = ? WHERE id = ?"
	_, err = db.Exec(updateQuery, string(newPasswordHash), userID)
	if err != nil {
		log.Printf("パスワード更新エラー（UserID=%d）: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// 更新成功のレスポンスを返す
	c.JSON(http.StatusOK, gin.H{"message": "パスワードを変更しました"})
}