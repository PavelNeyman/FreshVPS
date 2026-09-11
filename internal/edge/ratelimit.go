package edge

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/PavelNeyman/netductor/internal/paths"
)

type attempt struct {
	N     int   `json:"n"`
	Until int64 `json:"until"`
}

var (
	rlMu               sync.Mutex
	enrollMaxPerWindow = 10
	enrollWindow       = 15 * time.Minute
)

func rlPath() string {
	return filepath.Join(paths.EdgeDir(), "enroll_rl.json")
}

func loadRL() map[string]attempt {
	b, err := os.ReadFile(rlPath())
	if err != nil {
		return map[string]attempt{}
	}
	var m map[string]attempt
	if json.Unmarshal(b, &m) != nil || m == nil {
		return map[string]attempt{}
	}
	return m
}

func saveRL(m map[string]attempt) {
	_ = os.MkdirAll(filepath.Dir(rlPath()), 0o700)
	b, _ := json.Marshal(m)
	_ = os.WriteFile(rlPath(), b, 0o600)
}

// AllowEnroll returns false if IP exceeded enroll rate (persisted).
func AllowEnroll(ip string) bool {
	if ip == "" {
		ip = "unknown"
	}
	rlMu.Lock()
	defer rlMu.Unlock()
	now := time.Now().Unix()
	m := loadRL()
	a, ok := m[ip]
	if !ok || now > a.Until {
		m[ip] = attempt{N: 1, Until: time.Now().Add(enrollWindow).Unix()}
		saveRL(m)
		return true
	}
	if a.N >= enrollMaxPerWindow {
		return false
	}
	a.N++
	m[ip] = a
	saveRL(m)
	return true
}
