package edgeagent

import (
	"strings"
	"testing"
)

func TestVLESSClientConfig(t *testing.T) {
	link := "vless://11111111-2222-3333-4444-555555555555@1.2.3.4:443?encryption=none&flow=xtls-rprx-vision&security=reality&sni=www.cloudflare.com&fp=chrome&pbk=PUB&sid=abcd&type=tcp#t"
	b, err := VLESSClientConfig(link)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "1.2.3.4") || !strings.Contains(s, "PUB") || !strings.Contains(s, "reality") {
		t.Fatal(s)
	}
}
