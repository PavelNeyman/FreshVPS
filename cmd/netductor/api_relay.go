package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PavelNeyman/netductor/internal/relay"
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
		if id, tok, err := relay.IssueToken("relay"); err == nil {
			b.AgentID = id
			b.AgentToken = tok
			b.CoreAgentURL = "http://" + b.CoreIP + ":8788"
		}
		writeJSON(w, 200, b)
	})
	mux.HandleFunc("/api/relay/status", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		devs := relay.List()
		type row struct {
			relay.Device
			Online bool `json:"online"`
		}
		var out []row
		for _, d := range devs {
			out = append(out, row{Device: d, Online: relay.Online(d, 2*time.Minute)})
		}
		writeJSON(w, 200, map[string]any{
			"config_ver": relay.ConfigVer(),
			"devices":    out,
		})
	})
	mux.HandleFunc("/api/relay/exit", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		if r.Method == http.MethodPost {
			var body struct {
				Enabled *bool `json:"enabled"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			on := r.URL.Query().Get("enabled") == "1" || r.URL.Query().Get("on") == "1"
			if body.Enabled != nil {
				on = *body.Enabled
			}
			_ = relay.SetExitEnabled(on)
			_ = vpn.ApplyConfig()
		}
		writeJSON(w, 200, map[string]any{"exit_enabled": relay.ExitEnabled()})
	})
	mux.HandleFunc("/api/relay/links", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		// mobile links from last heartbeat identity
		devs := relay.List()
		var links []map[string]string
		reg, _ := vpn.ExportRelayBundle("ya.ru") // users only
		for _, d := range devs {
			if d.PublicIP == "" || d.PBK == "" {
				continue
			}
			for _, u := range reg.Users {
				links = append(links, map[string]string{
					"relay": d.ID,
					"name":  u.Name,
					"link":  vpn.ClientLinkForRelay(u.Name, u.UUID, d.PublicIP, d.PBK, d.SID, d.SNI),
				})
			}
		}
		writeJSON(w, 200, map[string]any{"links": links})
	})

	// Agent-facing (token = device token)
	mux.HandleFunc("/api/relay/agent/heartbeat", handleRelayAgentHeartbeat)
	mux.HandleFunc("/api/relay/agent/config", handleRelayAgentConfig)
}

func agentToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return r.Header.Get("X-Relay-Token")
}

func handleRelayAgentHeartbeat(w http.ResponseWriter, r *http.Request) {
	tok := agentToken(r)
	if tok == "" {
		http.Error(w, "unauthorized", 401)
		return
	}
	raw, _ := io.ReadAll(r.Body)
	var in relay.HeartbeatIn
	_ = json.Unmarshal(raw, &in)
	d, ver, err := relay.Heartbeat(tok, in)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	writeJSON(w, 200, map[string]any{
		"ok": true, "config_ver": ver, "need_sync": in.ConfigVer < ver, "id": d.ID,
	})
}

func handleRelayAgentConfig(w http.ResponseWriter, r *http.Request) {
	tok := agentToken(r)
	if relay.FindByToken(tok) == nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	b, err := vpn.ExportRelayBundle("ya.ru")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// do not issue new agent token on pull
	b.AgentToken = ""
	b.AgentID = ""
	writeJSON(w, 200, b)
}

// startRelayAgentListener binds :8788 for agent plane (all interfaces).
func startRelayAgentListener() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/relay/agent/heartbeat", handleRelayAgentHeartbeat)
	mux.HandleFunc("/api/relay/agent/config", handleRelayAgentConfig)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok\n"))
	})
	addr := os.Getenv("NETDUCTOR_RELAY_API")
	if addr == "" {
		addr = ":8788"
	}
	go func() {
		_ = http.ListenAndServe(addr, mux)
	}()
}
