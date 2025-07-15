# ---------- Build stage ----------
FROM node:20-alpine AS builder
WORKDIR /app

# 1. 依存解決
COPY app/package.json app/package-lock.json ./
RUN npm ci

# 2. ソースコピー＋ビルド
COPY app/ ./
RUN npm run build

# ---------- Run stage ----------
FROM node:20-alpine AS runner   
# 実行ステージ（npm が入っている）

WORKDIR /app

# ビルド成果物＋node_modules をコピー
COPY --from=builder /app ./

# Cloud Run 用
ENV NODE_ENV=production
ENV PORT=8080
EXPOSE 8080

# Next.js を npm 経由で起動
CMD ["sh", "-c", "npm start"]
