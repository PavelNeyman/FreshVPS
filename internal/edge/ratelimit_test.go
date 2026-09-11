package edge

import (
	"os"
	"testing"
)

func TestAllowEnroll(t *testing.T) {
	dir := t.TempDir()
	_ = os.Setenv("NETDUCTOR_EDGE_DIR", dir)
	_ = os.Setenv("NETDUCTOR_ETC", dir)
	t.Cleanup(func() {
		_ = os.Unsetenv("NETDUCTOR_EDGE_DIR")
		_ = os.Unsetenv("NETDUCTOR_ETC")
	})
	ip := "203.0.113.9"
	for i := 0; i < 10; i++ {
		if !AllowEnroll(ip) {
			t.Fatalf("blocked too early at %d", i)
		}
	}
	if AllowEnroll(ip) {
		t.Fatal("should rate limit")
	}
	if !AllowEnroll("203.0.113.10") {
		t.Fatal("other ip should pass")
	}
}
