package metrics

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var Dir = env("FRESHVPS_METRICS_DIR", "/var/lib/freshvps/metrics")

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func Collect() map[string]any {
	out := map[string]any{
		"ts": time.Now().Unix(),
	}
	// loadavg
	if b, err := os.ReadFile("/proc/loadavg"); err == nil {
		p := strings.Fields(string(b))
		if len(p) >= 3 {
			out["load1"], _ = strconv.ParseFloat(p[0], 64)
			out["load5"], _ = strconv.ParseFloat(p[1], 64)
			out["load15"], _ = strconv.ParseFloat(p[2], 64)
		}
	}
	// mem
	mem := map[string]int64{}
	if f, err := os.Open("/proc/meminfo"); err == nil {
		sc := bufio.NewScanner(f)
		kv := map[string]int64{}
		for sc.Scan() {
			fs := strings.Fields(sc.Text())
			if len(fs) >= 2 {
				n, _ := strconv.ParseInt(fs[1], 10, 64)
				kv[strings.TrimSuffix(fs[0], ":")] = n * 1024
			}
		}
		f.Close()
		mem["total"] = kv["MemTotal"]
		mem["available"] = kv["MemAvailable"]
		if mem["available"] == 0 {
			mem["available"] = kv["MemFree"]
		}
		mem["used"] = mem["total"] - mem["available"]
	}
	out["mem"] = mem
	// uptime
	if b, err := os.ReadFile("/proc/uptime"); err == nil {
		p := strings.Fields(string(b))
		if len(p) >= 1 {
			out["uptime_sec"], _ = strconv.ParseFloat(p[0], 64)
		}
	}
	return out
}

func History(limit int) []map[string]any {
	if limit <= 0 {
		limit = 180
	}
	path := filepath.Join(Dir, "history.jsonl")
	b, err := os.ReadFile(path)
	if err != nil {
		return []map[string]any{}
	}
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	out := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		var m map[string]any
		if json.Unmarshal([]byte(line), &m) == nil {
			out = append(out, m)
		}
	}
	return out
}
