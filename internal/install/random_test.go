package install

import (
	"testing"
)

func TestRandomHex(t *testing.T) {
	s := randomHex(24)
	if len(s) != 24 {
		t.Fatalf("len %d", len(s))
	}
	s2 := randomHex(24)
	if s == s2 {
		t.Fatal("not random")
	}
}
