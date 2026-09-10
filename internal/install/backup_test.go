package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "a.tar.gz")
	enc := filepath.Join(dir, "a.ndenc")
	out := filepath.Join(dir, "b.tar.gz")
	_ = os.WriteFile(in, []byte("netductor-backup-payload"), 0o600)
	if err := encryptFile(in, enc, "test-secret-key"); err != nil {
		t.Fatal(err)
	}
	if err := decryptFile(enc, out, "test-secret-key"); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(out)
	if string(b) != "netductor-backup-payload" {
		t.Fatalf("%q", b)
	}
	if decryptFile(enc, out, "wrong") == nil {
		t.Fatal("expected fail")
	}
}
