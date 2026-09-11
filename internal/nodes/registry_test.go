package nodes

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeHostname(t *testing.T) {
	if NormalizeHostname("ND Core 01!") != "nd-core-01" {
		t.Fatalf("got %q", NormalizeHostname("ND Core 01!"))
	}
}

func TestStableIDRename(t *testing.T) {
	dir := t.TempDir()
	_ = os.Setenv("NETDUCTOR_STATE", dir)
	_ = os.Setenv("NETDUCTOR_ETC", filepath.Join(dir, "etc"))
	t.Cleanup(func() {
		_ = os.Unsetenv("NETDUCTOR_STATE")
		_ = os.Unsetenv("NETDUCTOR_ETC")
	})
	_ = os.MkdirAll(filepath.Join(dir, "etc"), 0o755)

	id := LocalStableID()
	if id == "" {
		t.Fatal("empty id")
	}
	if LocalStableID() != id {
		t.Fatal("id must be stable")
	}
	_ = SelfRegisterLocal("nd-core-1", "core", "1.2.3.4")
	_, err := SetDesiredHostname(id, "nd-core-nl01")
	if err != nil {
		t.Fatal(err)
	}
	WriteLocalHostname("nd-core-nl01")
	if err := SyncLocalHostname(); err != nil {
		t.Fatal(err)
	}
	list, _ := List()
	if len(list) != 1 {
		t.Fatalf("%+v", list)
	}
	if list[0].ID != id || list[0].Hostname != "nd-core-nl01" {
		t.Fatalf("%+v", list[0])
	}
}
