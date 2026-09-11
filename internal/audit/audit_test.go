package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLog(t *testing.T) {
	dir := t.TempDir()
	_ = os.Setenv("NETDUCTOR_STATE", dir)
	t.Cleanup(func() { _ = os.Unsetenv("NETDUCTOR_STATE") })

	Log("op", "vpn.add", "alice", "note")
	Log("op", "edge.approve", "dev1", "")
	b, err := os.ReadFile(filepath.Join(dir, "audit.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	if len(lines) != 2 {
		t.Fatalf("%q", string(b))
	}
	if !strings.Contains(lines[0], `"action":"vpn.add"`) {
		t.Fatal(lines[0])
	}
}
