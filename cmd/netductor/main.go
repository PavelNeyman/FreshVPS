package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PavelNeyman/netductor/internal/edge"
	"github.com/PavelNeyman/netductor/internal/metrics"
	"github.com/PavelNeyman/netductor/internal/session"
)

var version = "0.7.0-dev"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}
	switch os.Args[1] {
	case "version", "-v", "--version":
		fmt.Printf("netductor %s\n", version)
	case "help", "-h", "--help":
		printHelp()
	case "doctor":
		runBridge(lookPath("freshvps-doctor"), os.Args[2:])
	case "vpn":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: netductor vpn ...")
			os.Exit(2)
		}
		runBridge(lookPath("freshvps-vpn"), os.Args[2:])
	case "edge":
		runEdgeCLI(os.Args[2:])
	case "status":
		runStatus()
	case "serve":
		runServe(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`netductor — network control plane

  version | doctor | status | vpn | edge | serve | help

serve:
  --bind ADDR   (default 127.0.0.1)
  --port PORT   (default 8790; use 8787 to replace Python)
  --legacy URL  (default http://127.0.0.1:8787)
  --no-proxy
`)
}

func lookPath(names ...string) string {
	for _, n := range names {
		if p, err := exec.LookPath(n); err == nil {
			return p
		}
		for _, d := range []string{"/usr/local/bin", "/opt/freshvps/bin", "/opt/netductor/bin"} {
			c := filepath.Join(d, n)
			if st, err := os.Stat(c); err == nil && !st.IsDir() {
				return c
			}
		}
	}
	return ""
}

func runBridge(bin string, args []string) {
	if bin == "" {
		fmt.Fprintln(os.Stderr, "binary not found")
		os.Exit(1)
	}
	c := exec.Command(bin, args...)
	c.Stdout, c.Stderr, c.Stdin = os.Stdout, os.Stderr, os.Stdin
	if err := c.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			os.Exit(ee.ExitCode())
		}
		os.Exit(1)
	}
}

func runEdgeCLI(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: netductor edge list|cmd ...")
		os.Exit(2)
	}
	if args[0] == "list" {
		if bin := lookPath("freshvps-vpn"); bin != "" {
			runBridge(bin, []string{"edge-list"})
			return
		}
		enc := json.NewEncoder(os.Stdout)
		for _, d := range edge.ListDevices() {
			_ = enc.Encode(d)
		}
		return
	}
	if args[0] == "cmd" {
		runBridge(lookPath("freshvps-vpn"), append([]string{"edge-cmd"}, args[1:]...))
		return
	}
	os.Exit(2)
}

func runStatus() {
	for _, u := range []string{"sing-box", "blocky", "freshvps-api", "freshvps-telegram-bot"} {
		out, _ := exec.Command("systemctl", "is-active", u).Output()
		st := strings.TrimSpace(string(out))
		if st == "" {
			st = "inactive"
		}
		fmt.Printf("  %s: %s\n", u, st)
	}
}

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

func requireSession(w http.ResponseWriter, r *http.Request) bool {
	tok := bearer(r)
	if !session.Valid(tok) {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return false
	}
	return true
}

func adminRoot() string {
	if v := os.Getenv("FRESHVPS_ADMIN_ROOT"); v != "" {
		return v
	}
	if v := os.Getenv("NETDUCTOR_ADMIN_ROOT"); v != "" {
		return v
	}
	for _, p := range []string{"/opt/freshvps/runtime/api/admin", "/opt/netductor/runtime/api/admin"} {
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p
		}
	}
	// repo path when developing
	if st, err := os.Stat("runtime/api/admin"); err == nil && st.IsDir() {
		return "runtime/api/admin"
	}
	return "/opt/freshvps/runtime/api/admin"
}

func runServe(args []string) {
	bind := envOr("NETDUCTOR_API_BIND", envOr("VPN_API_BIND", "127.0.0.1"))
	port := envOr("NETDUCTOR_API_PORT", envOr("VPN_API_PORT", "8790"))
	legacy := envOr("NETDUCTOR_LEGACY_API", "http://127.0.0.1:8787")
	proxyOn := true
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--bind":
			if i+1 < len(args) {
				bind, i = args[i+1], i+1
			}
		case "--port":
			if i+1 < len(args) {
				port, i = args[i+1], i+1
			}
		case "--legacy":
			if i+1 < len(args) {
				legacy, i = args[i+1], i+1
			}
		case "--no-proxy":
			proxyOn = false
		case "--help", "-h":
			printHelp()
			return
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"ok": true, "service": "netductor", "version": version, "time": time.Now().UTC().Format(time.RFC3339)})
	})

	// --- edge (device token) ---
	mux.HandleFunc("/api/edge/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, 405, map[string]string{"error": "method"})
			return
		}
		if !edge.ValidBearer(r.Header.Get("Authorization")) {
			writeJSON(w, 401, map[string]string{"error": "unauthorized"})
			return
		}
		edge.Heartbeat(readJSON(r))
		writeJSON(w, 200, map[string]bool{"ok": true})
	})
	mux.HandleFunc("/api/edge/commands", func(w http.ResponseWriter, r *http.Request) {
		if !edge.ValidBearer(r.Header.Get("Authorization")) {
			writeJSON(w, 401, map[string]string{"error": "unauthorized"})
			return
		}
		writeJSON(w, 200, map[string]any{"commands": edge.PollCommands(r.URL.Query().Get("device_id"))})
	})
	mux.HandleFunc("/api/edge/cmd_result", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, 405, map[string]string{"error": "method"})
			return
		}
		if !edge.ValidBearer(r.Header.Get("Authorization")) {
			writeJSON(w, 401, map[string]string{"error": "unauthorized"})
			return
		}
		edge.CmdResult(readJSON(r))
		writeJSON(w, 200, map[string]bool{"ok": true})
	})
	mux.HandleFunc("/api/edge/devices", func(w http.ResponseWriter, r *http.Request) {
		tok := bearer(r)
		if !edge.ValidBearer(r.Header.Get("Authorization")) && !session.Valid(tok) {
			writeJSON(w, 401, map[string]string{"error": "unauthorized"})
			return
		}
		writeJSON(w, 200, map[string]any{"devices": edge.ListDevices()})
	})
	mux.HandleFunc("/api/edge/cmd", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, 405, map[string]string{"error": "method"})
			return
		}
		tok := bearer(r)
		if !edge.ValidBearer(r.Header.Get("Authorization")) && !session.Valid(tok) {
			writeJSON(w, 401, map[string]string{"error": "unauthorized"})
			return
		}
		body := readJSON(r)
		did, _ := body["device_id"].(string)
		action, _ := body["action"].(string)
		arg, _ := body["arg"].(string)
		if did == "" || action == "" {
			writeJSON(w, 400, map[string]string{"error": "device_id and action required"})
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true, "id": edge.EnqueueCmd(did, action, arg)})
	})

	// --- operator session ---
	mux.HandleFunc("/api/session", func(w http.ResponseWriter, r *http.Request) {
		tok := bearer(r)
		exp, ok := session.Expiry(tok)
		if !ok {
			writeJSON(w, 401, map[string]string{"error": "unauthorized"})
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true, "expires_unix": exp})
	})
	mux.HandleFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		writeJSON(w, 200, metrics.Collect())
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		writeJSON(w, 200, metrics.Collect())
	})
	mux.HandleFunc("/api/metrics/history", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		limit := 180
		if s := r.URL.Query().Get("limit"); s != "" {
			if n, err := strconv.Atoi(s); err == nil {
				limit = n
			}
		}
		writeJSON(w, 200, map[string]any{"points": metrics.History(limit)})
	})

	// --- Admin SPA ---
	root := adminRoot()
	mux.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin/", http.StatusFound)
	})
	mux.Handle("/admin/", http.StripPrefix("/admin/", http.FileServer(http.Dir(root))))

	// --- legacy proxy for remaining API ---
	if proxyOn {
		u, err := url.Parse(legacy)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		rp := httputil.NewSingleHostReverseProxy(u)
		rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
			writeJSON(w, 502, map[string]string{"error": "legacy_api_unreachable", "detail": e.Error()})
		}
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			p := r.URL.Path
			if p == "/health" || strings.HasPrefix(p, "/api/edge/") || p == "/api/session" ||
				p == "/api/metrics" || p == "/api/metrics/history" || p == "/metrics" ||
				strings.HasPrefix(p, "/admin") {
				http.NotFound(w, r)
				return
			}
			rp.ServeHTTP(w, r)
		})
	}

	addr := bind + ":" + port
	fmt.Fprintf(os.Stderr, "netductor serve on http://%s admin=%s proxy=%v→%s\n", addr, root, proxyOn, legacy)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
