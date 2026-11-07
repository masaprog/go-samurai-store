package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

// データベース接続を初期化し、グローバル変数dbに格納する関数
func InitDB() {
	// 環境変数からDSN（データソース名）を取得する
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("エラー: 環境変数DB_DSNが設定されていません")
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