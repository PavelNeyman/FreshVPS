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
func InstallLampac() error {
	// Called only as component or from env hook at end of Run.
	if _, err := exec.LookPath("docker"); err != nil {
		if err := aptInstall("docker.io", "docker-cli"); err != nil {
			return fmt.Errorf("docker for lampac: %w", err)
		}
		_ = exec.Command("systemctl", "enable", "--now", "docker").Run()
	}
	_ = os.MkdirAll(filepath.Join(paths.OptDir(), "lampac"), 0o755)
	// stop existing
	_ = exec.Command("docker", "rm", "-f", "netductor-lampac").Run()
	img := env("NETDUCTOR_LAMPAC_IMAGE", "ghcr.io/lampac-nextgen/lampac:latest")
	port := env("NETDUCTOR_LAMPAC_PORT", "9118")
	cmd := exec.Command("docker", "run", "-d",
		"--name", "netductor-lampac",
		"--restart", "unless-stopped",
		"-p", port+":9118",
		"-v", filepath.Join(paths.OptDir(), "lampac")+":/home",
		img,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("lampac: %s %w", strings.TrimSpace(string(out)), err)
	}
	fmt.Fprintf(os.Stderr, "lampac listening :%s (docker %s)\n", port, img)
	return nil
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
