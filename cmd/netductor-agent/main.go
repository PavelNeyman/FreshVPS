// netductor-agent — outbound OpenWrt edge agent (G5).
package main

import (
	"bufio"
	"bytes"
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
			fmt.Print(`netductor-agent — outbound edge agent

  Config file (key=value): /etc/netductor-agent/config
    SERVER=https://vps.example
    TOKEN=...
    DEVICE_ID=site1
    INTERVAL=60

  Env overrides: NETDUCTOR_SERVER, NETDUCTOR_TOKEN, NETDUCTOR_DEVICE_ID, NETDUCTOR_INTERVAL
  Legacy path: 
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
	client := &http.Client{Timeout: 20 * time.Second}
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
	for _, path := range []string{
		os.Getenv("NETDUCTOR_AGENT_CONF"),
		"/etc/netductor-agent/config",
		"",
	} {
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
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		switch k {
		case "SERVER":
			c.Server = v
		case "TOKEN":
			c.Token = v
		case "DEVICE_ID":
			c.DeviceID = v
		case "INTERVAL":
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				c.Interval = n
			}
		}
	}
}

func hostname() string {
	b, err := os.ReadFile("/proc/sys/kernel/hostname")
	if err != nil {
		h, _ := os.Hostname()
		return h
	}
	return strings.TrimSpace(string(b))
}

func collect() map[string]any {
	m := map[string]any{
		"device_id": hostname(),
		"hostname":  hostname(),
		"ts":        time.Now().Unix(),
	}
	if b, err := os.ReadFile("/proc/uptime"); err == nil {
		f := strings.Fields(string(b))
		if len(f) > 0 {
			if u, err := strconv.ParseFloat(f[0], 64); err == nil {
				m["uptime"] = int64(u)
			}
		}
	}
	if b, err := os.ReadFile("/proc/loadavg"); err == nil {
		f := strings.Fields(string(b))
		if len(f) >= 3 {
			m["load"] = strings.Join(f[:3], " ")
		}
	}
	mem := map[string]int{}
	if f, err := os.Open("/proc/meminfo"); err == nil {
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			fs := strings.Fields(sc.Text())
			if len(fs) >= 2 {
				n, _ := strconv.Atoi(fs[1])
				mem[strings.TrimSuffix(fs[0], ":")] = n
			}
		}
		f.Close()
		m["mem_total_kb"] = mem["MemTotal"]
		m["mem_avail_kb"] = mem["MemAvailable"]
	}
	if b, err := os.ReadFile("/etc/openwrt_release"); err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			if strings.HasPrefix(line, "DISTRIB_RELEASE=") {
				m["openwrt"] = strings.Trim(strings.TrimPrefix(line, "DISTRIB_RELEASE="), "'\"")
			}
		}
	}
	for _, p := range []string{"/tmp/sysinfo/model", "/tmp/sysinfo/board_name"} {
		if b, err := os.ReadFile(p); err == nil {
			m["board"] = strings.TrimSpace(string(b))
			break
		}
	}
	m["wan_ip"] = wanIP()
	m["wifi_clients"] = wifiClients()
	return m
}

func wanIP() string {
	out, err := exec.Command("ip", "-4", "route", "get", "1.1.1.1").Output()
	if err != nil {
		return ""
	}
	fs := strings.Fields(string(out))
	for i, f := range fs {
		if f == "src" && i+1 < len(fs) {
			return fs[i+1]
		}
	}
	return ""
}

func wifiClients() int {
	if _, err := exec.LookPath("iwinfo"); err != nil {
		return 0
	}
	out, err := exec.Command("iwinfo").Output()
	if err != nil {
		return 0
	}
	total := 0
	var ifaces []string
	sc := bufio.NewScanner(bytes.NewReader(out))
	var prev string
	for sc.Scan() {
		line := sc.Text()
		if strings.Contains(line, "ESSID") && prev != "" {
			ifaces = append(ifaces, strings.Fields(prev)[0])
		}
		prev = line
	}
	for _, iface := range ifaces {
		o, err := exec.Command("iwinfo", iface, "assoclist").Output()
		if err != nil {
			continue
		}
		total += bytes.Count(o, []byte("dBm"))
	}
	return total
}

func doJSON(client *http.Client, method, url, token string, body any) ([]byte, error) {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
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
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return data, fmt.Errorf("HTTP %d: %s", resp.StatusCode, bytes.TrimSpace(data))
	}
	return data, nil
}

func heartbeat(client *http.Client, cfg config) error {
	payload := collect()
	payload["device_id"] = cfg.DeviceID
	_, err := doJSON(client, http.MethodPost, cfg.Server+"/api/edge/heartbeat", cfg.Token, payload)
	return err
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
		res := runCmd(c.Action, c.Arg)
		_, _ = doJSON(client, http.MethodPost, cfg.Server+"/api/edge/cmd_result", cfg.Token, map[string]any{
			"device_id": cfg.DeviceID,
			"cmd_id":    c.ID,
			"result":    res,
		})
	}
	return nil
}

func runCmd(action, arg string) string {
	switch action {
	case "ping":
		return "pong"
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
	case "logread":
		out, _ := exec.Command("logread").CombinedOutput()
		lines := strings.Split(string(out), "\n")
		if len(lines) > 50 {
			lines = lines[len(lines)-50:]
		}
		return strings.Join(lines, "\n")
	case "uci_set":
		// arg: path=value
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			return "uci_set: need path=value"
		}
		out, err := exec.Command("uci", "set", parts[0]+"="+parts[1]).CombinedOutput()
		if err != nil {
			return truncate(string(out)+" "+err.Error(), 8000)
		}
		_ = exec.Command("uci", "commit").Run()
		return "ok:" + truncate(string(out), 200)
	case "uci_commit":
		out, _ := exec.Command("uci", "commit").CombinedOutput()
		return truncate(string(out), 8000)
	case "status":
		return collectStatusJSON()
	default:
		return "denied:" + action
	}
}

func collectStatusJSON() string {
	host, _ := os.Hostname()
	return fmt.Sprintf(`{"hostname":%q,"board":%q,"ok":true}`, host, boardName())
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) > n {
		return s[:n]
	}
	return s
}

// silence unused on non-openwrt builds
var _ = filepath.Join

func boardName() string {
	b, err := os.ReadFile("/tmp/sysinfo/model")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}
