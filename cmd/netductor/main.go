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
	"strings"
	"time"

	"github.com/PavelNeyman/netductor/internal/edge"
)

var version = "0.7.0-dev"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	switch cmd {
	case "version", "-v", "--version":
		fmt.Printf("netductor %s\n", version)
	case "help", "-h", "--help":
		printHelp()
	case "doctor":
		runBridge(lookPath("freshvps-doctor", "netductor-doctor"), args)
	case "vpn":
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "usage: netductor vpn ...")
			os.Exit(2)
		}
		runBridge(lookPath("freshvps-vpn"), args)
	case "edge":
		runEdgeCLI(args)
	case "status":
		runStatus()
	case "serve":
		runServe(args)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`netductor — personal network control plane

Commands: version | doctor | status | vpn | edge | serve | help

  serve [--bind] [--port] [--legacy URL] [--no-proxy]
`)
}

func lookPath(names ...string) string {
	for _, n := range names {
		if p, err := exec.LookPath(n); err == nil {
			return p
		}
		for _, dir := range []string{"/usr/local/bin", "/opt/freshvps/bin", "/opt/netductor/bin"} {
			cand := filepath.Join(dir, n)
			if st, err := os.Stat(cand); err == nil && !st.IsDir() {
				return cand
			}
		}
	}
	return ""
}

func runBridge(bin string, args []string) {
	if bin == "" {
		fmt.Fprintln(os.Stderr, "command not found (install stack or wait for native implementation)")
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
	bin := lookPath("freshvps-vpn")
	if bin == "" {
		// native list
		if args[0] == "list" {
			for _, d := range edge.ListDevices() {
				b, _ := json.Marshal(d)
				fmt.Println(string(b))
			}
			return
		}
		fmt.Fprintln(os.Stderr, "edge: freshvps-vpn not found")
		os.Exit(1)
	}
	var vpnArgs []string
	switch args[0] {
	case "list":
		vpnArgs = []string{"edge-list"}
	case "cmd":
		vpnArgs = append([]string{"edge-cmd"}, args[1:]...)
	default:
		os.Exit(2)
	}
	runBridge(bin, vpnArgs)
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

func runServe(args []string) {
	bind := envOr("NETDUCTOR_API_BIND", "127.0.0.1")
	port := envOr("NETDUCTOR_API_PORT", "8790")
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
			fmt.Println("netductor serve [--bind] [--port] [--legacy] [--no-proxy]")
			return
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"ok": true, "service": "netductor", "version": version,
			"time": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Native edge API (same storage as Python)
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
		did := r.URL.Query().Get("device_id")
		writeJSON(w, 200, map[string]any{"commands": edge.PollCommands(did)})
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
		// operator path: still require proxy/session via legacy for now if Authorization looks like session
		// For migration: list is readable with edge token OR we list openly only on localhost — prefer edge token OR empty for local ops
		auth := r.Header.Get("Authorization")
		if edge.ValidBearer(auth) || strings.HasPrefix(auth, "Bearer ") {
			writeJSON(w, 200, map[string]any{"devices": edge.ListDevices()})
			return
		}
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
	})
	mux.HandleFunc("/api/edge/cmd", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, 405, map[string]string{"error": "method"})
			return
		}
		auth := r.Header.Get("Authorization")
		if !edge.ValidBearer(auth) && !strings.HasPrefix(auth, "Bearer ") {
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
		id := edge.EnqueueCmd(did, action, arg)
		writeJSON(w, 200, map[string]any{"ok": true, "id": id})
	})

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
			if strings.HasPrefix(r.URL.Path, "/api/edge/") || r.URL.Path == "/health" {
				http.NotFound(w, r)
				return
			}
			rp.ServeHTTP(w, r)
		})
	}

	addr := bind + ":" + port
	fmt.Fprintf(os.Stderr, "netductor serve on http://%s (edge native, legacy proxy %v → %s)\n", addr, proxyOn, legacy)
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
