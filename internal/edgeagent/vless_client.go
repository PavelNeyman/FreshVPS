package edgeagent

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// VLESSClientConfig builds a minimal sing-box client JSON from a vless:// link.
func VLESSClientConfig(link string) ([]byte, error) {
	link = strings.TrimSpace(link)
	if !strings.HasPrefix(link, "vless://") {
		return nil, fmt.Errorf("not vless link")
	}
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	uuid := u.User.Username()
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "443"
	}
	q := u.Query()
	sni := q.Get("sni")
	if sni == "" {
		sni = q.Get("serverName")
	}
	if sni == "" {
		sni = host
	}
	pbk := q.Get("pbk")
	sid := q.Get("sid")
	flow := q.Get("flow")
	fp := q.Get("fp")
	if fp == "" {
		fp = "chrome"
	}
	cfg := map[string]any{
		"log": map[string]any{"level": "info"},
		"inbounds": []any{
			map[string]any{
				"type": "tun", "tag": "tun-in",
				"interface_name": "nd-tun",
				"inet4_address":  "172.19.0.1/30",
				"auto_route":     true,
				"strict_route":   true,
				"stack":          "system",
			},
		},
		"outbounds": []any{
			map[string]any{
				"type":        "vless",
				"tag":         "proxy",
				"server":      host,
				"server_port": atoi(port),
				"uuid":        uuid,
				"flow":        flow,
				"tls": map[string]any{
					"enabled":     true,
					"server_name": sni,
					"utls":        map[string]any{"enabled": true, "fingerprint": fp},
					"reality": map[string]any{
						"enabled":    true,
						"public_key": pbk,
						"short_id":   sid,
					},
				},
			},
			map[string]any{"type": "direct", "tag": "direct"},
		},
		"route": map[string]any{
			"final": "proxy",
			"auto_detect_interface": true,
		},
	}
	return json.MarshalIndent(cfg, "", "  ")
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	if n == 0 {
		return 443
	}
	return n
}
