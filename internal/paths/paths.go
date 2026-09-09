// Package paths resolves Netductor vs legacy FreshVPS filesystem locations.
package paths

import (
	"os"
	"path/filepath"
)

// FirstExisting returns the first path that exists, or the first candidate if none exist.
func FirstExisting(candidates ...string) string {
	for _, p := range candidates {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if len(candidates) > 0 {
		return candidates[0]
	}
	return ""
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

// EtcDir — /etc/netductor or /etc/freshvps
func EtcDir() string {
	return FirstExisting(
		env("NETDUCTOR_ETC", ""),
		"/etc/netductor",
		env("FRESHVPS_ETC", "/etc/freshvps"),
		"/etc/freshvps",
	)
}

func StateDir() string {
	return FirstExisting(
		env("NETDUCTOR_STATE", ""),
		"/var/lib/netductor",
		"/var/lib/freshvps",
	)
}

func OptDir() string {
	return FirstExisting(
		env("NETDUCTOR_ROOT", ""),
		"/opt/netductor",
		"/opt/freshvps",
	)
}

func SessionsDir() string {
	if v := os.Getenv("NETDUCTOR_SESSIONS"); v != "" {
		return v
	}
	return filepath.Join(EtcDir(), "sessions")
}

func ClientsDir() string {
	if v := os.Getenv("NETDUCTOR_CLIENTS"); v != "" {
		return v
	}
	return filepath.Join(EtcDir(), "clients")
}

func EdgeTokenFile() string {
	if v := os.Getenv("NETDUCTOR_EDGE_TOKEN_FILE"); v != "" {
		return v
	}
	return FirstExisting(
		filepath.Join(EtcDir(), "secrets", "edge_token"),
		"/etc/freshvps/secrets/edge_token",
	)
}

func EdgeDir() string {
	if v := os.Getenv("FRESHVPS_EDGE_DIR"); v != "" {
		return v
	}
	if v := os.Getenv("NETDUCTOR_EDGE_DIR"); v != "" {
		return v
	}
	return filepath.Join(StateDir(), "edge")
}

func MetricsDir() string {
	if v := os.Getenv("FRESHVPS_METRICS_DIR"); v != "" {
		return v
	}
	if v := os.Getenv("NETDUCTOR_METRICS_DIR"); v != "" {
		return v
	}
	return filepath.Join(StateDir(), "metrics")
}

func ProbesCfg() string {
	if v := os.Getenv("NETDUCTOR_PROBES_CFG"); v != "" {
		return v
	}
	return filepath.Join(EtcDir(), "probes.json")
}

func AdminRoot() string {
	if v := os.Getenv("NETDUCTOR_ADMIN_ROOT"); v != "" {
		return v
	}
	if v := os.Getenv("FRESHVPS_ADMIN_ROOT"); v != "" {
		return v
	}
	return FirstExisting(
		filepath.Join(OptDir(), "runtime", "api", "admin"),
		"/opt/netductor/runtime/api/admin",
		"/opt/freshvps/runtime/api/admin",
		"runtime/api/admin",
	)
}

func VPNBin() string {
	if v := os.Getenv("NETDUCTOR_VPN_BIN"); v != "" {
		return v
	}
	return FirstExisting(
		"/usr/local/bin/freshvps-vpn",
		"/usr/local/bin/netductor-vpn",
		filepath.Join(OptDir(), "bin", "freshvps-vpn"),
	)
}
