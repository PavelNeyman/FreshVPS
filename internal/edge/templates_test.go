package edge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTemplateBind(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("NETDUCTOR_EDGE_DIR", dir)
	os.Setenv("NETDUCTOR_ETC", dir)
	defer os.Unsetenv("NETDUCTOR_EDGE_DIR")
	defer os.Unsetenv("NETDUCTOR_ETC")
	_ = os.MkdirAll(filepath.Join(dir, "secrets"), 0o700)
	EnsureDefaultTemplate()
	Enroll(map[string]any{"device_id": "s1"})
	_, _ = Approve("s1")
	if err := BindTemplate("s1", "default", map[string]any{"ssid": "SiteA", "lan_ip": "10.0.0.1"}); err != nil {
		t.Fatal(err)
	}
	tmpl, err := TemplateForDevice("s1")
	if err != nil {
		t.Fatal(err)
	}
	net, _ := tmpl["network"].(map[string]any)
	if net["lan_ip"] != "10.0.0.1" {
		t.Fatalf("%v", net)
	}
	wifi, _ := tmpl["wifi"].(map[string]any)
	if wifi["ssid"] != "SiteA" {
		t.Fatalf("%v", wifi)
	}
}
