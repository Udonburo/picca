package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// テスト用レスポンス形
type mlResp struct {
	Score       int     `json:"score"`
	Symmetry    float64 `json:"symmetry"`
	Power       float64 `json:"power"`
	Consistency float64 `json:"consistency"`
}

func newRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.POST("/api/v1/score", scoreHandler)
	r.GET("/healthz", func(c *gin.Context) { c.String(200, "ok") })
	return r
}

func TestScoreHandler_OK(t *testing.T) {
	// モックMLサーバ（/predict が200/JSONを返す）
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
	t.Setenv("API_ML_URL", ml.URL) // ← scoreHandler は末尾/を自前で整形

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
	// API_KEY 未設定 → 500
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
	// 3sタイムアウトを超える遅延を発生させる
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
