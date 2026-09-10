package session

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestValidToken(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("NETDUCTOR_ETC", dir)
	defer os.Unsetenv("NETDUCTOR_ETC")
	_ = os.MkdirAll(filepath.Join(dir, "sessions"), 0o700)
	tok := "abcd1234deadbeef"
	exp := time.Now().Add(time.Hour).Unix()
	_ = os.WriteFile(filepath.Join(dir, "sessions", tok), []byte(fmt.Sprintf("%d\n", exp)), 0o600)
	if !Valid(tok) {
		t.Fatal("expected valid")
	}
	if Valid("missing") {
		t.Fatal("expected invalid")
	}
	_ = os.WriteFile(filepath.Join(dir, "sessions", "old"), []byte("1\n"), 0o600)
	if Valid("old") {
		t.Fatal("expired should be invalid")
	}
}

func TestTokenFromAuth(t *testing.T) {
	if TokenFromAuth("Bearer xyz", "") != "xyz" {
		t.Fatal()
	}
	if TokenFromAuth("", "nd_session=abc; Path=/") != "abc" {
		t.Fatal(TokenFromAuth("", "nd_session=abc; Path=/"))
	}
}
