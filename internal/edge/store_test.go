package edge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnqueueAndPoll(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("NETDUCTOR_EDGE_DIR", dir)
	os.Setenv("NETDUCTOR_ETC", dir)
	defer os.Unsetenv("NETDUCTOR_EDGE_DIR")
	defer os.Unsetenv("NETDUCTOR_ETC")
	_ = os.MkdirAll(filepath.Join(dir, "secrets"), 0o700)
	id := EnqueueCmd("dev1", "ping", "")
	if id == "" {
		t.Fatal("empty id")
	}
	cmds := PollCommands("dev1")
	if len(cmds) < 1 {
		t.Fatalf("expected command, got %v", cmds)
	}
}
