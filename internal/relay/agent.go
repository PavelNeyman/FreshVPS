package relay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/PavelNeyman/netductor/internal/paths"
	"github.com/PavelNeyman/netductor/internal/vpn"
)

// AgentLoop runs on RU VPS: heartbeat + pull config when version drifts.
func AgentLoop(coreBase, token string, interval time.Duration) {
	if interval < 10*time.Second {
		interval = 30 * time.Second
	}
	client := &http.Client{Timeout: 20 * time.Second}
	applied := 0
	for {
		_ = agentTick(client, coreBase, token, &applied)
		time.Sleep(interval)
	}
}

func agentTick(client *http.Client, coreBase, token string, applied *int) error {
	coreBase = strings.TrimRight(coreBase, "/")
	pub := readSecret("singbox_reality_public")
	sid := readSecret("singbox_short_id")
	sni := strings.TrimSpace(readFile(filepath.Join(paths.EtcDir(), "secrets", "singbox_reality_sni")))
	if sni == "" {
		sni = "ya.ru"
	}
	ip := publicIP()
	sbOK := false
	if out, err := exec.Command("systemctl", "is-active", "sing-box").Output(); err == nil {
		sbOK = strings.TrimSpace(string(out)) == "active"
	}
	cpu, memU, memT, load1 := sampleMetrics()
	body, _ := json.Marshal(HeartbeatIn{
		PublicIP: ip, PBK: pub, SID: sid, SNI: sni,
		Version: "agent-1", SingBoxOK: sbOK, ConfigVer: *applied,
		CPUPercent: cpu, MemUsedMB: memU, MemTotalMB: memT, Load1: load1,
	})
	req, err := http.NewRequest(http.MethodPost, coreBase+"/api/relay/agent/heartbeat", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("heartbeat %s: %s", resp.Status, string(raw))
	}
	var hr struct {
		ConfigVer int `json:"config_ver"`
		NeedSync  bool `json:"need_sync"`
	}
	_ = json.Unmarshal(raw, &hr)
	if hr.NeedSync || hr.ConfigVer > *applied {
		if err := pullAndApply(client, coreBase, token); err != nil {
			return err
		}
		*applied = hr.ConfigVer
	}
	return nil
}

func pullAndApply(client *http.Client, coreBase, token string) error {
	req, err := http.NewRequest(http.MethodGet, coreBase+"/api/relay/agent/config", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("config %s: %s", resp.Status, string(raw))
	}
	var b vpn.RelayBundle
	if err := json.Unmarshal(raw, &b); err != nil {
		return err
	}
	priv := readSecret("singbox_reality_private")
	sid := readSecret("singbox_short_id")
	if err := vpn.WriteRelaySingBox(&b, priv, sid); err != nil {
		return err
	}
	_ = exec.Command("systemctl", "restart", "sing-box").Run()
	_ = os.WriteFile(filepath.Join(paths.StateDir(), "relay", "bundle.json"), raw, 0o600)
	return nil
}

func readSecret(name string) string {
	b, _ := os.ReadFile(filepath.Join(paths.EtcDir(), "secrets", name))
	return strings.TrimSpace(string(b))
}
func readFile(p string) string {
	b, _ := os.ReadFile(p)
	return string(b)
}
func publicIP() string {
	if v := os.Getenv("PUBLIC_IP"); v != "" {
		return v
	}
	out, err := exec.Command("curl", "-4", "-fsS", "--max-time", "5", "https://ifconfig.me").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}


func sampleMetrics() (cpu float64, memUsed, memTotal int64, load1 float64) {
	// loadavg
	if b, err := os.ReadFile("/proc/loadavg"); err == nil {
		fmt.Sscanf(string(b), "%f", &load1)
	}
	// meminfo
	if b, err := os.ReadFile("/proc/meminfo"); err == nil {
		var total, avail int64
		for _, line := range strings.Split(string(b), "\n") {
			if strings.HasPrefix(line, "MemTotal:") {
				fmt.Sscanf(line, "MemTotal: %d", &total)
			}
			if strings.HasPrefix(line, "MemAvailable:") {
				fmt.Sscanf(line, "MemAvailable: %d", &avail)
			}
		}
		memTotal = total / 1024
		if total > 0 {
			memUsed = (total - avail) / 1024
		}
	}
	cpu = load1 * 50 // rough indicator on small VPS
	if cpu > 100 {
		cpu = 100
	}
	return
}
