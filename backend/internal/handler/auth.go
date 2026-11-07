package handler

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/masaprog/go-samurai-store/backend/internal/database"
)

// --- 1. 型定義（struct） ---

// ログインリクエストの構造体
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// データベースから取得するユーザー情報の構造体
type userForAuth struct {
	ID       int
	Name     string
	Email    string
	Password string
	IsAdmin  bool
}

// JWTトークンに含めるクレーム（Payload）の構造体
type JWTCustomClaims struct {
	UserID               int    `json:"userId"`
	Name                 string `json:"name"`
	Email                string `json:"email"`
	IsAdmin              bool   `json:"isAdmin"`
	jwt.RegisteredClaims        // 標準クレーム（iss, exp, iatなど）を埋め込む
}

// JWTトークンの署名に使う秘密鍵を環境変数から取得する
var JWTSecret = []byte(os.Getenv("JWT_SECRET"))

// ブラウザのクッキーに保存する際のクッキー名
const AuthTokenCookieName = "authToken"

// JWTトークンを検証し、クレームを返す関数
func VerifyToken(tokenString string) (*JWTCustomClaims, error) {
	// 関数ParseWithClaimsを使ってJWTトークンをパース・署名検証し、
	// 結果のクレームをJWTCustomClaims構造体にマッピングする
	token, err := jwt.ParseWithClaims(tokenString, &JWTCustomClaims{}, func(token *jwt.Token) (any, error) {
		// 署名方法がHS256であることを確認する
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("予期しない署名方法です")
		}
		// 署名検証用の秘密鍵を返す
		return JWTSecret, nil
	})

	// パースまたは署名検証に失敗した場合はエラーを返す
	if err != nil {
		log.Printf("トークン検証エラー: %v", err)
		return nil, err
	}

	// クレームを*JWTCustomClaims型として型アサーションでき、かつJWTトークンが有効か確認する
	if claims, ok := token.Claims.(*JWTCustomClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("無効なトークンです")
	}
}

// --- 2. ハンドラ定義 ---

// ログイン処理を行う関数
func LoginHandler(c *gin.Context) {
	var req LoginRequest
	// HTTPリクエストボディをLoginRequest構造体にバインドする
	// 必須項目や形式が異なる場合はエラーを返す
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ログインリクエストのバインドエラー: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "メールアドレスまたはパスワードが正しくありません"})
		return
	}

	// データベースからユーザーを検索する（enabled = trueも確認）
	db := database.GetDB()
	var user userForAuth
	query := "SELECT id, name, email, password, is_admin FROM users WHERE email = ? AND enabled = TRUE"
	err := db.QueryRow(query, req.Email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.IsAdmin,
	)

	// ユーザーが見つからない場合のエラーハンドリングを行う
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("ログイン失敗 （ユーザーが見つからないか無効）: email=%s", req.Email)
		} else {
			log.Printf("ユーザー検索エラー: %v", err)
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "メールアドレスまたはパスワードが違います"})
		return
	}

	// パスワードを検証する
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		log.Printf("ログイン失敗 （パスワード不一致）: email=%s", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "メールアドレスまたはパスワードが違います"})
		return
	}

	// JWTトークンを生成する
	expirationTime := time.Now().Add(1 * time.Hour) // 有効期限: 1時間
	claims := &JWTCustomClaims{
		UserID:  user.ID,
		Name:    user.Name,
		Email:   user.Email,
		IsAdmin: user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// JWTトークンに署名する
	tokenString, err := token.SignedString(JWTSecret)
	if err != nil {
		log.Printf("JWT署名エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// JWTトークンをクッキーに保存する
	c.SetCookie(
		AuthTokenCookieName, // クッキー名
		tokenString,         // JWTトークンの文字列
		3600,                // 有効期限（秒）
		"/",                 // パス
		"",                  // ドメイン
		false,               // 暗号化通信（HTTPS）のみを許可するかどうか（本番環境ではtrue）
		true,                // JavaScriptからのアクセスを禁止するHttpOnly属性を付与するかどうか
	)

	// ログイン成功のレスポンスを返す
	c.JSON(http.StatusOK, gin.H{
		"message": "ログイン成功",
		"isAdmin": user.IsAdmin,
	})
}

// ログアウト処理を行う関数
func LogoutHandler(c *gin.Context) {
	// クッキーの有効期限を過去に設定することで削除する
	c.SetCookie(
		AuthTokenCookieName,
		"", // クッキー名は空文字列にする
		-1, // 有効期限を過去に設定する
		"/",
		"",
		false,
		true,
	)

	// フロントエンドのURL（トップページ）にリダイレクトさせる
	// クエリパラメータ（?logged-out=1）を付与し、ログアウトしたことを通知する
	c.Redirect(http.StatusFound, "/?logged-out=1")
}