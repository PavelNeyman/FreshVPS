package nodes

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/PavelNeyman/netductor/internal/paths"
)

const uuidFile = "node_uuid"

// LocalStableID returns persistent node id (not hostname). Creates on first use.
func LocalStableID() string {
	path := filepath.Join(paths.EtcDir(), uuidFile)
	if b, err := os.ReadFile(path); err == nil {
		if s := strings.TrimSpace(string(b)); s != "" {
			return s
		}
	}
	var raw [16]byte
	_, _ = rand.Read(raw[:])
	// UUID-like: 8-4-4-4-12 hex without version nibble fuss
	h := hex.EncodeToString(raw[:])
	id := h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
	_ = os.MkdirAll(paths.EtcDir(), 0o755)
	_ = os.WriteFile(path, []byte(id+"\n"), 0o644)
	return id
}

// LocalHostname reads preferred hostname for this host.
func LocalHostname() string {
	if b, err := os.ReadFile(filepath.Join(paths.EtcDir(), "node_id")); err == nil {
		if s := strings.TrimSpace(string(b)); s != "" {
			return NormalizeHostname(s)
		}
	}
	if b, err := os.ReadFile("/etc/hostname"); err == nil {
		return NormalizeHostname(string(b))
	}
	return ""
}

// WriteLocalHostname updates node_id + OS hostname files (does not touch stable id).
func WriteLocalHostname(name string) {
	name = NormalizeHostname(name)
	if name == "" {
		return
	}
	_ = os.MkdirAll(paths.EtcDir(), 0o755)
	_ = os.WriteFile(filepath.Join(paths.EtcDir(), "node_id"), []byte(name+"\n"), 0o644)
	_ = os.WriteFile("/etc/hostname", []byte(name+"\n"), 0o644)
}
