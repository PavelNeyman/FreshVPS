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
	"github.com/PavelNeyman/netductor/internal/install"
	"github.com/PavelNeyman/netductor/internal/notify"
	"github.com/PavelNeyman/netductor/internal/paths"
	"github.com/PavelNeyman/netductor/internal/probes"
)

var version = "0.7.0-dev"

func main() {
	if len(os.Args) < 2 {
		// interactive when terminal; else help
		if fi, err := os.Stdin.Stat(); err == nil && (fi.Mode()&os.ModeCharDevice) != 0 {
			runTUI(nil)
			return
		}
		printHelp()
		os.Exit(0)
	}
	switch os.Args[1] {
	case "version", "-v", "--version":
		fmt.Printf("netductor %s\n", version)
	case "tui", "menu":
		runTUI(os.Args[2:])
	case "help", "-h", "--help":
		printHelp()
	case "doctor":
		if len(os.Args) > 2 && os.Args[2] == "--legacy" {
			fmt.Fprintln(os.Stderr, "legacy doctor removed"); os.Exit(2)
			return
		}
		os.Exit(runDoctorNative())
	case "vpn":
		runVPN(os.Args[2:])
	case "edge":
		runEdgeCLI(os.Args[2:])
	case "status":
		runStatus()
	case "install":
		runInstall(os.Args[2:])
	case "probe":
		runProbe(os.Args[2:])
	case "collect":
		os.Exit(runCollect())
	case "backup":
		runBackupCmd()
	case "restore":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: netductor restore <archive>")
			os.Exit(2)
		}
		if err := install.Restore(os.Args[2]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("restored")
	case "self-install":
		runSelfInstall()
	case "serve":
		runServe(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`netductor — network control plane

  tui|menu [--mode vps|openwrt|workstation|operator]
  backup | self-install
  version | doctor | status | vpn | edge | serve | install | probe | collect | help

  (no args on a TTY → interactive menu)

serve:
  --bind ADDR   (default 127.0.0.1)
  --port PORT   (default 8787)
  --tls-cert PATH --tls-key PATH
  --legacy URL  (optional)
  --no-proxy
`)
}

func lookPath(names ...string) string {
	for _, n := range names {
		if p, err := exec.LookPath(n); err == nil {
			return p
		}
		for _, d := range []string{"/usr/local/bin", "/opt/netductor/bin"} {
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
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: netductor edge list|pending|approve|deny|revoke|cmd …")
		os.Exit(2)
	}
	switch args[0] {
	case "list":
		runEdgeList()
	case "pending":
		for _, d := range edge.ListPending() {
			fmt.Printf("%v\t%v\t%v\t%v\n", d["device_id"], d["board"], d["wan_ip"], d["hostname"])
		}
	case "approve":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: netductor edge approve <device_id>")
			os.Exit(2)
		}
		tok, err := edge.Approve(args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("approved", args[1], "token="+tok)
	case "deny":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: netductor edge deny <device_id>")
			os.Exit(2)
		}
		if err := edge.Deny(args[1]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("denied")
	case "revoke":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: netductor edge revoke <device_id>")
			os.Exit(2)
		}
		if err := edge.Revoke(args[1]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("revoked")
	case "cmd":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: netductor edge cmd <device_id> <action> [arg]")
			os.Exit(2)
		}
		arg := ""
		if len(args) > 3 {
			arg = strings.Join(args[3:], " ")
		}
		id := edge.EnqueueCmd(args[1], args[2], arg)
		if id == "" {
			fmt.Fprintln(os.Stderr, "enqueue failed (device not approved?)")
			os.Exit(1)
		}
		fmt.Println(id)
	case "provision":
		// netductor edge provision root@host --id site1 --server https://vps --key path --agent path
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: netductor edge provision user@host --id DEVICE [--server URL] [--key KEY] [--agent BIN]")
			os.Exit(2)
		}
		opts := edge.ProvisionOpts{SSHTarget: args[1], ServerURL: envOr("NETDUCTOR_PUBLIC_URL", "")}
		for i := 2; i < len(args); i++ {
			switch args[i] {
			case "--id":
				if i+1 < len(args) {
					i++
					opts.DeviceID = args[i]
				}
			case "--server":
				if i+1 < len(args) {
					i++
					opts.ServerURL = args[i]
				}
			case "--key":
				if i+1 < len(args) {
					i++
					opts.SSHKey = args[i]
				}
			case "--agent":
				if i+1 < len(args) {
					i++
					opts.AgentBin = args[i]
				}
			}
		}
		if opts.DeviceID == "" || opts.ServerURL == "" {
			fmt.Fprintln(os.Stderr, "--id and --server (or NETDUCTOR_PUBLIC_URL) required")
			os.Exit(2)
		}
		if err := edge.Provision(opts); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("provisioned", opts.DeviceID, "→ pending enroll")
	case "templates":
		edge.EnsureDefaultTemplate()
		for _, tmpl := range edge.ListTemplates() {
			fmt.Println(tmpl["id"])
		}
	case "bind-template":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: netductor edge bind-template <device_id> <template_id>")
			os.Exit(2)
		}
		if err := edge.BindTemplate(args[1], args[2], nil); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("bound")
	default:
		fmt.Fprintln(os.Stderr, "unknown edge subcommand")
		os.Exit(2)
	}
}

func runStatus() {
	for _, u := range []string{"sing-box", "blocky", "netductor-api", "netductor-telegram-bot"} {
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

func runBackupCmd() {
	path, err := install.Backup()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(path)
}

func runSelfInstall() {
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	dest := "/usr/local/bin/netductor"
	data, err := os.ReadFile(exe)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	tmp := dest + ".new"
	if err := os.WriteFile(tmp, data, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.Rename(tmp, dest); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("installed", dest)
}

func runInstall(args []string) {
	comps := []string{}
	for _, a := range args {
		if a == "--help" || a == "-h" {
			fmt.Println("netductor install [--component name ...]   default: core stack")
			fmt.Println("components: dirs singbox blocky vpn-users api metrics telegram")
			return
		}
		if a == "--component" || a == "-c" {
			continue
		}
		if strings.HasPrefix(a, "-") {
			continue
		}
		comps = append(comps, a)
	}
	// allow: install singbox blocky
	if err := install.Run(install.Options{Components: comps}); err != nil {
		fmt.Fprintln(os.Stderr, err)
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


func runVPN(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: netductor vpn add|list|link|note|disable|enable|revoke|apply|session|edge-list|edge-cmd ...")
		os.Exit(2)
	}
	cmd := args[0]
	rest := args[1:]
	switch cmd {
	case "list":
		users, err := vpn.List()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, u := range users {
			en := "off"
			if u.Enabled {
				en = "on"
			}
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", u.Name, en, u.UUID, u.Note, u.Created)
		}
	case "add":
		if len(rest) < 1 {
			fmt.Fprintln(os.Stderr, "name required")
			os.Exit(2)
		}
		note := ""
		if len(rest) > 1 {
			note = rest[1]
		}
		out, err := vpn.Add(rest[0], note)
		fmt.Println(out)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "note":
		if len(rest) < 1 {
			os.Exit(2)
		}
		note := ""
		if len(rest) > 1 {
			note = rest[1]
		}
		out, err := vpn.Note(rest[0], note)
		fmt.Println(out)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "disable", "enable", "revoke":
		if len(rest) < 1 {
			os.Exit(2)
		}
		var out string
		var err error
		switch cmd {
		case "disable":
			out, err = vpn.Disable(rest[0])
		case "enable":
			out, err = vpn.Enable(rest[0])
		case "revoke":
			out, err = vpn.Revoke(rest[0])
		}
		fmt.Println(out)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "link":
		if len(rest) < 1 {
			os.Exit(2)
		}
		kind := "sub"
		if len(rest) > 1 {
			kind = rest[1]
		}
		name := rest[0]
		var s string
		var ok bool
		switch kind {
		case "vless":
			s, ok = vpn.ReadClient(name, "link-vless.txt", "link.txt")
		case "hy2":
			s, ok = vpn.ReadClient(name, "link-hy2.txt")
		default:
			s, ok = vpn.ReadClient(name, "subscription.txt", "link.txt")
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "not found")
			os.Exit(1)
		}
		fmt.Println(s)
	case "apply":
		if err := vpn.ApplyConfig(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("config applied")
	case "session":
		hours := 72
		if len(rest) > 0 {
			fmt.Sscanf(rest[0], "%d", &hours)
		}
		tok, exp, err := vpn.CreateSession(hours)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(tok)
		fmt.Fprintf(os.Stderr, "expires_unix=%d hours=%d\n", exp, hours)
	case "edge-list":
		runEdgeList()
	case "edge-cmd":
		if len(rest) < 2 {
			fmt.Fprintln(os.Stderr, "usage: netductor vpn edge-cmd <device_id> <action> [arg]")
			os.Exit(2)
		}
		arg := ""
		if len(rest) > 2 {
			arg = rest[2]
		}
		id := edge.EnqueueCmd(rest[0], rest[1], arg)
		fmt.Println(id)
	case "api-bind":
		mode := "localhost"
		ufw := false
		for _, a := range rest {
			if a == "--ufw" {
				ufw = true
				continue
			}
			mode = a
		}
		if err := vpn.APIBind(mode, ufw); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "unknown vpn subcommand"); os.Exit(2)
	}
}

func runEdgeList() {
	now := time.Now().Unix()
	devs := edge.ListDevices()
	// sort by last_seen desc roughly
	for _, d := range devs {
		did, _ := d["device_id"].(string)
		var last int64
		switch v := d["last_seen"].(type) {
		case float64:
			last = int64(v)
		case int64:
			last = v
		}
		st := "offline"
		if healthy, _ := d["healthy"].(bool); healthy {
			st = "online"
		}
		host, _ := d["hostname"].(string)
		wan, _ := d["wan_ip"].(string)
		board, _ := d["board"].(string)
		age := now - last
		if last == 0 {
			age = -1
		}
		fmt.Printf("%s\t%s\t%ds\t%s\t%s\t%s\n", did, st, age, host, wan, board)
	}
}

func runServe(args []string) {
	edge.EnsureDefaultTemplate()
	bind := envOr("NETDUCTOR_API_BIND", "127.0.0.1")
	port := envOr("NETDUCTOR_API_PORT", "8787")
	legacy := envOr("NETDUCTOR_LEGACY_API", "")
	proxyOn := false
	tlsCert := envOr("NETDUCTOR_TLS_CERT", "")
	tlsKey := envOr("NETDUCTOR_TLS_KEY", "")
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
				proxyOn = true
			}
		case "--no-proxy":
			proxyOn = false
		case "--tls-cert":
			if i+1 < len(args) {
				tlsCert, i = args[i+1], i+1
			}
		case "--tls-key":
			if i+1 < len(args) {
				tlsKey, i = args[i+1], i+1
			}
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
			mux.HandleFunc("/api/edge/backups", func(w http.ResponseWriter, r *http.Request) {
		// agent download with edge token OR operator session
		auth := r.Header.Get("Authorization")
		okEdge := edge.ValidBearer(auth)
		okSess := false
		if !okEdge {
			okSess = requireSession(w, r)
			if !okSess {
				return
			}
		}
		_ = okSess
		did := r.URL.Query().Get("device_id")
		if did == "" {
			writeJSON(w, 400, map[string]string{"error": "device_id"})
			return
		}
		name := r.URL.Query().Get("name")
		if name != "" {
			path, err := edge.BackupPath(did, name)
			if err != nil {
				writeJSON(w, 404, map[string]string{"error": "not found"})
				return
			}
			b, err := os.ReadFile(path)
			if err != nil {
				writeJSON(w, 500, map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/gzip")
			w.Header().Set("Content-Disposition", "attachment; filename="+name)
			w.WriteHeader(200)
			_, _ = w.Write(b)
			return
		}
		writeJSON(w, 200, map[string]any{"backups": edge.ListBackups(did)})
	})
	mux.HandleFunc("/api/edge/metrics/history", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		did := r.URL.Query().Get("device_id")
		writeJSON(w, 200, map[string]any{"metrics": edge.ListMetricsTail(did, 100)})
	})
	mux.HandleFunc("/api/edge/backup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, 405, map[string]string{"error": "method"})
			return
		}
		if !edge.ValidBearer(r.Header.Get("Authorization")) {
			writeJSON(w, 401, map[string]string{"error": "unauthorized"})
			return
		}
		did := r.Header.Get("X-Device-ID")
		if did == "" {
			did = "unknown"
		}
		path, err := edge.SaveBackup(did, r.Body)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]string{"ok": "true", "path": path})
	})
	mux.HandleFunc("/api/edge/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, 405, map[string]string{"error": "method"})
			return
		}
		if !edge.ValidBearer(r.Header.Get("Authorization")) {
			writeJSON(w, 401, map[string]string{"error": "unauthorized"})
			return
		}
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		did, _ := payload["device_id"].(string)
		if did == "" {
			did = "unknown"
		}
		_ = edge.SaveMetrics(did, payload)
		writeJSON(w, 200, map[string]string{"ok": "true"})
	})
			mux.HandleFunc("/api/edge/template", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		did := r.URL.Query().Get("device_id")
		if did == "" {
			did = edge.DeviceIDFromAuth(auth)
		}
		if r.Method == http.MethodGet {
			if !edge.RequireApproved(auth, did) {
				writeJSON(w, 403, map[string]string{"error": "not_approved"})
				return
			}
			tmpl, err := edge.TemplateForDevice(did)
			if err != nil {
				writeJSON(w, 404, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, 200, tmpl)
			return
		}
		writeJSON(w, 405, map[string]string{"error": "method"})
	})
	mux.HandleFunc("/api/edge/templates", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		edge.EnsureDefaultTemplate()
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, 200, map[string]any{"templates": edge.ListTemplates()})
		case http.MethodPost:
			body := readJSON(r)
			id, _ := body["id"].(string)
			if id == "" {
				writeJSON(w, 400, map[string]string{"error": "id"})
				return
			}
			if err := edge.SaveTemplate(id, edge.Template(body)); err != nil {
				writeJSON(w, 500, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, 200, map[string]any{"ok": true})
		default:
			writeJSON(w, 405, map[string]string{"error": "method"})
		}
	})
	mux.HandleFunc("/api/edge/bind-template", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !requireSession(w, r) {
			return
		}
		body := readJSON(r)
		did, _ := body["device_id"].(string)
		tid, _ := body["template_id"].(string)
		ov, _ := body["overlay"].(map[string]any)
		if err := edge.BindTemplate(did, tid, ov); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/edge/enroll", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, 405, map[string]string{"error": "method"})
			return
		}
		if !edge.ValidBootstrap(r.Header.Get("Authorization")) {
			writeJSON(w, 401, map[string]string{"error": "unauthorized"})
			return
		}
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		st, dtok, isNew := edge.Enroll(payload)
		if isNew && st == edge.StatusPending {
			did, _ := payload["device_id"].(string)
			board, _ := payload["board"].(string)
			wan, _ := payload["wan_ip"].(string)
			_ = notify.Telegram(fmt.Sprintf("⏳ Edge pending: <b>%s</b>\nboard=%s wan=%s", did, board, wan))
		}
		writeJSON(w, 200, map[string]any{"status": st, "device_token": dtok})
	})
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
	mux.HandleFunc("/api/edge/results", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		writeJSON(w, 200, map[string]any{"results": edge.ListResults()})
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
		mux.HandleFunc("/api/edge/approve", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !requireSession(w, r) {
			return
		}
		body := readJSON(r)
		did, _ := body["device_id"].(string)
		tok, err := edge.Approve(did)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		_ = notify.Telegram(fmt.Sprintf("✅ Edge approved: <b>%s</b>", did))
		writeJSON(w, 200, map[string]any{"ok": true, "device_id": did, "device_token": tok})
	})
	mux.HandleFunc("/api/edge/deny", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !requireSession(w, r) {
			return
		}
		body := readJSON(r)
		did, _ := body["device_id"].(string)
		if err := edge.Deny(did); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/edge/revoke", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !requireSession(w, r) {
			return
		}
		body := readJSON(r)
		did, _ := body["device_id"].(string)
		if err := edge.Revoke(did); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/edge/pending", func(w http.ResponseWriter, r *http.Request) {
		if !requireSession(w, r) {
			return
		}
		writeJSON(w, 200, map[string]any{"pending": edge.ListPending()})
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
		case action == "subscription" && r.Method == http.MethodGet:
			sub, ok := vpn.ReadClient(name, "subscription.txt", "link.txt")
			if !ok {
				writeJSON(w, 404, map[string]string{"error": "not found"})
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(sub + "\n"))
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
	if tlsCert != "" && tlsKey != "" {
		fmt.Fprintf(os.Stderr, "netductor serve TLS on https://%s\n", addr)
		if err := http.ListenAndServeTLS(addr, tlsCert, tlsKey, mux); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
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
