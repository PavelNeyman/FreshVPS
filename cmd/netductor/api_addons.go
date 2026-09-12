package main

import (
	"net/http"

	"github.com/PavelNeyman/netductor/internal/addons"
)

func registerAddonsAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/addons", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		writeJSON(w, 200, addons.ListAddons())
	})
	mux.HandleFunc("/api/addons/lampac", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		writeJSON(w, 200, addons.CollectLampac())
	})
}
