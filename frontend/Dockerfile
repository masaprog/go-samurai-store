# --- 1. ビルドステージ ---
# Node.jsの公式イメージをベースにする
FROM node:22-alpine AS builder
LABEL stage=builder

# コンテナ内の作業ディレクトリ
WORKDIR /app

# 依存関係をインストールする
COPY package.json package-lock.json ./
RUN npm ci

# Next.jsの設定ファイルとアプリケーションコードをコピーする
COPY next.config.* ./
COPY public ./public
COPY src ./src
# 他にもビルドに必要なファイルがあればコピーする
COPY tsconfig.json ./
COPY postcss.config.mjs ./

# ビルド時に必要な環境変数を設定する
ARG API_BASE_URL
ENV API_BASE_URL=${API_BASE_URL}

# Next.jsの本番ビルドを実行する
# これにより.next/standaloneディレクトリと.next/staticディレクトリが生成される
RUN npm run build


# --- 2. 実行ステージ ---
# Node.jsの公式イメージをベースにする
FROM node:22-alpine AS runner

# コンテナ内の作業ディレクトリ
WORKDIR /app

# 環境変数を設定する
ENV NODE_ENV=production
ENV PORT=3000
# （API_BASE_URLなど、実行時に必要な環境変数はCloud Run側で設定する）

# ビルドステージから必要なファイルのみをコピーする
# standalone出力に必要なファイル群
COPY --from=builder /app/public ./public
COPY --from=builder --chown=node:node /app/.next/standalone ./
COPY --from=builder --chown=node:node /app/.next/static ./.next/static

# Node.jsユーザーに変更する（セキュリティ向上のため）
USER node

# アプリケーションがリッスンするポート
EXPOSE 3000

# コンテナ起動時に実行するコマンド
# .next/standaloneディレクトリ内のserver.jsファイルを実行してNext.jsサーバーを起動する
CMD ["node", "server.js"]