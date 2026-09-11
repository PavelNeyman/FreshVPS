package edge

import (
	"os"
	"testing"
)

func TestRSCSaveRead(t *testing.T) {
	dir := t.TempDir()
	_ = os.Setenv("NETDUCTOR_EDGE_DIR", dir)
	t.Cleanup(func() { _ = os.Unsetenv("NETDUCTOR_EDGE_DIR") })
	name, err := SaveRSC("mt-home", "/system identity set name=x\n")
	if err != nil || name == "" {
		t.Fatal(err, name)
	}
	b, err := ReadRSC(name)
	if err != nil || string(b) == "" {
		t.Fatal(err, b)
	}
	if _, err := ReadRSC("../etc/passwd"); err == nil {
		t.Fatal("traversal")
	}
	if sanitize("a!b@c") != "abc" && sanitize("a!b@c") == "a!b@c" {
		t.Fatal(sanitize("a!b@c"))
	}
}
