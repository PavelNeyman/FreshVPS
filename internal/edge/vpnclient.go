package edge

import (
	"fmt"
	"strings"

	"github.com/PavelNeyman/netductor/internal/vpn"
)

func edgeVPNUser(deviceID string) string {
	id := sanitizeID(deviceID)
	if id == "" {
		id = "device"
	}
	return "edge-" + id
}

// EnsureVPNClient creates VPN user for device and returns subscription links.
func EnsureVPNClient(deviceID string) (map[string]string, error) {
	name := edgeVPNUser(deviceID)
	users, _ := vpn.ListNative()
	found := false
	for _, u := range users {
		if u.Name == name {
			found = true
			break
		}
	}
	if !found {
		if _, err := vpn.AddNative(name, "edge device "+deviceID); err != nil {
			// may fail apply without sing-box — still try links
			_ = err
		}
	}
	sub, _ := vpn.ReadClient(name, "subscription.txt", "link.txt")
	vless, _ := vpn.ReadClient(name, "link-vless.txt", "link.txt")
	hy2, _ := vpn.ReadClient(name, "link-hy2.txt")
	if sub == "" && vless == "" {
		return nil, fmt.Errorf("no vpn links for %s — is sing-box installed?", name)
	}
	return map[string]string{
		"user":         name,
		"subscription": sub,
		"vless":        vless,
		"hy2":          hy2,
	}, nil
}

func TemplateWithVPN(deviceID string) (Template, error) {
	t, err := TemplateForDevice(deviceID)
	if err != nil {
		return nil, err
	}
	vpnSec, _ := t["vpn"].(map[string]any)
	enabled := false
	if vpnSec != nil {
		switch v := vpnSec["enabled"].(type) {
		case bool:
			enabled = v
		case string:
			enabled = v == "true" || v == "1"
		}
	}
	if !enabled {
		return t, nil
	}
	links, err := EnsureVPNClient(deviceID)
	if err != nil {
		t["vpn_error"] = err.Error()
		return t, nil
	}
	if vpnSec == nil {
		vpnSec = map[string]any{}
	}
	for k, v := range links {
		vpnSec[k] = v
	}
	t["vpn"] = vpnSec
	return t, nil
}

// used only to avoid unused strings import if sanitize only
var _ = strings.TrimSpace
