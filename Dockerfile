# ---------- Build stage ----------
FROM node:20-alpine AS builder
WORKDIR /app

# 依存解決だけ先に
COPY app/package*.json ./
RUN npm ci

# ソースとビルド
COPY app ./app
WORKDIR /app/app
RUN npm run build

# ---------- Run stage ----------
FROM gcr.io/distroless/nodejs20-debian12
WORKDIR /app
# builder からビルド成果物一式をコピー
COPY --from=builder /app/app ./

# Cloud Run が渡してくる PORT を拾う
ENV PORT=8080 NODE_ENV=production
EXPOSE 8080

# Next.js のランタイムサーバーで起動
CMD ["npm","start"]
