package edge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTemplateCRUD(t *testing.T) {
	dir := t.TempDir()
	_ = os.Setenv("NETDUCTOR_EDGE_DIR", dir)
	_ = os.Setenv("NETDUCTOR_ETC", dir)
	t.Cleanup(func() {
		_ = os.Unsetenv("NETDUCTOR_EDGE_DIR")
		_ = os.Unsetenv("NETDUCTOR_ETC")
	})
	_ = os.MkdirAll(filepath.Join(dir, "secrets"), 0o700)

	if sanitizeID("../x y") == "../x y" {
		t.Fatal("sanitize failed")
	}
	EnsureDefaultTemplate()
	list := ListTemplates()
	if len(list) < 1 {
		t.Fatal("expected default")
	}
	tmpl := Template{"id": "site", "network": map[string]any{"lan_ip": "192.168.50.1"}}
	if err := SaveTemplate("site", tmpl); err != nil {
		t.Fatal(err)
	}
	got, err := GetTemplate("site")
	if err != nil {
		t.Fatal(err)
	}
	net, _ := got["network"].(map[string]any)
	if net["lan_ip"] != "192.168.50.1" {
		t.Fatal(got)
	}
	if err := DeleteTemplate("site"); err != nil {
		t.Fatal(err)
	}
	if _, err := GetTemplate("site"); err == nil {
		t.Fatal("deleted")
	}
}
