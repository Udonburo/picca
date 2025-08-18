package main

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var httpClient = &http.Client{Timeout: 3 * time.Second}

func logReq(id string, status int, upstreamMs int64) {
	log.Printf(`{"service":"go-gateway","request_id":"%s","status":%d,"upstream_ms":%d}`, id, status, upstreamMs)
}

func maxBodyBytes() int64 {
	if v := os.Getenv("MAX_BODY_BYTES"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			return n
		}
	}
	return 1 << 20 // 1 MiB default
}

func scoreHandler(c *gin.Context) {
	reqID := c.GetHeader("X-Request-Id")
	if reqID == "" {
		reqID = time.Now().UTC().Format(time.RFC3339Nano)
	}
	c.Header("X-Request-Id", reqID)

	expectedKey := os.Getenv("API_KEY")
	if expectedKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server misconfigured", "reason_code": "MISCONFIGURED_API_KEY"})
		logReq(reqID, http.StatusInternalServerError, 0)
		return
	}
	if c.GetHeader("X-API-Key") != expectedKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "reason_code": "INVALID_API_KEY"})
		logReq(reqID, http.StatusUnauthorized, 0)
		return
	}

	if ct := c.GetHeader("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "unsupported media", "reason_code": "UNSUPPORTED_MEDIA_TYPE"})
		logReq(reqID, http.StatusUnsupportedMediaType, 0)
		return
	}

	// limit body size
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes())
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body", "reason_code": "INVALID_BODY"})
		logReq(reqID, http.StatusBadRequest, 0)
		return
	}

	mlURL := strings.TrimRight(os.Getenv("API_ML_URL"), "/")
	if mlURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server misconfigured", "reason_code": "MISCONFIGURED_UPSTREAM"})
		logReq(reqID, http.StatusInternalServerError, 0)
		return
	}

	upstreamReq, err := http.NewRequest(http.MethodPost, mlURL+"/predict", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "ml upstream error", "reason_code": "UPSTREAM_FAILURE"})
		logReq(reqID, http.StatusBadGateway, 0)
		return
	}
	upstreamReq.Header.Set("Content-Type", "application/json")
	upstreamReq.Header.Set("Accept", "application/json")
	upstreamReq.Header.Set("X-Request-Id", reqID)

	start := time.Now()
	resp, err := httpClient.Do(upstreamReq)
	duration := time.Since(start).Milliseconds()
	if err != nil {
		// Distinguish timeout vs other upstream failures
		var nerr net.Error
		if errors.As(err, &nerr) && nerr.Timeout() {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "ml upstream timeout", "reason_code": "UPSTREAM_TIMEOUT"})
			logReq(reqID, http.StatusGatewayTimeout, duration)
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": "ml upstream error", "reason_code": "UPSTREAM_FAILURE"})
		logReq(reqID, http.StatusBadGateway, duration)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "ml upstream error", "reason_code": "UPSTREAM_FAILURE"})
		logReq(reqID, http.StatusBadGateway, duration)
		return
	}

	c.Header("X-Request-Id", reqID)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
	logReq(reqID, resp.StatusCode, duration)
}

func main() {
	r := gin.Default()

	r.POST("/api/v1/score", scoreHandler)
	r.GET("/healthz", func(c *gin.Context) {
		c.String(200, "ok")
	})
	r.GET("/v1/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"msg": "pong"})
	})

	// Cloud Run sets $PORT; fall back to 8080 for local runs
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 listening on :%s …", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
