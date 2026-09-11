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

func TestUpsertDesiredAndPrune(t *testing.T) {
	dir := t.TempDir()
	_ = os.Setenv("NETDUCTOR_STATE", dir)
	_ = os.Setenv("NETDUCTOR_ETC", filepath.Join(dir, "etc"))
	t.Cleanup(func() {
		_ = os.Unsetenv("NETDUCTOR_STATE")
		_ = os.Unsetenv("NETDUCTOR_ETC")
	})
	_ = os.MkdirAll(filepath.Join(dir, "etc"), 0o755)

	_, err := UpsertFromDevice(Node{ID: "nd-core-1", Hostname: "nd-core-1", Role: "core", Kind: "vps", PublicIP: "1.2.3.4"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = SetDesiredHostname("nd-core-1", "nd-core-nl01")
	if err != nil {
		t.Fatal(err)
	}
	// simulate applied
	_, err = UpsertFromDevice(Node{ID: "nd-core-nl01", Hostname: "nd-core-nl01", Role: "core", Kind: "vps", PublicIP: "1.2.3.4"})
	if err != nil {
		t.Fatal(err)
	}
	_ = Delete("nd-core-1")
	_ = pruneStaleVPS("nd-core-nl01", "1.2.3.4")
	list, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != "nd-core-nl01" {
		t.Fatalf("%+v", list)
	}
}
