package edgeagent

import (
	"fmt"
	"strings"
)

// DesiredFromTemplate extracts uci set lines from template JSON-like map.
func DesiredUCI(tmpl map[string]any) []string {
	var lines []string
	if net, ok := tmpl["network"].(map[string]any); ok {
		if ip, ok := net["lan_ip"].(string); ok && ip != "" {
			lines = append(lines, "network.lan.ipaddr="+ip)
		}
		if mask, ok := net["lan_mask"].(string); ok && mask != "" {
			lines = append(lines, "network.lan.netmask="+mask)
		}
	}
	if wifi, ok := tmpl["wifi"].(map[string]any); ok {
		ssid, _ := wifi["ssid"].(string)
		key, _ := wifi["key"].(string)
		enc, _ := wifi["encryption"].(string)
		if enc == "" {
			enc = "psk2"
		}
		// OpenWrt default radios — best-effort common paths
		if ssid != "" {
			lines = append(lines,
				"wireless.default_radio0.ssid="+ssid,
				"wireless.default_radio0.encryption="+enc,
			)
			if key != "" {
				lines = append(lines, "wireless.default_radio0.key="+key)
			}
			lines = append(lines,
				"wireless.default_radio1.ssid="+ssid,
				"wireless.default_radio1.encryption="+enc,
			)
			if key != "" {
				lines = append(lines, "wireless.default_radio1.key="+key)
			}
		}
	}
	return lines
}

// DiffUCI returns sets needed when current values differ (current map path->value).
func DiffUCI(desired []string, current map[string]string) (toSet []string) {
	for _, line := range desired {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		if current[parts[0]] == parts[1] {
			continue
		}
		toSet = append(toSet, line)
	}
	return toSet
}

func FormatApplyReport(toSet []string) string {
	if len(toSet) == 0 {
		return "no changes"
	}
	return fmt.Sprintf("changing %d keys", len(toSet))
}
