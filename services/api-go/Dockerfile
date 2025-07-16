# ---------- Build stage ----------
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o server .

# ---------- Run stage ----------
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/server .

ENV PORT=8080
EXPOSE 8080
CMD ["./server"]
