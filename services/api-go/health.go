package main

import (
	"fmt"
	"net/http"
)

// NewHealthHandler serves liveness/readiness probes with run_id information.
func NewHealthHandler(getRunID func() string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Cache-Control", "no-store, max-age=0")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		if r.Method == http.MethodHead {
			return
		}
		fmt.Fprintf(w, `{"ok":true,"run_id":"%s"}`, getRunID())
	})
}
