package paths

import (
	"os"
	"testing"
)

func TestEnvOverride(t *testing.T) {
	os.Setenv("NETDUCTOR_ETC", "/tmp/nd-test-etc")
	defer os.Unsetenv("NETDUCTOR_ETC")
	if EtcDir() != "/tmp/nd-test-etc" {
		t.Fatal(EtcDir())
	}
}
