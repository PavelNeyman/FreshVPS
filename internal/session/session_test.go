package session

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateValidRevoke(t *testing.T) {
	dir := t.TempDir()
	_ = os.Setenv("NETDUCTOR_ETC", dir)
	t.Cleanup(func() { _ = os.Unsetenv("NETDUCTOR_ETC") })
	_ = os.MkdirAll(filepath.Join(dir, "sessions"), 0o700)

	tok, exp, err := Create(1, "test", "127.0.0.1")
	if err != nil || tok == "" || exp == 0 {
		t.Fatal(err, tok, exp)
	}
	if !Valid(tok) {
		t.Fatal("should be valid")
	}
	// hash stored, not plaintext name
	ents, _ := os.ReadDir(Dir())
	for _, e := range ents {
		if e.Name() == tok {
			t.Fatal("token must not be filename")
		}
	}
	Revoke(tok)
	if Valid(tok) {
		t.Fatal("revoked")
	}
}

func TestClampHours(t *testing.T) {
	if clampHours(0) != DefaultHours {
		t.Fatal()
	}
	if clampHours(999) != MaxHours {
		t.Fatal()
	}
}
