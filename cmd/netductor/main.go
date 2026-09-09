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
	"github.com/PavelNeyman/netductor/internal/vpn"
	"github.com/PavelNeyman/netductor/internal/paths"
	"github.com/PavelNeyman/netductor/internal/probes"
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
	case "install":
		runInstall(os.Args[2:])
	case "probe":
		runProbe(os.Args[2:])
	case "serve":
		runServe(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`netductor — network control plane

  version | doctor | status | vpn | edge | serve | install | probe | help

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
	return paths.AdminRoot()
}

func runInstall(args []string) {
	// G4 scaffold: bridge to legacy install.sh if present in cwd or /opt/freshvps
	candidates := []string{"install.sh", "/opt/freshvps/install.sh", "/opt/netductor/install.sh"}
	var script string
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			script = c
			break
		}
	}
	if script == "" {
		fmt.Fprintln(os.Stderr, "install.sh not found — clone repo or use bootstrap.sh")
		os.Exit(1)
	}
	a := append([]string{script}, args...)
	c := exec.Command("bash", a...)
	c.Stdout, c.Stderr, c.Stdin = os.Stdout, os.Stderr, os.Stdin
	if err := c.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			os.Exit(ee.ExitCode())
		}
		os.Exit(1)
	}
}

func runProbe(args []string) {
	cfg := probes.Load()
	results := probes.Run(cfg)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(map[string]any{"probes": results})
	fail := 0
	for _, r := range results {
		if ok, _ := r["ok"].(bool); !ok {
			fail++
		}
	}
	if fail > 0 {
		os.Exit(1)
	}
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



	mux.HandleFunc("/api/probes", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		writeJSON(w, 200, map[string]any{"probes": probes.Run(probes.Load()), "config": probes.Load()})
	})
	mux.HandleFunc("/api/latest", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		writeJSON(w, 200, map[string]any{"metrics": metrics.Collect(), "probes": probes.Run(probes.Load())})
	})
	mux.HandleFunc("/api/probes/uptime", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		limit := 1440
		if s := r.URL.Query().Get("limit"); s != "" {
			if n, err := strconv.Atoi(s); err == nil {
				limit = n
			}
		}
		writeJSON(w, 200, map[string]any{"uptime": metrics.ProbeUptime(limit)})
	})
	mux.HandleFunc("/api/probes/config", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, 200, probes.Load())
		case http.MethodPost, http.MethodPut:
			cfg := readJSON(r)
			if err := probes.Save(cfg); err != nil {
				writeJSON(w, 500, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, 200, map[string]any{"ok": true})
		default:
			writeJSON(w, 405, map[string]string{"error": "method"})
		}
	})
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		m := metrics.Collect()
		writeJSON(w, 200, map[string]any{
			"ok": true, "service": "netductor", "version": version,
			"metrics": m, "probes": probes.Run(probes.Load()),
		})
	})
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		m := metrics.Collect()
		writeJSON(w, 200, map[string]any{"ok": true, "metrics": m, "probes": probes.Run(probes.Load())})
	})

	// --- VPN users (operator session) ---
	mux.HandleFunc("/vpn/users", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		switch r.Method {
		case http.MethodGet:
			users, err := vpn.List()
			if err != nil {
				writeJSON(w, 500, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, 200, map[string]any{"users": users})
		case http.MethodPost:
			body := readJSON(r)
			name, _ := body["name"].(string)
			note, _ := body["note"].(string)
			name = strings.TrimSpace(name)
			note = strings.TrimSpace(note)
			if !vpn.ValidName(name) {
				writeJSON(w, 400, map[string]string{"error": "bad name"})
				return
			}
			out, err := vpn.Add(name, note)
			if err != nil {
				writeJSON(w, 500, map[string]string{"error": strings.TrimSpace(out)})
				return
			}
			writeJSON(w, 200, map[string]any{"ok": true, "output": strings.TrimSpace(out), "name": name})
		default:
			writeJSON(w, 405, map[string]string{"error": "method"})
		}
	})
	mux.HandleFunc("/vpn/users/", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/vpn/users/")
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) < 2 {
			writeJSON(w, 404, map[string]string{"error": "not found"})
			return
		}
		name, action := parts[0], parts[1]
		if !vpn.ValidName(name) {
			writeJSON(w, 400, map[string]string{"error": "bad name"})
			return
		}
		switch {
		case action == "qr" && r.Method == http.MethodGet:
			qp := vpn.QRPath(name)
			b, err := os.ReadFile(qp)
			if err != nil {
				writeJSON(w, 404, map[string]string{"error": "no qr"})
				return
			}
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(200)
			_, _ = w.Write(b)
		case action == "link" && r.Method == http.MethodGet:
			sub, ok := vpn.ReadClient(name, "subscription.txt", "link.txt")
			if !ok {
				writeJSON(w, 404, map[string]string{"error": "not found"})
				return
			}
			vless, _ := vpn.ReadClient(name, "link-vless.txt", "link.txt")
			hy2, _ := vpn.ReadClient(name, "link-hy2.txt")
			writeJSON(w, 200, map[string]any{"name": name, "subscription": sub, "vless": vless, "hy2": hy2})
		case action == "note" && r.Method == http.MethodPost:
			note, _ := readJSON(r)["note"].(string)
			out, err := vpn.Note(name, note)
			if err != nil {
				writeJSON(w, 500, map[string]string{"error": strings.TrimSpace(out)})
				return
			}
			writeJSON(w, 200, map[string]any{"ok": true, "output": strings.TrimSpace(out)})
		case action == "disable" && r.Method == http.MethodPost:
			out, err := vpn.Disable(name)
			if err != nil {
				writeJSON(w, 500, map[string]string{"error": strings.TrimSpace(out)})
				return
			}
			writeJSON(w, 200, map[string]any{"ok": true, "output": strings.TrimSpace(out)})
		case action == "enable" && r.Method == http.MethodPost:
			out, err := vpn.Enable(name)
			if err != nil {
				writeJSON(w, 500, map[string]string{"error": strings.TrimSpace(out)})
				return
			}
			writeJSON(w, 200, map[string]any{"ok": true, "output": strings.TrimSpace(out)})
		case action == "revoke" && r.Method == http.MethodPost:
			out, err := vpn.Revoke(name)
			if err != nil {
				writeJSON(w, 500, map[string]string{"error": strings.TrimSpace(out)})
				return
			}
			writeJSON(w, 200, map[string]any{"ok": true, "output": strings.TrimSpace(out)})
		default:
			writeJSON(w, 404, map[string]string{"error": "not found"})
		}
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
				p == "/api/probes" || p == "/api/probes/uptime" || p == "/api/probes/config" ||
				p == "/api/latest" || p == "/api/status" || p == "/status" ||
				p == "/api/metrics" || p == "/api/metrics/history" || p == "/metrics" ||
				strings.HasPrefix(p, "/admin") || strings.HasPrefix(p, "/vpn/") {
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
