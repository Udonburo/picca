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
	"os/signal"
	"strconv"
	"strings"
	"time"

	"syscall"

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

	runID = deriveRunID()
)

func deriveRunID() string {
	script := strings.TrimSpace(os.Getenv("PICCA_SCRIPT_SHA"))
	if script == "" {
		script = "dev"
	}
	ckpt := strings.TrimSpace(os.Getenv("MODEL_CKPT_SHA"))
	if ckpt == "" {
		ckpt = "na"
	}
	return fmt.Sprintf("%s·%s", script, ckpt)
}

func apiKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID(c)
		if !validateAPIKey(c) {
			c.Abort()
			return
		}
		c.Next()
	}
}

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
	if sub, err := fs.Sub(webFS, "web"); err == nil {
		fsys := http.FS(sub)
		fileServer := http.StripPrefix("/demo/", http.FileServer(fsys))
		serveDemo := func(c *gin.Context) { c.FileFromFS("demo.html", fsys) }
		r.GET("/demo", serveDemo)
		r.GET("/demo/*filepath", func(c *gin.Context) {
			p := strings.TrimPrefix(c.Param("filepath"), "/")
			if p == "" || p == "index.html" {
				serveDemo(c)
				return
			}
			fileServer.ServeHTTP(c.Writer, c.Request)
		})
		log.Println("mounted /demo (embedded)")
		return
	}

	r.GET("/demo", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, `<!doctype html><html><head><meta charset="utf-8"><title>picca demo</title></head>
<body><h1>picca demo</h1>
<label>API Key: <input id="k" value="dummy-api-key-for-picca"></label>
<button id="run">Score</button><pre id="out"></pre>
<script>
const base=location.origin;
document.getElementById('run').onclick=async()=>{
  const key=(document.getElementById('k').value||'').trim();
  const body={fps:30,keypoints:[{x:0.10,y:0.20},{x:0.30,y:0.40}]};
  const r=await fetch(base+"/api/v1/score",{method:"POST",headers:{"Content-Type":"application/json","X-API-Key":key},body:JSON.stringify(body)});
  const j=await r.json().catch(()=>({error:"bad json"}));
  document.getElementById('out').textContent=JSON.stringify(j,null,2);
};
</script></body></html>`)
	})
	log.Println("mounted /demo (inline)")
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

func logReq(c *gin.Context, status int, latencyMs int64, inputHash, outputHash string) {
	if status < 400 {
		return
	}
	reqID := requestID(c)
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	if latencyMs < 0 {
		latencyMs = 0
	}
	entry := map[string]any{
		"ts":          time.Now().UTC().Format(time.RFC3339Nano),
		"run_id":      runID,
		"path":        path,
		"status":      status,
		"latency_ms":  latencyMs,
		"req_id":      reqID,
		"input_hash":  inputHash,
		"output_hash": outputHash,
	}
	line, err := json.Marshal(entry)
	if err != nil {
		log.Printf("log marshal error: %v", err)
		return
	}
	fmt.Println(string(line))
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
	if v, ok := c.Get("req_id"); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	reqID := strings.TrimSpace(c.GetHeader("X-Request-Id"))
	if reqID == "" {
		reqID = time.Now().UTC().Format(time.RFC3339Nano)
	}
	c.Header("X-Request-Id", reqID)
	c.Set("req_id", reqID)
	return reqID
}

func validateAPIKey(c *gin.Context) bool {
	expectedKey := os.Getenv("API_KEY")
	if expectedKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server misconfigured", "reason_code": "MISCONFIGURED_API_KEY"})
		logReq(c, http.StatusInternalServerError, 0, "", "")
		return false
	}
	if c.GetHeader("X-API-Key") != expectedKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "reason_code": "INVALID_API_KEY"})
		logReq(c, http.StatusUnauthorized, 0, "", "")
		return false
	}
	return true
}

func ensureJSONContentType(c *gin.Context) bool {
	if !isJSON(c.GetHeader("Content-Type")) {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "unsupported media", "reason_code": "UNSUPPORTED_MEDIA_TYPE"})
		logReq(c, http.StatusUnsupportedMediaType, 0, "", "")
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

	if !validateAPIKey(c) {
		return
	}
	if !ensureJSONContentType(c) {
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes())
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body", "reason_code": "INVALID_BODY"})
		logReq(c, http.StatusBadRequest, 0, "", "")
		return
	}

	mlURL := strings.TrimRight(os.Getenv("API_ML_URL"), "/")
	if mlURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server misconfigured", "reason_code": "MISCONFIGURED_UPSTREAM"})
		logReq(c, http.StatusInternalServerError, 0, "", "")
		return
	}

	upstreamReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, mlURL+"/predict", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "ml upstream error", "reason_code": "UPSTREAM_FAILURE"})
		logReq(c, http.StatusBadGateway, 0, "", "")
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
			logReq(c, http.StatusGatewayTimeout, duration, "", "")
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": "ml upstream error", "reason_code": "UPSTREAM_FAILURE"})
		logReq(c, http.StatusBadGateway, duration, "", "")
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "ml upstream error", "reason_code": "UPSTREAM_FAILURE"})
		logReq(c, http.StatusBadGateway, duration, "", "")
		return
	}

	c.Header("X-Request-Id", reqID)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
	logReq(c, resp.StatusCode, duration, "", "")
}

func explainOptionsHandler(c *gin.Context) {
	c.Header("Allow", "OPTIONS, POST")
	c.Header("Access-Control-Allow-Methods", "OPTIONS, POST")
	c.Header("Access-Control-Allow-Headers", "Content-Type,X-API-Key")
	c.Status(http.StatusOK)
}

func explainHandler(c *gin.Context) {
	reqID := requestID(c)

	if !validateAPIKey(c) {
		return
	}
	if !ensureJSONContentType(c) {
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes())
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body", "reason_code": "INVALID_BODY"})
		logReq(c, http.StatusBadRequest, 0, "", "")
		return
	}

	var payload explainRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body", "reason_code": "INVALID_BODY"})
		logReq(c, http.StatusBadRequest, 0, "", "")
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
		logReq(c, http.StatusInternalServerError, 0, "", "")
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
		logReq(c, http.StatusInternalServerError, 0, "", "")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	client, err := newVertexClient(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "vertex auth error", "reason_code": "VERTEX_AUTH_FAILURE"})
		logReq(c, http.StatusInternalServerError, 0, "", "")
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
		logReq(c, http.StatusBadGateway, 0, "", "")
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
			logReq(c, http.StatusGatewayTimeout, duration, "", "")
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": "vertex upstream error", "reason_code": "VERTEX_UPSTREAM_FAILURE"})
		logReq(c, http.StatusBadGateway, duration, "", "")
		return
	}
	defer vertexResp.Body.Close()

	respBody, err := io.ReadAll(vertexResp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "vertex upstream error", "reason_code": "VERTEX_UPSTREAM_FAILURE"})
		logReq(c, http.StatusBadGateway, duration, "", "")
		return
	}

	if vertexResp.StatusCode < 200 || vertexResp.StatusCode >= 300 {
		c.Data(vertexResp.StatusCode, vertexResp.Header.Get("Content-Type"), respBody)
		logReq(c, vertexResp.StatusCode, duration, "", "")
		return
	}

	summary, err := extractVertexSummary(respBody)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "vertex upstream error", "reason_code": "VERTEX_INVALID_RESPONSE"})
		logReq(c, http.StatusBadGateway, duration, "", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": summary,
		"model":   model,
		"region":  region,
	})
	logReq(c, http.StatusOK, duration, "", "")
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		requestID(c)
		c.Next()
	})

	mountDemo(r)
	mountAPI(r)

	r.GET("/v1/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "pong", "run_id": runID})
	})

	getRunID := func() string { return runID }
	healthHandler := NewHealthHandler(getRunID)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	mux := http.NewServeMux()
	mux.Handle("/healthz", healthHandler)
	mux.Handle("/livez", healthHandler)
	mux.Handle("/readyz", healthHandler)
	mux.Handle("/ops", OpsHandler())
	mux.Handle("/ops/", OpsHandler())
	mux.Handle("/api/", OTSMiddleware(runID, r))
	mux.Handle("/", r)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(shutdownCh)

	go func() {
		sig := <-shutdownCh
		log.Printf("server: received %s, initiating shutdown", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}()

	log.Printf("server ready on %s; run_id=%s", addr, runID)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %v", err)
	}
}
