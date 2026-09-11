package metrics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestHistoryAndUptime(t *testing.T) {
	dir := t.TempDir()
	_ = os.Setenv("NETDUCTOR_STATE", dir)
	t.Cleanup(func() { _ = os.Unsetenv("NETDUCTOR_STATE") })
	_ = os.MkdirAll(Dir(), 0o755)

	row := map[string]any{
		"ts": 1,
		"probes": []any{
			map[string]any{"name": "a", "ok": true},
			map[string]any{"name": "b", "ok": false},
		},
	}
	b, _ := json.Marshal(row)
	path := filepath.Join(Dir(), "history.jsonl")
	_ = os.WriteFile(path, append(b, '\n'), 0o644)
	b2, _ := json.Marshal(map[string]any{
		"ts": 2,
		"probes": []any{
			map[string]any{"name": "a", "ok": true},
			map[string]any{"name": "b", "ok": true},
		},
	})
	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	_, _ = f.Write(append(b2, '\n'))
	_ = f.Close()

	h := History(10)
	if len(h) != 2 {
		t.Fatalf("%d", len(h))
	}
	u := ProbeUptime(10)
	a, ok := u["a"].(map[string]any)
	if !ok {
		t.Fatal(u)
	}
	if a["ok"].(int) != 2 {
		t.Fatalf("%v", a)
	}
	lp := LatestProbes()
	if len(lp) != 2 {
		t.Fatal(lp)
	}
}

func TestCollectShape(t *testing.T) {
	dir := t.TempDir()
	_ = os.Setenv("NETDUCTOR_STATE", dir)
	t.Cleanup(func() { _ = os.Unsetenv("NETDUCTOR_STATE") })
	m := Collect()
	if m["ts"] == nil {
		t.Fatal(m)
	}
}
