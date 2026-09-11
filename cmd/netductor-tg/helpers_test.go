package main

import "testing"

func TestEsc(t *testing.T) {
	s := esc(`<b>&"hi"`)
	if s == `<b>&"hi"` {
		t.Fatal("should escape")
	}
	if !containsAll(s, "&lt;", "&amp;") {
		t.Fatal(s)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !contains(s, p) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestFormatVPNList(t *testing.T) {
	in := "alice\ton\tuuid1\tnote\t2026\nbob\toff\tuuid2\t\t2026\n"
	out := formatVPNList(in)
	if out == "" {
		t.Fatal()
	}
}
