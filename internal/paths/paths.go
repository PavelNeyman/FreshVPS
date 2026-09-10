package paths

import (
	"os"
	"path/filepath"
)

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

// Prefer Netductor paths; legacy FreshVPS only if already present and netductor missing.

func EtcDir() string {
	if v := os.Getenv("NETDUCTOR_ETC"); v != "" {
		return v
	}
	if _, err := os.Stat("/etc/netductor"); err == nil {
		return "/etc/netductor"
	}
	if _, err := os.Stat("/etc/freshvps"); err == nil {
		return "/etc/freshvps"
	}
	return "/etc/netductor"
}

func StateDir() string {
	if v := os.Getenv("NETDUCTOR_STATE"); v != "" {
		return v
	}
	if _, err := os.Stat("/var/lib/netductor"); err == nil {
		return "/var/lib/netductor"
	}
	if _, err := os.Stat("/var/lib/freshvps"); err == nil {
		return "/var/lib/freshvps"
	}
	return "/var/lib/netductor"
}

func OptDir() string {
	if v := os.Getenv("NETDUCTOR_ROOT"); v != "" {
		return v
	}
	if _, err := os.Stat("/opt/netductor"); err == nil {
		return "/opt/netductor"
	}
	if _, err := os.Stat("/opt/freshvps"); err == nil {
		return "/opt/freshvps"
	}
	return "/opt/netductor"
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
	return filepath.Join(EtcDir(), "secrets", "edge_token")
}

func EdgeDir() string {
	if v := os.Getenv("NETDUCTOR_EDGE_DIR"); v != "" {
		return v
	}
	if v := os.Getenv("FRESHVPS_EDGE_DIR"); v != "" {
		return v
	}
	return filepath.Join(StateDir(), "edge")
}

func MetricsDir() string {
	if v := os.Getenv("NETDUCTOR_METRICS_DIR"); v != "" {
		return v
	}
	if v := os.Getenv("FRESHVPS_METRICS_DIR"); v != "" {
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
		"/usr/local/bin/netductor", // vpn subcommand preferred
		"/usr/local/bin/freshvps-vpn",
		filepath.Join(OptDir(), "bin", "netductor"),
	)
}

func EnsureLayout() error {
	for _, d := range []string{
		EtcDir(),
		filepath.Join(EtcDir(), "secrets"),
		filepath.Join(EtcDir(), "sessions"),
		filepath.Join(EtcDir(), "clients"),
		StateDir(),
		filepath.Join(StateDir(), "metrics"),
		filepath.Join(StateDir(), "edge"),
		OptDir(),
		filepath.Join(OptDir(), "bin"),
		filepath.Join(OptDir(), "runtime", "api", "admin"),
		filepath.Join(OptDir(), "runtime", "telegram"),
	} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	_ = os.Chmod(filepath.Join(EtcDir(), "secrets"), 0o700)
	_ = os.Chmod(filepath.Join(EtcDir(), "sessions"), 0o700)
	_ = os.Chmod(filepath.Join(EtcDir(), "clients"), 0o700)
	return nil
}
