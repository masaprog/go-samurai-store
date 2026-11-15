# --- 1. ビルドステージ ---
# Goの公式イメージをベースにする
FROM golang:1.25-alpine AS builder
LABEL stage=builder

# ビルドに必要なツール（git, ca-certificates）をインストール
# ca-certificatesはGoアプリがHTTPS通信（例: Stripe API）を行うために必要
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# コンテナ内の作業ディレクトリ
WORKDIR /app

# Goモジュールの依存関係を先にコピーしてダウンロードする
# （ソースコード変更時に毎回ダウンロードし直さないようにするため）
COPY go.mod go.sum ./
RUN go mod download

# アプリケーションの全ソースコードをコピーする
COPY . .

# アプリケーションをビルドする
# CGO_ENABLED=0: C言語のライブラリを使わない静的バイナリを作成する
# GOOS=linux: Linux 環境向けのバイナリを作成する
# -ldflags="-w -s": デバッグ情報などを削除し、バイナリサイズを削減する
# -o /app/server: 出力ファイル名を指定する
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server/main.go


# --- 2. 実行ステージ ---
# 軽量なAlpine Linuxイメージをベースにする
FROM alpine:latest

# ビルドステージでインストールしたca-certificatesとtzdataをコピーする
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# タイムゾーンを日本時間に設定する
ENV TZ=Asia/Tokyo

# コンテナ内の作業ディレクトリ
WORKDIR /app

# ビルドステージからコンパイル済みの実行バイナリのみをコピーする
COPY --from=builder /app/server /app/server
# ビルドステージからマイグレーションファイルもコピーする
COPY --from=builder /app/db /app/db
# ビルドステージからアップロードファイル保存用ディレクトリもコピーする
COPY --from=builder /app/uploads /app/uploads

# .envファイルをイメージに含めたい場合はここでコピーするが、Cloud Runでは環境変数で渡すため通常は含めない
# COPY .env .env

# アプリケーションがリッスンするポート（Cloud Runは通常8080を期待）
EXPOSE 8080

# コンテナ起動時に実行するコマンド
# /app/serverバイナリを実行する
CMD ["/app/server"]