package edgeagent

import "strings"

// ParseUCIBatch returns operations: set path=value, or special commit/network_reload/wifi_reload
func ParseUCIBatch(arg string) (sets [][2]string, specials []string, bad []string) {
	for _, line := range strings.Split(arg, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		switch line {
		case "commit", "network_reload", "wifi_reload":
			specials = append(specials, line)
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			bad = append(bad, line)
			continue
		}
		sets = append(sets, [2]string{parts[0], parts[1]})
	}
	return sets, specials, bad
}
