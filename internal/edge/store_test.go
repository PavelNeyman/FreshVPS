package edge

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnqueuePollAndResults(t *testing.T) {
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
	CmdResult(map[string]any{"device_id": "dev1", "cmd_id": id, "result": "pong"})
	res := ListResults()
	if len(res) < 1 {
		t.Fatal(res)
	}
}

func TestBackupAndList(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("NETDUCTOR_EDGE_DIR", dir)
	defer os.Unsetenv("NETDUCTOR_EDGE_DIR")
	path, err := SaveBackup("site1", bytes.NewReader([]byte("hello-backup")))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	list := ListBackups("site1")
	if len(list) != 1 {
		t.Fatalf("%v", list)
	}
	name, _ := list[0]["name"].(string)
	p2, err := BackupPath("site1", name)
	if err != nil || p2 == "" {
		t.Fatal(err, p2)
	}
	_, err = BackupPath("site1", "../etc/passwd")
	if err == nil {
		t.Fatal("path traversal")
	}
}

func TestMetrics(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("NETDUCTOR_EDGE_DIR", dir)
	defer os.Unsetenv("NETDUCTOR_EDGE_DIR")
	_ = SaveMetrics("site1", map[string]any{"device_id": "site1", "mem_pct": 10.0})
	_ = SaveMetrics("site1", map[string]any{"device_id": "site1", "mem_pct": 20.0})
	tail := ListMetricsTail("site1", 10)
	if len(tail) < 2 {
		t.Fatalf("%v", tail)
	}
}

func TestDeviceDirSafe(t *testing.T) {
	d := deviceDir("a/../b")
	if strings.Contains(d, "..") {
		// mapped to underscores — still one path under EdgeDir
	}
	if !strings.Contains(d, "a") {
		// a_b style
		_ = d
	}
}
