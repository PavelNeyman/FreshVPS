package edge

import (
	"net"
	"time"
)

// CheckRelayTCP probes relay:443. Used by OpenWrt agent for WAN-fallback decisions.
func CheckRelayTCP(host string, timeout time.Duration) bool {
	if host == "" {
		return false
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	c, err := net.DialTimeout("tcp", net.JoinHostPort(host, "443"), timeout)
	if err != nil {
		return false
	}
	_ = c.Close()
	return true
}
