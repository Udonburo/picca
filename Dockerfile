FROM node:20-alpine

# 1. 作業ディレクトリ
WORKDIR /app

# 2. package*.json を先にコピーして依存解決
COPY package.json package-lock.json ./
RUN npm ci --omit=dev

# 3. ソースコードをコピー
COPY . .

# 4. （ここは好きに変えてOK）サンプルとして math.js を実行
CMD ["node", "src/math.js"]
