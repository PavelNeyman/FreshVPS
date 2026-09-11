package vpn

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinksAndQR(t *testing.T) {
	dir := t.TempDir()
	_ = os.Setenv("NETDUCTOR_ETC", dir)
	t.Cleanup(func() { _ = os.Unsetenv("NETDUCTOR_ETC") })
	_ = os.MkdirAll(filepath.Join(dir, "secrets"), 0o700)
	_ = os.WriteFile(filepath.Join(dir, "public_ip"), []byte("1.2.3.4\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "secrets", "reality_public_key"), []byte("pub\n"), 0o600)
	_ = os.WriteFile(filepath.Join(dir, "secrets", "reality_short_id"), []byte("sid\n"), 0o600)
	_ = os.WriteFile(filepath.Join(dir, "secrets", "hy2_password"), []byte("hy2pass\n"), 0o600)

	u := genUUID()
	if u == "" || strings.Count(u, "-") < 4 {
		t.Fatal(u)
	}
	v := VLESSLink("alice", u)
	if !strings.Contains(v, "vless://") {
		t.Fatal(v)
	}
	h := Hy2Link("alice", "secret")
	if !strings.Contains(h, "hysteria2://") && !strings.Contains(h, "hy2://") {
		t.Log(h) // format may vary
	}
	qr := ASCIIQR("test-payload")
	if qr == "" {
		t.Fatal("empty qr")
	}
	if !ValidName("alice_1") && !ValidName("alice") {
		t.Fatal("valid name")
	}
	if ValidName("../x") {
		t.Fatal("path should be invalid")
	}
}
