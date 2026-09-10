// netductor-agent — outbound OpenWrt edge agent
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var version = "0.7.0-dev"

type config struct {
	Server   string
	Token    string
	DeviceID string
	Interval int
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "-v", "--version":
			fmt.Printf("netductor-agent %s\n", version)
			return
		case "help", "-h", "--help":
			fmt.Print(`netductor-agent — outbound edge agent for OpenWrt

Config /etc/netductor-agent/config:
  SERVER=http://vps:8787
  TOKEN=...
  DEVICE_ID=site1
  INTERVAL=60

Commands (from VPS):
  ping, status, metrics
  config_backup          — tar /etc/config → VPS
  uci_get|show|set|commit|batch
  wifi_reload, network_reload, reboot
  agent_update           — arg: URL or URL|sha256
  sysupgrade             — arg: URL|sha256|confirm=yes
`)
			return
		}
	}
	cfg := loadConfig()
	if cfg.Server == "" || cfg.Token == "" {
		fmt.Fprintln(os.Stderr, "SERVER and TOKEN required")
		os.Exit(1)
	}
	cfg.Server = strings.TrimRight(cfg.Server, "/")
	client := &http.Client{Timeout: 120 * time.Second}
	for {
		if err := heartbeat(client, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "heartbeat: %v\n", err)
		}
		if err := pollCmds(client, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "poll: %v\n", err)
		}
		time.Sleep(time.Duration(cfg.Interval) * time.Second)
	}
}

func loadConfig() config {
	c := config{Interval: 60, DeviceID: hostname()}
	for _, path := range []string{os.Getenv("NETDUCTOR_AGENT_CONF"), "/etc/netductor-agent/config"} {
		if path == "" {
			continue
		}
		if b, err := os.ReadFile(path); err == nil {
			parseKV(string(b), &c)
			break
		}
	}
	if v := os.Getenv("NETDUCTOR_SERVER"); v != "" {
		c.Server = v
	}
	if v := os.Getenv("NETDUCTOR_TOKEN"); v != "" {
		c.Token = v
	}
	if v := os.Getenv("NETDUCTOR_DEVICE_ID"); v != "" {
		c.DeviceID = v
	}
	if v := os.Getenv("NETDUCTOR_INTERVAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.Interval = n
		}
	}
	if c.DeviceID == "" {
		c.DeviceID = hostname()
	}
	return c
}

func parseKV(s string, c *config) {
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(k) {
		case "SERVER":
			c.Server = strings.TrimSpace(v)
		case "TOKEN":
			c.Token = strings.TrimSpace(v)
		case "DEVICE_ID":
			c.DeviceID = strings.TrimSpace(v)
		case "INTERVAL":
			if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n > 0 {
				c.Interval = n
			}
		}
	}
}

func hostname() string {
	h, _ := os.Hostname()
	return h
}

func boardName() string {
	for _, p := range []string{"/tmp/sysinfo/model", "/proc/device-tree/model"} {
		if b, err := os.ReadFile(p); err == nil {
			return strings.TrimSpace(string(b))
		}
	}
	return ""
}

func doJSON(client *http.Client, method, url, token string, body any) ([]byte, error) {
	var rdr io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, rdr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return data, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(data), 200))
	}
	return data, nil
}

func collectMetrics() map[string]any {
	m := map[string]any{
		"hostname": hostname(),
		"board":    boardName(),
		"agent":    version,
		"ts":       time.Now().Unix(),
	}
	// uptime
	if b, err := os.ReadFile("/proc/uptime"); err == nil {
		fields := strings.Fields(string(b))
		if len(fields) > 0 {
			m["uptime_sec"], _ = strconv.ParseFloat(fields[0], 64)
		}
	}
	// mem
	if b, err := os.ReadFile("/proc/meminfo"); err == nil {
		var total, avail float64
		for _, line := range strings.Split(string(b), "\n") {
			if strings.HasPrefix(line, "MemTotal:") {
				fmt.Sscanf(line, "MemTotal: %f", &total)
			}
			if strings.HasPrefix(line, "MemAvailable:") {
				fmt.Sscanf(line, "MemAvailable: %f", &avail)
			}
		}
		if total > 0 {
			m["mem_total_kb"] = total
			m["mem_avail_kb"] = avail
			m["mem_pct"] = (1 - avail/total) * 100
		}
	}
	// load
	if b, err := os.ReadFile("/proc/loadavg"); err == nil {
		f := strings.Fields(string(b))
		if len(f) >= 3 {
			m["load"] = map[string]string{"1": f[0], "5": f[1], "15": f[2]}
		}
	}
	// WAN IP
	if out, err := exec.Command("ip", "-4", "route", "get", "1.1.1.1").Output(); err == nil {
		// ... src x.x.x.x
		parts := strings.Fields(string(out))
		for i, p := range parts {
			if p == "src" && i+1 < len(parts) {
				m["wan_ip"] = parts[i+1]
				break
			}
		}
	}
	// wifi clients rough
	if out, err := exec.Command("iwinfo").Output(); err == nil {
		m["iwinfo"] = truncate(string(out), 1500)
	}
	return m
}

func heartbeat(client *http.Client, cfg config) error {
	payload := collectMetrics()
	payload["device_id"] = cfg.DeviceID
	_, err := doJSON(client, http.MethodPost, cfg.Server+"/api/edge/heartbeat", cfg.Token, payload)
	if err != nil {
		return err
	}
	// also store metrics history on VPS
	_, _ = doJSON(client, http.MethodPost, cfg.Server+"/api/edge/metrics", cfg.Token, payload)
	return nil
}

func pollCmds(client *http.Client, cfg config) error {
	url := fmt.Sprintf("%s/api/edge/commands?device_id=%s", cfg.Server, cfg.DeviceID)
	data, err := doJSON(client, http.MethodGet, url, cfg.Token, nil)
	if err != nil {
		return err
	}
	var wrap struct {
		Commands []struct {
			ID     string `json:"id"`
			Action string `json:"action"`
			Arg    string `json:"arg"`
		} `json:"commands"`
	}
	if json.Unmarshal(data, &wrap) != nil {
		return nil
	}
	for _, c := range wrap.Commands {
		if c.ID == "" || c.Action == "" {
			continue
		}
		res := runCmd(client, cfg, c.Action, c.Arg)
		_, _ = doJSON(client, http.MethodPost, cfg.Server+"/api/edge/cmd_result", cfg.Token, map[string]any{
			"device_id": cfg.DeviceID,
			"cmd_id":    c.ID,
			"action":    c.Action,
			"result":    res,
			"ts":        time.Now().Unix(),
		})
	}
	return nil
}

func runCmd(client *http.Client, cfg config, action, arg string) string {
	switch action {
	case "ping":
		return "pong"
	case "status", "metrics":
		b, _ := json.Marshal(collectMetrics())
		return string(b)
	case "reboot":
		go func() {
			time.Sleep(2 * time.Second)
			_ = exec.Command("reboot").Run()
		}()
		return "reboot scheduled"
	case "wifi_reload":
		out, _ := exec.Command("wifi", "reload").CombinedOutput()
		return truncate(string(out), 8000)
	case "network_reload":
		out, _ := exec.Command("/etc/init.d/network", "reload").CombinedOutput()
		return truncate(string(out), 8000)
	case "uci_get":
		out, _ := exec.Command("uci", "-q", "get", arg).CombinedOutput()
		return truncate(string(out), 8000)
	case "uci_show":
		args := []string{"-q", "show"}
		if arg != "" {
			args = append(args, arg)
		}
		out, _ := exec.Command("uci", args...).CombinedOutput()
		return truncate(string(out), 8000)
	case "uci_set":
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			return "uci_set: need path=value"
		}
		out, err := exec.Command("uci", "set", parts[0]+"="+parts[1]).CombinedOutput()
		if err != nil {
			return truncate(string(out)+" "+err.Error(), 8000)
		}
		return "ok"
	case "uci_commit":
		out, _ := exec.Command("uci", "commit").CombinedOutput()
		return truncate(string(out), 4000)
	case "uci_batch":
		return uciBatch(arg)
	case "logread":
		out, _ := exec.Command("logread").CombinedOutput()
		lines := strings.Split(string(out), "\n")
		if len(lines) > 80 {
			lines = lines[len(lines)-80:]
		}
		return strings.Join(lines, "\n")
	case "config_backup":
		return configBackup(client, cfg)
	case "agent_update":
		return agentUpdate(arg)
	case "sysupgrade":
		return doSysupgrade(arg)
	default:
		return "denied:" + action
	}
}

func uciBatch(arg string) string {
	var ok, fail int
	var logs []string
	for _, line := range strings.Split(arg, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if line == "commit" {
			out, err := exec.Command("uci", "commit").CombinedOutput()
			logs = append(logs, "commit:"+truncate(string(out), 200))
			if err != nil {
				fail++
			} else {
				ok++
			}
			continue
		}
		if line == "network_reload" {
			out, _ := exec.Command("/etc/init.d/network", "reload").CombinedOutput()
			logs = append(logs, "network_reload:"+truncate(string(out), 200))
			ok++
			continue
		}
		if line == "wifi_reload" {
			out, _ := exec.Command("wifi", "reload").CombinedOutput()
			logs = append(logs, "wifi_reload:"+truncate(string(out), 200))
			ok++
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			logs = append(logs, "bad:"+line)
			fail++
			continue
		}
		out, err := exec.Command("uci", "set", parts[0]+"="+parts[1]).CombinedOutput()
		if err != nil {
			logs = append(logs, "fail:"+parts[0]+":"+truncate(string(out), 100))
			fail++
		} else {
			ok++
		}
	}
	return fmt.Sprintf("ok=%d fail=%d\n%s", ok, fail, strings.Join(logs, "\n"))
}

func configBackup(client *http.Client, cfg config) string {
	tmp := filepath.Join(os.TempDir(), "nd-cfg-"+cfg.DeviceID+".tar.gz")
	_ = os.Remove(tmp)
	cmd := exec.Command("tar", "-czf", tmp, "-C", "/etc", "config")
	if out, err := cmd.CombinedOutput(); err != nil {
		// fallback busybox
		cmd = exec.Command("tar", "-czf", tmp, "/etc/config")
		out2, err2 := cmd.CombinedOutput()
		if err2 != nil {
			return "tar failed: " + truncate(string(out)+" "+string(out2), 500)
		}
	}
	defer os.Remove(tmp)
	f, err := os.Open(tmp)
	if err != nil {
		return err.Error()
	}
	defer f.Close()
	req, err := http.NewRequest(http.MethodPost, cfg.Server+"/api/edge/backup", f)
	if err != nil {
		return err.Error()
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	req.Header.Set("Content-Type", "application/gzip")
	req.Header.Set("X-Device-ID", cfg.DeviceID)
	resp, err := client.Do(req)
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Sprintf("upload HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	return "uploaded: " + truncate(string(body), 300)
}

func agentUpdate(arg string) string {
	url, wantSHA, _ := splitArg(arg)
	if url == "" {
		return "agent_update: need URL or URL|sha256"
	}
	tmp := filepath.Join(os.TempDir(), "netductor-agent.new")
	if err := downloadFile(url, tmp); err != nil {
		return "download: " + err.Error()
	}
	if wantSHA != "" {
		sum, err := fileSHA256(tmp)
		if err != nil || !strings.EqualFold(sum, wantSHA) {
			_ = os.Remove(tmp)
			return fmt.Sprintf("sha256 mismatch got=%s want=%s", sum, wantSHA)
		}
	}
	_ = os.Chmod(tmp, 0o755)
	dest := "/usr/sbin/netductor-agent"
	if _, err := os.Stat(dest); err != nil {
		dest = "/usr/bin/netductor-agent"
	}
	if err := os.Rename(tmp, dest); err != nil {
		// cross-device
		in, _ := os.ReadFile(tmp)
		if err2 := os.WriteFile(dest, in, 0o755); err2 != nil {
			return err2.Error()
		}
		_ = os.Remove(tmp)
	}
	return "updated " + dest + " — restart agent to run new binary"
}

func doSysupgrade(arg string) string {
	// URL|sha256|confirm=yes
	url, wantSHA, rest := splitArg(arg)
	if url == "" {
		return "sysupgrade: URL|sha256|confirm=yes"
	}
	if !strings.Contains(rest, "confirm=yes") && !strings.Contains(arg, "confirm=yes") {
		return "sysupgrade refused: add confirm=yes"
	}
	if _, err := exec.LookPath("sysupgrade"); err != nil {
		return "sysupgrade binary not found"
	}
	img := filepath.Join(os.TempDir(), "nd-firmware.bin")
	if err := downloadFile(url, img); err != nil {
		return "download: " + err.Error()
	}
	if wantSHA != "" {
		sum, err := fileSHA256(img)
		if err != nil || !strings.EqualFold(sum, wantSHA) {
			_ = os.Remove(img)
			return fmt.Sprintf("sha256 mismatch got=%s want=%s", sum, wantSHA)
		}
	}
	// -n keep config by default
	go func() {
		time.Sleep(3 * time.Second)
		_ = exec.Command("sysupgrade", "-n", img).Run()
	}()
	return "sysupgrade scheduled (keep config flags: default -n keep; image " + img + ")"
}

func splitArg(arg string) (url, sha, rest string) {
	parts := strings.Split(arg, "|")
	if len(parts) == 0 {
		return "", "", ""
	}
	url = strings.TrimSpace(parts[0])
	if len(parts) > 1 {
		sha = strings.TrimSpace(parts[1])
	}
	if len(parts) > 2 {
		rest = strings.Join(parts[2:], "|")
	}
	return url, sha, rest
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}
