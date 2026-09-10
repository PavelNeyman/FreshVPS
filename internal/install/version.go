package install

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/PavelNeyman/netductor/internal/paths"
)

func stateVersion(name string) string {
	b, err := os.ReadFile(filepath.Join(paths.StateDir(), name+"-version"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func writeStateVersion(name, ver string) {
	_ = os.MkdirAll(paths.StateDir(), 0o755)
	_ = os.WriteFile(filepath.Join(paths.StateDir(), name+"-version"), []byte(ver+"\n"), 0o644)
}
