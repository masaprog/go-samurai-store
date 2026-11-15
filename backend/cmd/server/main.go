package main

import (
	"database/sql"
	"fmt"
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

	// CORSを設定する
	frontendOrigin := os.Getenv("FRONTEND_BASE_URL")
	if frontendOrigin == "" {
		frontendOrigin = "http://localhost:3000"
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{frontendOrigin},

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
			authorized.GET("/favorites", handler.ListFavoritesHandler)
			authorized.GET("/favorites", handler.GetFavoriteStatusHandler)


			authorized.POST("/products/:id/reviews", handler.CreateReviewHandler)
			authorized.POST("/favorites", handler.AddFavoriteHandler)
			
			authorized.DELETE("/favorites", handler.RemoveFavoriteHandler)

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

	// Ginサーバーを起動する
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Ginサーバーをポート%sで起動します", port)
	router.Run(":" + port)
}

// マイグレーションを実行する関数
func runMigration() {
	// 本番環境ではCloud Runで設定した環境変数を使ってDSNを構築する
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSocketPath := os.Getenv("DB_SOCKET_PATH")

	var dsn string
	if dbSocketPath != "" {
		dsn = fmt.Sprintf("%s:%s@unix(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true",
			dbUser, dbPass, dbSocketPath, dbName)
	} else {
		// 開発環境では.envファイルから環境変数DB_DSNを直接読み込む
		dsn = os.Getenv("DB_DSN")
		if dsn == "" {
			log.Fatal("エラー: 環境変数DB_DSN（開発環境）かDB_SOCKET_PATH（本番環境）のいずれかを設定してください")
		}
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