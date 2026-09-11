package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestParseKV(t *testing.T) {
	var c config
	parseKV(`
# comment
SERVER=https://example.com
TOKEN=secret
DEVICE_ID=ow-1
INTERVAL=30
BADLINE
INTERVAL=nope
`, &c)
	if c.Server != "https://example.com" || c.Token != "secret" || c.DeviceID != "ow-1" || c.Interval != 30 {
		t.Fatalf("%+v", c)
	}
}

func TestLoadConfigEnv(t *testing.T) {
	dir := t.TempDir()
	conf := filepath.Join(dir, "cfg")
	_ = os.WriteFile(conf, []byte("SERVER=http://127.0.0.1:8787\nTOKEN=tok\n"), 0o600)
	_ = os.Setenv("NETDUCTOR_AGENT_CONF", conf)
	_ = os.Setenv("NETDUCTOR_DEVICE_ID", "dev-env")
	_ = os.Setenv("NETDUCTOR_INTERVAL", "15")
	t.Cleanup(func() {
		_ = os.Unsetenv("NETDUCTOR_AGENT_CONF")
		_ = os.Unsetenv("NETDUCTOR_DEVICE_ID")
		_ = os.Unsetenv("NETDUCTOR_INTERVAL")
	})
	c := loadConfig()
	if c.Server != "http://127.0.0.1:8787" || c.Token != "tok" || c.DeviceID != "dev-env" || c.Interval != 15 {
		t.Fatalf("%+v", c)
	}
}

func TestTruncate(t *testing.T) {
	if truncate("  hi  ", 10) != "hi" {
		t.Fatal()
	}
	s := truncate("abcdefghijKLM", 5)
	if s != "abcde…" {
		t.Fatal(s)
	}
}

func TestFileSHA256(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f")
	_ = os.WriteFile(p, []byte("abc"), 0o644)
	got, err := fileSHA256(p)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte("abc"))
	if got != hex.EncodeToString(sum[:]) {
		t.Fatal(got)
	}
}

func TestDoJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer t" {
			w.WriteHeader(401)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer srv.Close()
	b, err := doJSON(srv.Client(), http.MethodGet, srv.URL, "t", nil)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	if m["ok"] != true {
		t.Fatal(m)
	}
	_, err = doJSON(srv.Client(), http.MethodGet, srv.URL, "bad", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCollectMetricsShape(t *testing.T) {
	m := collectMetrics()
	if m == nil {
		t.Fatal()
	}
}
