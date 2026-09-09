package vpn

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/PavelNeyman/netductor/internal/paths"
)

const singboxConf = "/usr/local/etc/sing-box/config.json"

func ApplyConfig() error {
	if err := EnsureDirs(); err != nil {
		return err
	}
	priv := secret("singbox_reality_private")
	sid := secret("singbox_short_id")
	if priv == "" || sid == "" {
		return fmt.Errorf("Reality secrets missing")
	}
	r, err := loadRegistry()
	if err != nil {
		return err
	}
	type vu struct {
		UUID string `json:"uuid"`
		Flow string `json:"flow"`
	}
	type hu struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	var vusers []vu
	var husers []hu
	for _, u := range r.Users {
		if !u.Enabled {
			continue
		}
		vusers = append(vusers, vu{UUID: u.UUID, Flow: "xtls-rprx-vision"})
		pass := u.Hy2Password
		if pass == "" {
			pass = u.UUID
		}
		husers = append(husers, hu{Name: u.Name, Password: pass})
	}
	if vusers == nil {
		vusers = []vu{}
	}
	if husers == nil {
		husers = []hu{}
	}

	sniVal := sni()
	cfg := map[string]any{
		"log": map[string]any{"level": "info", "timestamp": true},
		"dns": map[string]any{
			"servers": []any{
				map[string]any{"type": "udp", "tag": "blocky", "server": "127.0.0.1"},
				map[string]any{"type": "local", "tag": "local"},
			},
			"final": "blocky", "strategy": "ipv4_only",
		},
		"inbounds": []any{
			map[string]any{
				"type": "vless", "tag": "vless-reality", "listen": "::", "listen_port": vlessPort(),
				"users": vusers,
				"tls": map[string]any{
					"enabled": true, "server_name": sniVal,
					"reality": map[string]any{
						"enabled": true,
						"handshake": map[string]any{"server": sniVal, "server_port": 443},
						"private_key": priv, "short_id": []string{sid},
					},
				},
			},
			map[string]any{
				"type": "hysteria2", "tag": "hy2", "listen": "::", "listen_port": hy2Port(),
				"users": husers,
				"tls": map[string]any{
					"enabled": true, "alpn": []string{"h3"},
					"certificate_path": "/etc/sing-box/certs/hy2.crt",
					"key_path":         "/etc/sing-box/certs/hy2.key",
				},
				"masquerade": "https://" + sniVal,
			},
		},
		"outbounds": []any{map[string]any{"type": "direct", "tag": "direct"}},
		"route": map[string]any{
			"rules": []any{
				map[string]any{"action": "sniff"},
				map[string]any{"protocol": "dns", "action": "hijack-dns"},
			},
			"final": "direct", "auto_detect_interface": true, "default_domain_resolver": "blocky",
		},
	}

	work, err := os.MkdirTemp("", "nd-sb-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)

	writeJSON := func(path string, v any) error {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(path, append(b, '\n'), 0o600)
	}
	aPath := filepath.Join(work, "a.json")
	if err := writeJSON(aPath, cfg); err != nil {
		return err
	}
	// variant b: local dns only
	cfgB := cloneMap(cfg)
	cfgB["dns"] = map[string]any{"servers": []any{map[string]any{"type": "local", "tag": "local"}}, "final": "local"}
	if route, ok := cfgB["route"].(map[string]any); ok {
		route["default_domain_resolver"] = "local"
	}
	bPath := filepath.Join(work, "b.json")
	_ = writeJSON(bPath, cfgB)
	// variant c: minimal route
	cfgC := cloneMap(cfg)
	delete(cfgC, "dns")
	cfgC["route"] = map[string]any{"final": "direct", "auto_detect_interface": true}
	cPath := filepath.Join(work, "c.json")
	_ = writeJSON(cPath, cfgC)

	chosen, variant := "", ""
	for _, pair := range []struct{ p, v string }{{aPath, "a"}, {bPath, "b"}, {cPath, "c"}} {
		if tryCheck(pair.p) {
			chosen, variant = pair.p, pair.v
			break
		}
	}
	if chosen == "" {
		return fmt.Errorf("sing-box config validation failed for all variants")
	}
	_ = os.MkdirAll("/usr/local/etc/sing-box", 0o755)
	b, err := os.ReadFile(chosen)
	if err != nil {
		return err
	}
	if err := os.WriteFile(singboxConf, b, 0o600); err != nil {
		return err
	}
	_ = os.WriteFile(filepath.Join(paths.EtcDir(), "singbox-config-variant"), []byte(variant+"\n"), 0o644)
	_ = exec.Command("systemctl", "restart", "sing-box").Run()
	fmt.Fprintf(os.Stderr, "sing-box config variant=%s\n", variant)
	return nil
}

func tryCheck(conf string) bool {
	bin := "/usr/local/bin/sing-box"
	if _, err := os.Stat(bin); err != nil {
		return true // no binary → accept first
	}
	return exec.Command(bin, "check", "-c", conf).Run() == nil
}

func cloneMap(m map[string]any) map[string]any {
	b, _ := json.Marshal(m)
	var out map[string]any
	_ = json.Unmarshal(b, &out)
	return out
}
