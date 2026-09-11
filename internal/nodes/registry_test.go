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

func TestUpsertDesired(t *testing.T) {
	dir := t.TempDir()
	_ = os.Setenv("NETDUCTOR_STATE", dir)
	t.Cleanup(func() { _ = os.Unsetenv("NETDUCTOR_STATE") })

	n, err := UpsertFromDevice(Node{ID: "nd-core-1", Hostname: "nd-core-1", Role: "core", Kind: "vps"})
	if err != nil {
		t.Fatal(err)
	}
	if n.Hostname != "nd-core-1" {
		t.Fatal(n)
	}
	n2, err := SetDesiredHostname("nd-core-1", "nd-core-nl01")
	if err != nil {
		t.Fatal(err)
	}
	if n2.DesiredHN != "nd-core-nl01" {
		t.Fatal(n2)
	}
	// device applies
	n3, err := UpsertFromDevice(Node{ID: "nd-core-1", Hostname: "nd-core-nl01", Role: "core", Kind: "vps"})
	if err != nil {
		t.Fatal(err)
	}
	if n3.DesiredHN != "" || n3.Hostname != "nd-core-nl01" {
		t.Fatalf("%+v", n3)
	}
	_ = filepath.Join(dir, "nodes")
}
