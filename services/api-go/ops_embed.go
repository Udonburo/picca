package main

import (
	_ "embed"
	"net/http"
)

//go:embed ops.html
var opsHTML []byte

// OpsHandler serves the embedded ops.html over net/http.
func OpsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(opsHTML)
	})
}
