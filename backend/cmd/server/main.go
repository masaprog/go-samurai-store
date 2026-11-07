package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/masaprog/go-samurai-store/backend/internal/database"
	"github.com/masaprog/go-samurai-store/backend/internal/handler"
	"github.com/masaprog/go-samurai-store/backend/internal/middleware"
)

func main() {
	// 環境変数JWT_SECRETが設定されているか確認する
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("エラー: 環境変数JWT_SECRETが設定されていません")
	}

	//マイグレーションを実行する
	runMigration()

	//データベース接続を初期化する
	database.InitDB()

	// Ginのデフォルトルーターを作成する
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000"},

		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},

		AllowHeaders: []string{"Content-Type"},

		AllowCredentials: true,

		MaxAge: 12 * time.Hour,
	}))

	router.StaticFS("/uploads", http.Dir("uploads"))

	// /api以下の各ルートを定義する
	api := router.Group("/api")
	{
		api.GET("/sample", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "ok",
				"message": "Hello, World!",
			})
		})

		api.GET("/products", handler.GetProductsHandler)
		api.GET("/products/:id", handler.GetProductByIDHandler)
		api.GET("/home", handler.GetHomePageProductsHandler)
		api.GET("/products/:id/reviews", handler.GetReviewsHandler)


		api.POST("/users", handler.RegisterUserHandler)
		api.POST("/inquiries", handler.CreateInquiryHandler)
		api.POST("/orders/webhook", handler.StripeWebhookHandler)


		auth := api.Group("/auth")
		{
			auth.POST("/login", handler.LoginHandler)
			auth.POST("/logout", handler.LogoutHandler)
		}
	}

		// 認証が必要なルートグループ
		// これらのルートは、先にミドルウェア関数AuthMiddlewareが実行される
		authorized := api.Group("/")
		authorized.Use(middleware.AuthMiddleware())
		{
			authorized.GET("/users/me", handler.GetUserMeHandler)
			authorized.PUT("/users", handler.UpdateUserHandler)
			authorized.PUT("/users/password", handler.UpdatePasswordHandler)


			authorized.POST("/orders/checkout", handler.CreateCheckoutSessionHandler)
			authorized.GET("/orders", handler.GetOrdersHandler)
			authorized.GET("/inquiries", handler.ListInquiriesHandler)


			authorized.POST("/products/:id/reviews", handler.CreateReviewHandler)


		}

		// 管理者権限が必要なルートグループ
		// これらのルートは、先にミドルウェア関数AuthMiddlewareとAdminAuthMiddlewareが実行される
		admin := api.Group("/")
		admin.Use(middleware.AuthMiddleware())
		admin.Use(middleware.AdminAuthMiddleware())
		{
			// ここに管理者権限が必要なルートを追加する
			admin.POST("/products", handler.AdminCreateProductHandler)
			admin.PUT("/products/:id", handler.AdminUpdateProductHandler)
			admin.DELETE("/products/:id", handler.AdminDeleteProductHandler)
		}

	// Ginサーバーをポート8080で起動する
	log.Println("Ginサーバーをポート8080で起動します")
	router.Run(":8080")
}

// マイグレーションを実行する関数
func runMigration() {
	// 環境変数からDSN（データソース名）を取得する
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("エラー: 環境変数DB_DSNが設定されていません")
	}

	log.Println("マイグレーションを開始します...")

	// データベースへの接続準備を行う
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("データベースへの接続準備に失敗しました: %v", err)
	}

	// マイグレーションが終わったらデータベース接続を閉じる
	defer db.Close()

	// データベースへの接続を試みる
	err = db.Ping()
	if err != nil {
		log.Fatalf("データベースへの接続に失敗しました: %v", err)
	}

	// マイグレーション用のデータベースドライバーを作成する
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatalf("マイグレーション用のデータベースドライバーの作成に失敗しました: %v", err)
	}

	// マイグレーションのソース（SQLファイル）の場所を指定する
	// Dockerfile.devファイルのWORKDIR（/app）からの相対パス
	sourceURL := "file://db/migrations"

	// マイグレーション用のインスタンスを作成する
	m, err := migrate.NewWithDatabaseInstance(sourceURL, "mysql", driver)
	if err != nil {
		log.Fatalf("マイグレーションのセットアップに失敗しました: %v", err)
	}

	// マイグレーションを実行する
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("マイグレーションに失敗しました: %v", err)
	}

	log.Println("マイグレーションに成功しました！")
}