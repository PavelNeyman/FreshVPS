package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/PavelNeyman/netductor/internal/paths"
	"github.com/PavelNeyman/netductor/internal/vpn"
)

func registerRelayAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/relay/export", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		sni := r.URL.Query().Get("sni")
		if sni == "" {
			sni = "ya.ru"
		}
		b, err := vpn.ExportRelayBundle(sni)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(b)
	})
	mux.HandleFunc("/api/relay/status", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		p := filepath.Join(paths.StateDir(), "relay", "bundle.json")
		out := map[string]any{"path": p, "has_bundle": false}
		if raw, err := os.ReadFile(p); err == nil {
			out["has_bundle"] = true
			var v any
			_ = json.Unmarshal(raw, &v)
			out["bundle"] = v
		}
		writeJSON(w, 200, out)
	})
}
