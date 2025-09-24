package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2/google"
)

//go:embed web/*
var webFS embed.FS

var (
	httpClient = &http.Client{Timeout: 3 * time.Second}

	newVertexClient = func(ctx context.Context) (*http.Client, error) {
		return google.DefaultClient(ctx, "https://www.googleapis.com/auth/cloud-platform")
	}
)

type explainRequest struct {
	Score       float64 `json:"score"`
	Symmetry    float64 `json:"symmetry"`
	Power       float64 `json:"power"`
	Consistency float64 `json:"consistency"`
}

type vertexResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func mountDemo(r *gin.Engine) {
	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		panic("failed to load embedded web assets: " + err.Error())
	}

	fsys := http.FS(sub)
	fileServer := http.StripPrefix("/demo/", http.FileServer(fsys))

	serveDemo := func(c *gin.Context) {
		c.FileFromFS("demo.html", fsys)
	}

	r.GET("/demo", serveDemo)
	r.GET("/demo/*filepath", func(c *gin.Context) {
		path := strings.TrimPrefix(c.Param("filepath"), "/")
		if path == "" || path == "index.html" {
			serveDemo(c)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}

func apiKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := requestID(c)
		if !validateAPIKey(c, reqID) {
			c.Abort()
			return
		}
		c.Next()
	}
}

func mountAPI(r *gin.Engine) {
	apiV1 := r.Group("/api/v1")
	apiV1.POST("/score", scoreHandler)
	apiV1.OPTIONS("/explain", explainOptionsHandler)

	api := r.Group("/api/v1", apiKeyMiddleware())
	api.POST("/explain", explainHandler)
	log.Println("mounted /api/v1/explain")

	for _, alias := range []string{"/explain", "/api/explain", "/v1/explain"} {
		r.POST(alias, apiKeyMiddleware(), explainHandler)
	}
}

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

func requestID(c *gin.Context) string {
	reqID := c.GetHeader("X-Request-Id")
	if reqID == "" {
		reqID = time.Now().UTC().Format(time.RFC3339Nano)
	}
	c.Header("X-Request-Id", reqID)
	return reqID
}

func validateAPIKey(c *gin.Context, reqID string) bool {
	expectedKey := os.Getenv("API_KEY")
	if expectedKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server misconfigured", "reason_code": "MISCONFIGURED_API_KEY"})
		logReq(reqID, http.StatusInternalServerError, 0)
		return false
	}
	if c.GetHeader("X-API-Key") != expectedKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "reason_code": "INVALID_API_KEY"})
		logReq(reqID, http.StatusUnauthorized, 0)
		return false
	}
	return true
}

func ensureJSONContentType(c *gin.Context, reqID string) bool {
	if ct := c.GetHeader("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "unsupported media", "reason_code": "UNSUPPORTED_MEDIA_TYPE"})
		logReq(reqID, http.StatusUnsupportedMediaType, 0)
		return false
	}
	return true
}

func resolveProjectID(ctx context.Context) (string, error) {
	if v := strings.TrimSpace(os.Getenv("PROJECT_ID")); v != "" {
		return v, nil
	}

	client := metadata.NewClient(&http.Client{Timeout: 2 * time.Second})
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	projectID, err := client.ProjectIDWithContext(ctx)
	if err != nil {
		return "", err
	}
	if projectID == "" {
		return "", errors.New("empty project id from metadata")
	}
	return projectID, nil
}

func extractVertexSummary(body []byte) (string, error) {
	var resp vertexResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			if strings.TrimSpace(part.Text) != "" {
				return part.Text, nil
			}
		}
	}
	return "", errors.New("no summary in response")
}

func scoreHandler(c *gin.Context) {
	reqID := requestID(c)

	if !validateAPIKey(c, reqID) {
		return
	}
	if !ensureJSONContentType(c, reqID) {
		return
	}

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

	upstreamReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, mlURL+"/predict", bytes.NewReader(body))
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

func explainOptionsHandler(c *gin.Context) {
	c.Header("Allow", "OPTIONS, POST")
	c.Header("Access-Control-Allow-Methods", "OPTIONS, POST")
	c.Header("Access-Control-Allow-Headers", "Content-Type,X-API-Key")
	c.Status(http.StatusOK)
}

func explainHandler(c *gin.Context) {
	reqID := requestID(c)

	if !validateAPIKey(c, reqID) {
		return
	}
	if !ensureJSONContentType(c, reqID) {
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes())
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body", "reason_code": "INVALID_BODY"})
		logReq(reqID, http.StatusBadRequest, 0)
		return
	}

	var payload explainRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body", "reason_code": "INVALID_BODY"})
		logReq(reqID, http.StatusBadRequest, 0)
		return
	}

	region := strings.TrimSpace(os.Getenv("VERTEX_REGION"))
	if region == "" {
		region = "us-central1"
	}

	model := strings.TrimSpace(os.Getenv("VERTEX_MODEL"))
	if model == "" {
		model = "gemini-2.5-flash-lite"
	}

	projectID, err := resolveProjectID(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server misconfigured", "reason_code": "MISCONFIGURED_PROJECT_ID"})
		logReq(reqID, http.StatusInternalServerError, 0)
		return
	}

	prompt := fmt.Sprintf("Summarize these metrics: score=%g, symmetry=%g, power=%g, consistency=%g. 1-2 sentences.", payload.Score, payload.Symmetry, payload.Power, payload.Consistency)

	vertexPayload := map[string]any{
		"contents": []map[string]any{
			{
				"role": "user",
				"parts": []map[string]any{
					{
						"text": prompt,
					},
				},
			},
		},
	}
	reqBytes, err := json.Marshal(vertexPayload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error", "reason_code": "VERTEX_REQUEST_MARSHAL_ERROR"})
		logReq(reqID, http.StatusInternalServerError, 0)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	client, err := newVertexClient(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "vertex auth error", "reason_code": "VERTEX_AUTH_FAILURE"})
		logReq(reqID, http.StatusInternalServerError, 0)
		return
	}

	escapedProject := url.PathEscape(projectID)
	escapedRegion := url.PathEscape(region)
	escapedModel := url.PathEscape(model)

	host := fmt.Sprintf("%s-aiplatform.googleapis.com", region)
	if region == "global" {
		host = "aiplatform.googleapis.com"
	}
	vertexURL := fmt.Sprintf("https://%s/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent", host, escapedProject, escapedRegion, escapedModel)

	vertexReq, err := http.NewRequestWithContext(ctx, http.MethodPost, vertexURL, bytes.NewReader(reqBytes))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "vertex upstream error", "reason_code": "VERTEX_REQUEST_BUILD_FAILURE"})
		logReq(reqID, http.StatusBadGateway, 0)
		return
	}
	vertexReq.Header.Set("Content-Type", "application/json")
	vertexReq.Header.Set("Accept", "application/json")
	vertexReq.Header.Set("X-Request-Id", reqID)

	start := time.Now()
	vertexResp, err := client.Do(vertexReq)
	duration := time.Since(start).Milliseconds()
	if err != nil {
		var nerr net.Error
		if errors.As(err, &nerr) && nerr.Timeout() {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "vertex upstream timeout", "reason_code": "VERTEX_UPSTREAM_TIMEOUT"})
			logReq(reqID, http.StatusGatewayTimeout, duration)
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": "vertex upstream error", "reason_code": "VERTEX_UPSTREAM_FAILURE"})
		logReq(reqID, http.StatusBadGateway, duration)
		return
	}
	defer vertexResp.Body.Close()

	respBody, err := io.ReadAll(vertexResp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "vertex upstream error", "reason_code": "VERTEX_UPSTREAM_FAILURE"})
		logReq(reqID, http.StatusBadGateway, duration)
		return
	}

	if vertexResp.StatusCode < 200 || vertexResp.StatusCode >= 300 {
		c.Data(vertexResp.StatusCode, vertexResp.Header.Get("Content-Type"), respBody)
		logReq(reqID, vertexResp.StatusCode, duration)
		return
	}

	summary, err := extractVertexSummary(respBody)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "vertex upstream error", "reason_code": "VERTEX_INVALID_RESPONSE"})
		logReq(reqID, http.StatusBadGateway, duration)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": summary,
		"model":   model,
		"region":  region,
	})
	logReq(reqID, http.StatusOK, duration)
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	mountDemo(r)
	mountAPI(r)

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
