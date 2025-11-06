package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

const maxCaptureBytes = 1 << 20 // 1 MiB cap for hashing/logging

type captureRW struct {
	http.ResponseWriter
	status int
	buf    bytes.Buffer
	wroteH bool
}

func (w *captureRW) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *captureRW) WriteHeader(code int) {
	if w.wroteH {
		return
	}
	w.status = code
	w.wroteH = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *captureRW) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	if w.buf.Len() < maxCaptureBytes {
		remaining := maxCaptureBytes - w.buf.Len()
		if remaining > len(b) {
			remaining = len(b)
		}
		w.buf.Write(b[:remaining])
	}
	return w.ResponseWriter.Write(b)
}

func sha16(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])[:16]
}

func newReqID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func isJSON(ct string) bool {
	ct = strings.ToLower(ct)
	return strings.Contains(ct, "json")
}

// OTSMiddleware wraps handlers to emit Observation Trace Sheet JSONL lines.
func OTSMiddleware(runID string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var inHash string
		if r.Body != nil && r.ContentLength != 0 && isJSON(r.Header.Get("Content-Type")) {
			body, err := io.ReadAll(io.LimitReader(r.Body, maxCaptureBytes+1))
			if err == nil && len(body) > 0 {
				if len(body) <= maxCaptureBytes {
					inHash = sha16(body)
				}
				r.Body = io.NopCloser(bytes.NewReader(body))
			} else {
				r.Body = http.NoBody
			}
		}

		crw := &captureRW{ResponseWriter: w}
		start := time.Now()
		next.ServeHTTP(crw, r)

		if crw.status == 0 {
			crw.status = http.StatusOK
		}
		latMS := math.Round(time.Since(start).Seconds()*1000*1e3) / 1e3

		var outHash string
		if isJSON(crw.Header().Get("Content-Type")) && crw.buf.Len() > 0 {
			if crw.buf.Len() <= maxCaptureBytes {
				outHash = sha16(crw.buf.Bytes())
			}
		}

		reqID := crw.Header().Get("X-Request-Id")
		if reqID == "" {
			reqID = r.Header.Get("X-Request-Id")
		}
		if reqID == "" {
			reqID = newReqID()
			crw.Header().Set("X-Request-Id", reqID)
		}

		rec := map[string]any{
			"ts":          time.Now().UTC().Format(time.RFC3339Nano),
			"run_id":      runID,
			"path":        r.URL.Path,
			"status":      crw.status,
			"latency_ms":  latMS,
			"req_id":      reqID,
			"input_hash":  inHash,
			"output_hash": outHash,
		}
		if line, err := json.Marshal(rec); err == nil {
			fmt.Println(string(line))
		}
	})
}
