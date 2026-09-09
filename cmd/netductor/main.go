// Netductor — network control plane CLI (orchestrator).
package main

import (
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
		runDoctor(args)
	case "vpn":
		runVPN(args)
	case "edge":
		runEdge(args)
	case "status":
		runStatus()
	case "serve":
		runServe(args)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`netductor — personal network control plane

Usage:
  netductor <command> [args]

Commands:
  version          Print version
  doctor           Health checks (bridge to freshvps-doctor)
  status           Short systemd status
  vpn list|...     VPN users (bridge to freshvps-vpn)
  edge list|cmd    Edge agents (bridge)
  serve            HTTP API (Go; proxies legacy Python API)
  help             This help

Legacy FreshVPS CLIs remain supported during migration.
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

func runDoctor(args []string) {
	bin := lookPath("freshvps-doctor", "netductor-doctor")
	if bin == "" {
		fmt.Fprintln(os.Stderr, "doctor: freshvps-doctor not found")
		os.Exit(1)
	}
	c := exec.Command(bin, args...)
	c.Stdout, c.Stderr, c.Stdin = os.Stdout, os.Stderr, os.Stdin
	if err := c.Run(); err != nil {
		os.Exit(exitCode(err))
	}
}

func runVPN(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: netductor vpn list|add|link|...")
		os.Exit(2)
	}
	bin := lookPath("freshvps-vpn")
	if bin == "" {
		fmt.Fprintln(os.Stderr, "vpn: freshvps-vpn not found")
		os.Exit(1)
	}
	c := exec.Command(bin, args...)
	c.Stdout, c.Stderr, c.Stdin = os.Stdout, os.Stderr, os.Stdin
	if err := c.Run(); err != nil {
		os.Exit(exitCode(err))
	}
}

func runEdge(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: netductor edge list|cmd <device_id> <action> [arg]")
		os.Exit(2)
	}
	bin := lookPath("freshvps-vpn")
	if bin == "" {
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
		fmt.Fprintf(os.Stderr, "edge: unknown subcommand %q\n", args[0])
		os.Exit(2)
	}
	c := exec.Command(bin, vpnArgs...)
	c.Stdout, c.Stderr, c.Stdin = os.Stdout, os.Stderr, os.Stdin
	if err := c.Run(); err != nil {
		os.Exit(exitCode(err))
	}
}

func runStatus() {
	units := []string{"sing-box", "blocky", "freshvps-api", "freshvps-telegram-bot"}
	fmt.Println("netductor status")
	for _, u := range units {
		out, err := exec.Command("systemctl", "is-active", u).Output()
		st := strings.TrimSpace(string(out))
		if err != nil && st == "" {
			st = "inactive"
		}
		fmt.Printf("  %s: %s\n", u, st)
	}
}

// runServe: G3 — Go front door. /health is native; everything else can proxy to legacy Python API.
func runServe(args []string) {
	bind := envOr("NETDUCTOR_API_BIND", "127.0.0.1")
	port := envOr("NETDUCTOR_API_PORT", "8790")
	legacy := envOr("NETDUCTOR_LEGACY_API", "http://127.0.0.1:8787")
	proxyOn := true
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--bind":
			if i+1 < len(args) {
				bind = args[i+1]
				i++
			}
		case "--port":
			if i+1 < len(args) {
				port = args[i+1]
				i++
			}
		case "--legacy":
			if i+1 < len(args) {
				legacy = args[i+1]
				i++
			}
		case "--no-proxy":
			proxyOn = false
		case "--help", "-h":
			fmt.Println("netductor serve [--bind ADDR] [--port PORT] [--legacy URL] [--no-proxy]")
			fmt.Println("  /health — native Go")
			fmt.Println("  other paths — reverse-proxy to legacy Python API (default http://127.0.0.1:8787)")
			return
		}
	}
	addr := bind + ":" + port
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"ok":true,"service":"netductor","version":%q,"time":%q}`, version, time.Now().UTC().Format(time.RFC3339))
	})
	if proxyOn {
		u, err := url.Parse(legacy)
		if err != nil {
			fmt.Fprintln(os.Stderr, "bad --legacy URL:", err)
			os.Exit(1)
		}
		rp := httputil.NewSingleHostReverseProxy(u)
		rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintf(w, `{"error":"legacy_api_unreachable","detail":%q,"legacy":%q}`, e.Error(), legacy)
		}
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				return
			}
			rp.ServeHTTP(w, r)
		})
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			io.WriteString(w, `{"error":"not_implemented","hint":"enable proxy or use Python :8787"}`)
		})
	}
	fmt.Fprintf(os.Stderr, "netductor serve on http://%s (legacy proxy %v → %s)\n", addr, proxyOn, legacy)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode()
	}
	return 1
}
