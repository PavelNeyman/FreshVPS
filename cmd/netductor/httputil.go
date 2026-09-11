package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/PavelNeyman/netductor/internal/paths"
	"github.com/PavelNeyman/netductor/internal/session"
)

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request) map[string]any {
	defer r.Body.Close()
	var m map[string]any
	_ = json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&m)
	if m == nil {
		m = map[string]any{}
	}
	return m
}

func bearer(r *http.Request) string {
	return session.TokenFromAuth(r.Header.Get("Authorization"), r.Header.Get("Cookie"))
}

func securityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
}

func requireSession(w http.ResponseWriter, r *http.Request) bool {
	securityHeaders(w)
	tok := bearer(r)
	if tok == "" || !session.Valid(tok) {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return false
	}
	return true
}

func adminRoot() string {
	return paths.AdminRoot()
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
