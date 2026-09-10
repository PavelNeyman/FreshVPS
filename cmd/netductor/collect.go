package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/PavelNeyman/netductor/internal/metrics"
	"github.com/PavelNeyman/netductor/internal/notify"
	"github.com/PavelNeyman/netductor/internal/paths"
	"github.com/PavelNeyman/netductor/internal/probes"
)

func runCollect() int {
	dir := paths.MetricsDir()
	_ = os.MkdirAll(dir, 0o755)

	lockPath := filepath.Join(dir, "collector.lock")
	lf, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer lf.Close()
	// best-effort exclusive lock via O_EXCL style: write pid if empty
	// (full flock needs syscall; skip concurrent races for MVP)

	cfg := probes.Load()
	// ensure probes.cfg exists
	if _, err := os.Stat(probes.CfgPath()); err != nil {
		_ = probes.Save(cfg)
	}

	m := metrics.Collect()
	mem, _ := m["mem"].(map[string]int64)
	if mem != nil {
		total := mem["total"]
		if total <= 0 {
			total = 1
		}
		m["mem_pct"] = float64(mem["used"]) / float64(total) * 100
	}
	disk, _ := m["disk"].(map[string]any)
	if disk != nil {
		var total, used float64
		switch v := disk["total"].(type) {
		case int64:
			total = float64(v)
		case int:
			total = float64(v)
		case float64:
			total = v
		}
		switch v := disk["used"].(type) {
		case int64:
			used = float64(v)
		case int:
			used = float64(v)
		case float64:
			used = v
		}
		if total <= 0 {
			total = 1
		}
		m["disk_pct"] = used / total * 100
	}

	live := probes.Run(cfg)
	m["probes"] = live

	// history
	hist := filepath.Join(dir, "history.jsonl")
	line, _ := json.Marshal(m)
	f, err := os.OpenFile(hist, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err == nil {
		_, _ = f.Write(append(line, '\n'))
		_ = f.Close()
		// trim to last 10000 lines
		if b, err := os.ReadFile(hist); err == nil {
			lines := splitKeep(b, 10000)
			_ = os.WriteFile(hist, lines, 0o644)
		}
	}
	latest := filepath.Join(dir, "latest.json")
	_ = os.WriteFile(latest, append(line, '\n'), 0o644)

	evaluateSimpleAlerts(m, live, cfg)
	fmt.Printf("collected ts=%v probes=%d → %s\n", m["ts"], len(live), latest)
	return 0
}

func splitKeep(b []byte, max int) []byte {
	// keep last max lines
	n := 0
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] == '\n' {
			n++
			if n > max {
				return b[i+1:]
			}
		}
	}
	return b
}

func evaluateSimpleAlerts(m map[string]any, live []map[string]any, cfg map[string]any) {
	al, _ := cfg["alerts"].(map[string]any)
	if al == nil {
		return
	}
	send := func(msg string) { _ = notify.Telegram(msg) }
	// probe failures
	if v, ok := al["probe_fail"].(bool); ok && v {
		for _, p := range live {
			if ok, _ := p["ok"].(bool); !ok {
				name, _ := p["name"].(string)
				err, _ := p["error"].(string)
				send(fmt.Sprintf("⚠️ Probe %s failed: %s", name, err))
			}
		}
	}
	_ = time.Now() // keep import if needed
	_ = m
}
