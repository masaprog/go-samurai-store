package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

// データベース接続を初期化し、グローバル変数dbに格納する関数
func InitDB() {
	// 本番環境ではCloud Runで設定した環境変数を使ってDSNを構築する
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSocketPath := os.Getenv("DB_SOCKET_PATH")

	var dsn string
	if dbSocketPath != "" {
		dsn = fmt.Sprintf("%s:%s@unix(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
			dbUser, dbPass, dbSocketPath, dbName)
	} else {
		// 開発環境では.envファイルから環境変数DB_DSNを直接読み込む
		dsn = os.Getenv("DB_DSN")
		if dsn == "" {
			log.Fatal("エラー: 環境変数DB_DSN（開発環境）またはDB_SOCKET_PATH（本番環境）が設定されていません")
		}
	}

	// データベースへの接続準備を行う
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("データベースへの接続準備に失敗しました: %v", err)
	}

	// データベースへの接続を試みる
	err = conn.Ping()
	if err != nil {
		log.Fatalf("データベースへの接続に失敗しました: %v", err)
	}

	log.Println("データベースへの接続に成功しました！")
	// グローバル変数dbに接続を保存する
	db = conn
}

// 初期化されたデータベース接続を返す関数
func GetDB() *sql.DB {
	if db == nil {
		log.Fatal("データベースが初期化されていません")
	}
	return db
}