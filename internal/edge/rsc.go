package edge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PavelNeyman/netductor/internal/paths"
)

func rscDir() string {
	return filepath.Join(paths.EdgeDir(), "rsc")
}

func SaveRSC(deviceID, script string) (string, error) {
	_ = os.MkdirAll(rscDir(), 0o700)
	name := fmt.Sprintf("%s-%d.rsc", sanitize(deviceID), time.Now().Unix())
	path := filepath.Join(rscDir(), name)
	if err := os.WriteFile(path, []byte(script), 0o600); err != nil {
		return "", err
	}
	return name, nil
}

func ReadRSC(name string) ([]byte, error) {
	name = filepath.Base(name)
	if name == "." || name == ".." || strings.Contains(name, "/") {
		return nil, fmt.Errorf("bad name")
	}
	return os.ReadFile(filepath.Join(rscDir(), name))
}

func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "dev"
	}
	return b.String()
}
