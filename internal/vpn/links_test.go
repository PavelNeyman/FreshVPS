package vpn

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVLESSLinkFormat(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("NETDUCTOR_ETC", dir)
	os.Setenv("PUBLIC_IP", "1.2.3.4")
	defer os.Unsetenv("NETDUCTOR_ETC")
	defer os.Unsetenv("PUBLIC_IP")
	_ = os.MkdirAll(filepath.Join(dir, "secrets"), 0o700)
	_ = os.WriteFile(filepath.Join(dir, "secrets", "singbox_reality_public"), []byte("PUBKEY\n"), 0o600)
	_ = os.WriteFile(filepath.Join(dir, "secrets", "singbox_short_id"), []byte("abcd\n"), 0o600)
	link := VLESSLink("alice", "11111111-2222-3333-4444-555555555555")
	if !strings.HasPrefix(link, "vless://") || !strings.Contains(link, "1.2.3.4") || !strings.Contains(link, "PUBKEY") {
		t.Fatal(link)
	}
	h := Hy2Link("alice", "secretpass")
	if !strings.HasPrefix(h, "hysteria2://") {
		t.Fatal(h)
	}
}
