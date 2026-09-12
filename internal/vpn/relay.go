package vpn

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/PavelNeyman/netductor/internal/paths"
)

const RelayUplinkName = "relay-uplink"

// RelayBundle is generated on core and consumed on RU relay VPS.
type RelayBundle struct {
	Version    int    `json:"version"`
	CreatedAt  string `json:"created_at"`
	CoreIP     string `json:"core_ip"`
	CoreVless  int    `json:"core_vless_port"`
	CoreSNI    string `json:"core_sni"`
	CorePBK    string `json:"core_pbk"`
	CoreSID    string `json:"core_sid"`
	UplinkUUID string `json:"uplink_uuid"`
	// Inbound on relay (own Reality keypair)
	RelaySNI string `json:"relay_sni"`
	// Users allowed on relay inbound (same UUIDs as core for familiar client profiles)
	Users []RelayUser `json:"users"`
}

type RelayUser struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
}

// EnsureRelayUplink creates a dedicated core user for RU→core traffic.
func EnsureRelayUplink() (uuid string, err error) {
	r, err := loadRegistry()
	if err != nil {
		return "", err
	}
	if u := findUser(r, RelayUplinkName); u != nil {
		return u.UUID, nil
	}
	msg, err := AddNative(RelayUplinkName, "RU relay uplink — do not give to end users")
	if err != nil {
		return "", fmt.Errorf("%s: %w", msg, err)
	}
	r, err = loadRegistry()
	if err != nil {
		return "", err
	}
	u := findUser(r, RelayUplinkName)
	if u == nil {
		return "", fmt.Errorf("uplink user missing after create")
	}
	return u.UUID, nil
}

// ExportRelayBundle builds join material for a RU VPS.
func ExportRelayBundle(relaySNI string) (*RelayBundle, error) {
	if relaySNI == "" {
		relaySNI = "ya.ru"
	}
	up, err := EnsureRelayUplink()
	if err != nil {
		return nil, err
	}
	r, err := loadRegistry()
	if err != nil {
		return nil, err
	}
	var users []RelayUser
	for _, u := range r.Users {
		if !u.Enabled || u.Name == RelayUplinkName {
			continue
		}
		users = append(users, RelayUser{Name: u.Name, UUID: u.UUID})
	}
	b := &RelayBundle{
		Version:    1,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		CoreIP:     publicIP(),
		CoreVless:  vlessPort(),
		CoreSNI:    sni(),
		CorePBK:    secret("singbox_reality_public"),
		CoreSID:    secret("singbox_short_id"),
		UplinkUUID: up,
		RelaySNI:   relaySNI,
		Users:      users,
	}
	if b.CorePBK == "" || b.CoreSID == "" {
		return nil, fmt.Errorf("core Reality secrets missing")
	}
	dir := filepath.Join(paths.StateDir(), "relay")
	_ = os.MkdirAll(dir, 0o700)
	raw, _ := json.MarshalIndent(b, "", "  ")
	_ = os.WriteFile(filepath.Join(dir, "bundle.json"), append(raw, '\n'), 0o600)
	return b, nil
}

// WriteRelaySingBox generates /usr/local/etc/sing-box/config.json for RU hop.
// Inbound: VLESS Reality (relay keys) for end users.
// Outbound: VLESS Vision to core as relay-uplink.
func WriteRelaySingBox(b *RelayBundle, privKey, shortID string) error {
	if b == nil {
		return fmt.Errorf("nil bundle")
	}
	if privKey == "" || shortID == "" {
		return fmt.Errorf("relay reality keys required")
	}
	type vu struct {
		UUID string `json:"uuid"`
		Flow string `json:"flow"`
	}
	var users []vu
	for _, u := range b.Users {
		users = append(users, vu{UUID: u.UUID, Flow: "xtls-rprx-vision"})
	}
	if len(users) == 0 {
		// still allow uplink-only test with one placeholder skipped — require ≥1 user
		return fmt.Errorf("bundle has no end users — add vpn users on core first")
	}
	cfg := map[string]any{
		"log": map[string]any{"level": "info", "timestamp": true},
		"inbounds": []any{
			map[string]any{
				"type": "vless", "tag": "relay-in", "listen": "::", "listen_port": 443,
				"users": users,
				"tls": map[string]any{
					"enabled": true, "server_name": b.RelaySNI,
					"reality": map[string]any{
						"enabled": true,
						"handshake": map[string]any{
							"server": b.RelaySNI, "server_port": 443,
						},
						"private_key": privKey,
						"short_id":    []string{shortID},
					},
				},
			},
		},
		"outbounds": []any{
			map[string]any{
				"type": "vless", "tag": "uplink",
				"server": b.CoreIP, "server_port": b.CoreVless,
				"uuid": b.UplinkUUID, "flow": "xtls-rprx-vision",
				"tls": map[string]any{
					"enabled": true, "server_name": b.CoreSNI,
					"utls":    map[string]any{"enabled": true, "fingerprint": "chrome"},
					"reality": map[string]any{
						"enabled": true,
						"public_key": b.CorePBK,
						"short_id":   b.CoreSID,
					},
				},
			},
			map[string]any{"type": "direct", "tag": "direct"},
		},
		"route": map[string]any{
			"rules": []any{
				map[string]any{"action": "sniff"},
				map[string]any{"inbound": []string{"relay-in"}, "outbound": "uplink"},
			},
			"final": "uplink",
		},
	}
	_ = os.MkdirAll(filepath.Dir(singboxConf), 0o755)
	raw, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp := singboxConf + ".tmp"
	if err := os.WriteFile(tmp, append(raw, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, singboxConf)
}

// ClientLinkForRelay builds VLESS URI pointing at RU public IP (not core).
func ClientLinkForRelay(name, uuid, relayIP, pbk, sid, sniName string) string {
	if sniName == "" {
		sniName = "ya.ru"
	}
	return fmt.Sprintf(
		"vless://%s@%s:443?encryption=none&flow=xtls-rprx-vision&security=reality&sni=%s&fp=chrome&pbk=%s&sid=%s&type=tcp#%s-relay",
		uuid, relayIP, sniName, pbk, sid, name,
	)
}
