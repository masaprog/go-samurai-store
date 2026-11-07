package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/masaprog/go-samurai-store/backend/internal/handler"
)

// HTTPリクエストのクッキーに保存されたJWTトークンを検証するミドルウェア関数
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// HTTPリクエストのクッキーに保存されたJWTトークンを取得する
		tokenString, err := c.Cookie(handler.AuthTokenCookieName)
		// クッキーが存在しない場合、401 Unauthorizedエラーを返し、以降の処理を中断する
		if err != nil {
			log.Println("認証ミドルウェア: クッキーがありません")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
			return
		}

		// 関数VerifyTokenを使ってJWTトークンを検証する
		claims, err := handler.VerifyToken(tokenString)
		// JWTトークンが無効の場合、401 Unauthorizedエラーを返し、以降の処理を中断する
		if err != nil {
			log.Printf("認証ミドルウェア: トークンが無効です: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "認証トークンが無効です"})
			return
		}

		// ユーザー情報をHTTPリクエストのコンテキストに保存する
		// （後続のミドルウェア関数またはハンドラ関数がc.Get("user")で取得できるようにする）
		c.Set("user", claims)

		// 次のミドルウェア関数またはハンドラ関数を実行する
		c.Next()
	}
}

// 管理者権限をチェックするミドルウェア関数
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// HTTPリクエストのコンテキストからユーザー情報を取得する
		// （ミドルウェア関数AuthMiddlewareが実行されていることが前提）
		userClaims, exists := c.Get("user")
		// コンテキストにユーザー情報が存在しない場合、401 Unauthorizedエラーを返す
		if !exists {
			log.Println("管理者権限ミドルウェア: ユーザー情報がコンテキストに存在しません")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
			return
		}

		// 型アサーションでクレームを取得する
		claims, ok := userClaims.(*handler.JWTCustomClaims)
		// クレームの型が異なる、またはIsAdminフィールドがfalseの場合、403 Forbiddenエラーを返す
		if !ok || !claims.IsAdmin {
			log.Printf("管理者権限ミドルウェア: 権限がありません (IsAdmin=%v)", claims.IsAdmin)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "アクセス権限がありません"})
			return
		}

		// 次のミドルウェア関数またはハンドラ関数を実行する
		c.Next()
	}
}