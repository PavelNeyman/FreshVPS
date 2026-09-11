package edge

import (
	"sync"
	"time"
)

type attempt struct {
	n    int
	until time.Time
}

var (
	rlMu   sync.Mutex
	rlMap  = map[string]attempt{}
	// production defaults
	enrollMaxPerWindow = 10
	enrollWindow       = 15 * time.Minute
)

// AllowEnroll returns false if IP exceeded enroll rate.
func AllowEnroll(ip string) bool {
	if ip == "" {
		ip = "unknown"
	}
	rlMu.Lock()
	defer rlMu.Unlock()
	now := time.Now()
	a, ok := rlMap[ip]
	if !ok || now.After(a.until) {
		rlMap[ip] = attempt{n: 1, until: now.Add(enrollWindow)}
		return true
	}
	if a.n >= enrollMaxPerWindow {
		return false
	}
	a.n++
	rlMap[ip] = a
	return true
}
