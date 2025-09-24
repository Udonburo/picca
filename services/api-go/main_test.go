package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// チE��ト用レスポンス形
// (retaining legacy comment text from original test file)
type mlResp struct {
	Score       int     `json:"score"`
	Symmetry    float64 `json:"symmetry"`
	Power       float64 `json:"power"`
	Consistency float64 `json:"consistency"`
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func newRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.POST("/api/v1/score", scoreHandler)
	r.POST("/api/v1/explain", explainHandler)
	r.GET("/healthz", func(c *gin.Context) { c.String(200, "ok") })
	return r
}

func TestScoreHandler_OK(t *testing.T) {
	// モチE��MLサーバ！Epredict ぁE00/JSONを返す�E�E
	ml := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/predict" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		io := mlResp{Score: 77, Symmetry: 0.8, Power: 0.7, Consistency: 0.9}
		_ = json.NewEncoder(w).Encode(io)
	}))
	defer ml.Close()

	t.Setenv("API_KEY", "secret")
	t.Setenv("API_ML_URL", ml.URL) // ↁEscoreHandler は末尾/を�E前で整形

	r := newRouter()

	body := []byte(`{"keypoints":[{"x":0.1,"y":0.2}],"fps":30}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/score", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "secret")
	req.Header.Set("X-Request-Id", "test-123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d; body=%s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("X-Request-Id"); got == "" {
		t.Fatalf("missing X-Request-Id echo")
	}
	var out mlResp
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if out.Score == 0 {
		t.Fatalf("unexpected score: %+v", out)
	}
}

func TestScoreHandler_InvalidAPIKey(t *testing.T) {
	ml := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ml.Close()

	t.Setenv("API_KEY", "secret")
	t.Setenv("API_ML_URL", ml.URL)

	r := newRouter()
	req := httptest.NewRequest("POST", "/api/v1/score", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "wrong")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestScoreHandler_UnsupportedMediaType(t *testing.T) {
	ml := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ml.Close()

	t.Setenv("API_KEY", "secret")
	t.Setenv("API_ML_URL", ml.URL)

	r := newRouter()
	req := httptest.NewRequest("POST", "/api/v1/score", bytes.NewBufferString("x"))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("X-API-Key", "secret")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("want 415, got %d", w.Code)
	}
}

func TestScoreHandler_MisconfiguredEnv(t *testing.T) {
	// API_KEY 未設宁EↁE500
	os.Unsetenv("API_KEY")
	os.Unsetenv("API_ML_URL")

	r := newRouter()
	req := httptest.NewRequest("POST", "/api/v1/score", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", w.Code)
	}
}

func TestScoreHandler_UpstreamTimeout(t *testing.T) {
	// 3sタイムアウトを趁E��る遅延を発生させる
	ml := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3500 * time.Millisecond)
	}))
	defer ml.Close()

	t.Setenv("API_KEY", "secret")
	t.Setenv("API_ML_URL", ml.URL)

	r := newRouter()
	req := httptest.NewRequest("POST", "/api/v1/score", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "secret")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusGatewayTimeout {
		t.Fatalf("want 504, got %d (body=%s)", w.Code, w.Body.String())
	}
}

func TestExplainHandler_OK(t *testing.T) {
	prevClient := newVertexClient
	t.Cleanup(func() { newVertexClient = prevClient })

	newVertexClient = func(ctx context.Context) (*http.Client, error) {
		return &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			req.Body.Close()
			if !strings.Contains(string(body), "Summarize these metrics") {
				return nil, fmt.Errorf("unexpected prompt: %s", string(body))
			}
			respBody := `{"candidates":[{"content":{"parts":[{"text":"Metrics look strong overall."}]}}]}`
			r := &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(respBody)),
			}
			r.Header.Set("Content-Type", "application/json")
			return r, nil
		})}, nil
	}

	t.Setenv("API_KEY", "secret")
	t.Setenv("PROJECT_ID", "demo-project")
	t.Setenv("VERTEX_REGION", "us-central1")
	t.Setenv("VERTEX_MODEL", "gemini-2.5-flash-lite")

	r := newRouter()

	body := []byte(`{"score":88.5,"symmetry":0.92,"power":0.81,"consistency":0.77}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/explain", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "secret")
	req.Header.Set("X-Request-Id", "test-explain")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d; body=%s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("X-Request-Id"); got == "" {
		t.Fatalf("missing X-Request-Id header")
	}

	var resp struct {
		Summary string `json:"summary"`
		Model   string `json:"model"`
		Region  string `json:"region"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.Summary == "" {
		t.Fatalf("missing summary in response: %+v", resp)
	}
	if resp.Model != "gemini-2.5-flash-lite" {
		t.Fatalf("unexpected model: %s", resp.Model)
	}
	if resp.Region != "us-central1" {
		t.Fatalf("unexpected region: %s", resp.Region)
	}
}
