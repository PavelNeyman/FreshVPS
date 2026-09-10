package edgeagent

import "strings"

// SplitArg parses URL|sha256|rest
func SplitArg(arg string) (url, sha, rest string) {
	parts := strings.Split(arg, "|")
	if len(parts) == 0 {
		return "", "", ""
	}
	url = strings.TrimSpace(parts[0])
	if len(parts) > 1 {
		sha = strings.TrimSpace(parts[1])
	}
	if len(parts) > 2 {
		rest = strings.Join(parts[2:], "|")
	}
	return url, sha, rest
}

func SysupgradeAllowed(arg string) bool {
	return strings.Contains(arg, "confirm=yes")
}
