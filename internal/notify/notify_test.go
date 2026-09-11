package notify

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTelegramMissingSecrets(t *testing.T) {
	dir := t.TempDir()
	_ = os.Setenv("NETDUCTOR_ETC", dir)
	t.Cleanup(func() { _ = os.Unsetenv("NETDUCTOR_ETC") })
	_ = os.MkdirAll(filepath.Join(dir, "secrets"), 0o700)
	err := Telegram("hi")
	if err == nil {
		t.Fatal("expected error without secrets")
	}
}
