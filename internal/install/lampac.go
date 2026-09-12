package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/PavelNeyman/netductor/internal/paths"
)

// InstallLampac pulls optional media aggregator (Docker). Off by default.
// Network: bound to 127.0.0.1 only — not on public WAN. Reach via VPN
// (proxy mode → http://127.0.0.1:9118 on server) or SSH -L 9118:127.0.0.1:9118.
func InstallLampac() error {
	if _, err := exec.LookPath("docker"); err != nil {
		if err := aptInstall("docker.io", "docker-cli"); err != nil {
			return fmt.Errorf("docker for lampac: %w", err)
		}
		_ = exec.Command("systemctl", "enable", "--now", "docker").Run()
	}
	_ = os.MkdirAll(filepath.Join(paths.OptDir(), "lampac"), 0o755)
	_ = exec.Command("docker", "rm", "-f", "netductor-lampac").Run()

	img := env("NETDUCTOR_LAMPAC_IMAGE", "ghcr.io/lampac-nextgen/lampac:latest")
	port := env("NETDUCTOR_LAMPAC_PORT", "9118")
	// Default: localhost only. Override with NETDUCTOR_LAMPAC_BIND=0.0.0.0 (not recommended).
	bind := env("NETDUCTOR_LAMPAC_BIND", "127.0.0.1")
	publish := bind + ":" + port + ":9118"

	cmd := exec.Command("docker", "run", "-d",
		"--name", "netductor-lampac",
		"--restart", "unless-stopped",
		"-p", publish,
		"-v", filepath.Join(paths.OptDir(), "lampac")+":/home",
		img,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("lampac: %s %w", strings.TrimSpace(string(out)), err)
	}

	lockLampacToLocalhost(port)
	fmt.Fprintf(os.Stderr, "lampac on %s:%s only (not public WAN); image %s\n", bind, port, img)
	return nil
}

// lockLampacToLocalhost drops direct WAN hits to lampac port (defense in depth).
func lockLampacToLocalhost(port string) {
	// Prefer ufw if active
	if out, err := exec.Command("ufw", "status").Output(); err == nil && strings.Contains(string(out), "active") {
		_ = exec.Command("ufw", "deny", port+"/tcp").Run()
		_ = exec.Command("ufw", "allow", "from", "127.0.0.1", "to", "any", "port", port, "proto", "tcp").Run()
	}
	// iptables: accept lo, drop the rest for this port (idempotent insert)
	_ = exec.Command("iptables", "-C", "INPUT", "-i", "lo", "-p", "tcp", "--dport", port, "-j", "ACCEPT").Run()
	if err := exec.Command("iptables", "-C", "INPUT", "-i", "lo", "-p", "tcp", "--dport", port, "-j", "ACCEPT").Run(); err != nil {
		_ = exec.Command("iptables", "-I", "INPUT", "1", "-i", "lo", "-p", "tcp", "--dport", port, "-j", "ACCEPT").Run()
	}
	if err := exec.Command("iptables", "-C", "INPUT", "-p", "tcp", "--dport", port, "-j", "DROP").Run(); err != nil {
		_ = exec.Command("iptables", "-A", "INPUT", "-p", "tcp", "--dport", port, "-j", "DROP").Run()
	}
}

var currentComps []string

func containsComp(name string) bool {
	for _, c := range currentComps {
		if c == name {
			return true
		}
	}
	return false
}
