# ---------- Build stage ----------
FROM node:20-alpine AS builder
WORKDIR /app

# 1. 依存解決
COPY package*.json ./
RUN npm ci --prefer-offline --no-audit

# 2. ソースコピー＋ビルド
COPY . .
RUN npm run build       # `next build`

# ---------- Run stage ----------
FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production PORT=8080

COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./static
COPY --from=builder /app/public ./public

EXPOSE 8080

CMD ["node", "server.js"]

