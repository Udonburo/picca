package main

import (
	"log"
	"net/http"
)

// /healthz: simple liveness. GET/HEAD -> 200, others -> 405.
func healthz(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	case http.MethodHead:
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func init() {
	http.HandleFunc("/healthz", healthz)
	log.Printf("mounted /healthz")
}
