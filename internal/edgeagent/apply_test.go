package edgeagent

import "testing"

func TestDesiredAndDiff(t *testing.T) {
	tmpl := map[string]any{
		"network": map[string]any{"lan_ip": "192.168.50.1", "lan_mask": "255.255.255.0"},
		"wifi":    map[string]any{"ssid": "Test", "key": "secret123", "encryption": "psk2"},
	}
	d := DesiredUCI(tmpl)
	if len(d) < 3 {
		t.Fatal(d)
	}
	cur := map[string]string{"network.lan.ipaddr": "192.168.50.1"}
	diff := DiffUCI(d, cur)
	if len(diff) == 0 {
		t.Fatal("expected some diffs")
	}
	// all match
	cur2 := map[string]string{}
	for _, line := range d {
		p := splitKV(line)
		cur2[p[0]] = p[1]
	}
	if len(DiffUCI(d, cur2)) != 0 {
		t.Fatal("expected no changes")
	}
}

func splitKV(line string) [2]string {
	i := 0
	for i < len(line) && line[i] != '=' {
		i++
	}
	return [2]string{line[:i], line[i+1:]}
}
