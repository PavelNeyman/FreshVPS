package edgeagent

import "testing"

func TestParseUCIBatch(t *testing.T) {
	sets, sp, bad := ParseUCIBatch("network.lan.ipaddr=1.2.3.4\n#c\ncommit\nbadline\nwifi_reload\n")
	if len(sets) != 1 || sets[0][0] != "network.lan.ipaddr" {
		t.Fatalf("%v", sets)
	}
	if len(sp) != 2 {
		t.Fatalf("%v", sp)
	}
	if len(bad) != 1 {
		t.Fatalf("%v", bad)
	}
}
