package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func main() {
	// 64バイトのランダムデータを生成する
	b := make([]byte, 64)

	// エラーが発生したらプログラムを停止する
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	// バイト列を16進数文字列に変換して出力する
	fmt.Println(hex.EncodeToString(b))
}