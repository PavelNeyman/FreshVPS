package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteReadSecret(t *testing.T) {
	dir := t.TempDir()
	_ = os.Setenv("NETDUCTOR_ETC", dir)
	t.Cleanup(func() { _ = os.Unsetenv("NETDUCTOR_ETC") })
	if err := writeSecret("k", "v123"); err != nil {
		t.Fatal(err)
	}
	if readSecret("k") != "v123" {
		t.Fatal(readSecret("k"))
	}
	st, err := os.Stat(filepath.Join(dir, "secrets", "k"))
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm()&0o077 != 0 {
		// may be 0600
		if st.Mode().Perm() != 0o600 && st.Mode().Perm() != 0o640 {
			t.Logf("perm %o", st.Mode().Perm())
		}
	}
}

func TestDefaultComponents(t *testing.T) {
	c := DefaultComponents()
	if len(c) < 3 {
		t.Fatal(c)
	}
}

func TestArch(t *testing.T) {
	a := arch()
	if a == "" {
		t.Fatal()
	}
}
