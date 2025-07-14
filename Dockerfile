# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
# 1. 依存解決だけ先に
COPY app/package*.json ./app/
WORKDIR /app/app
RUN npm ci
# 2. ソースコードをコピーしてビルド
COPY app/ .
RUN npm run build
# 開発用依存を削除
RUN npm prune --omit=dev

# Run stage
FROM gcr.io/distroless/nodejs20-debian12
WORKDIR /app
# ビルド成果物をコピー
COPY --from=builder /app/app ./
# Cloud Run が渡す PORT を利用
ENV NODE_ENV=production PORT=8080
EXPOSE 8080
# Express ベースのサーバーで起動
CMD ["node","server.js"]
