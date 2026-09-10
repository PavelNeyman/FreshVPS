package edge

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/PavelNeyman/netductor/internal/paths"
)

var mu sync.Mutex

func ensure() error {
	return os.MkdirAll(paths.EdgeDir(), 0o700)
}

func Token() string {
	b, err := os.ReadFile(paths.EdgeTokenFile())
	if err != nil {
		return ""
	}
	return string(bytesTrim(b))
}

func bytesTrim(b []byte) []byte {
	for len(b) > 0 && (b[len(b)-1] == '\n' || b[len(b)-1] == '\r' || b[len(b)-1] == ' ') {
		b = b[:len(b)-1]
	}
	return b
}

func ValidBearer(auth string) bool {
	tok := Token()
	if tok == "" || len(auth) < 8 {
		return false
	}
	const p = "Bearer "
	if len(auth) > len(p) && auth[:len(p)] == p {
		return auth[len(p):] == tok
	}
	return false
}

type Device map[string]any

func devicesPath() string  { return filepath.Join(paths.EdgeDir(), "devices.json") }
func commandsPath() string { return filepath.Join(paths.EdgeDir(), "commands.json") }
func resultsPath() string  { return filepath.Join(paths.EdgeDir(), "results.jsonl") }

func loadDevices() map[string]Device {
	_ = ensure()
	b, err := os.ReadFile(devicesPath())
	if err != nil {
		return map[string]Device{}
	}
	var wrap struct {
		Devices map[string]Device `json:"devices"`
	}
	if json.Unmarshal(b, &wrap) != nil || wrap.Devices == nil {
		return map[string]Device{}
	}
	return wrap.Devices
}

func saveDevices(m map[string]Device) error {
	_ = ensure()
	b, _ := json.MarshalIndent(map[string]any{"devices": m}, "", "  ")
	tmp := devicesPath() + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, devicesPath())
}

func Heartbeat(payload map[string]any) {
	mu.Lock()
	defer mu.Unlock()
	m := loadDevices()
	did, _ := payload["device_id"].(string)
	if did == "" {
		did = "unknown"
	}
	payload["last_seen"] = time.Now().Unix()
	m[did] = Device(payload)
	_ = saveDevices(m)
}

func ListDevices() []Device {
	mu.Lock()
	defer mu.Unlock()
	m := loadDevices()
	now := time.Now().Unix()
	out := make([]Device, 0, len(m))
	for id, d := range m {
		row := Device{}
		for k, v := range d {
			row[k] = v
		}
		row["device_id"] = id
		var last int64
		switch v := d["last_seen"].(type) {
		case float64:
			last = int64(v)
		case int64:
			last = v
		}
		row["healthy"] = (now - last) < 120
		out = append(out, row)
	}
	return out
}

func EnqueueCmd(deviceID, action, arg string) string {
	mu.Lock()
	defer mu.Unlock()
	_ = ensure()
	var wrap struct {
		Pending []map[string]any `json:"pending"`
	}
	if b, err := os.ReadFile(commandsPath()); err == nil {
		_ = json.Unmarshal(b, &wrap)
	}
	id := randomID()
	wrap.Pending = append(wrap.Pending, map[string]any{
		"id": id, "device_id": deviceID, "action": action, "arg": arg, "created": time.Now().Unix(),
	})
	b, _ := json.MarshalIndent(wrap, "", "  ")
	tmp := commandsPath() + ".tmp"
	_ = os.WriteFile(tmp, append(b, '\n'), 0o600)
	_ = os.Rename(tmp, commandsPath())
	return id
}

func PollCommands(deviceID string) []map[string]any {
	mu.Lock()
	defer mu.Unlock()
	var wrap struct {
		Pending []map[string]any `json:"pending"`
	}
	if b, err := os.ReadFile(commandsPath()); err == nil {
		_ = json.Unmarshal(b, &wrap)
	}
	var mine, rest []map[string]any
	for _, c := range wrap.Pending {
		if c["device_id"] == deviceID {
			mine = append(mine, c)
		} else {
			rest = append(rest, c)
		}
	}
	wrap.Pending = rest
	b, _ := json.MarshalIndent(wrap, "", "  ")
	tmp := commandsPath() + ".tmp"
	_ = os.WriteFile(tmp, append(b, '\n'), 0o600)
	_ = os.Rename(tmp, commandsPath())
	if mine == nil {
		mine = []map[string]any{}
	}
	return mine
}

func ListResults() []map[string]any {
	mu.Lock()
	defer mu.Unlock()
	b, err := os.ReadFile(resultsPath())
	if err != nil {
		return []map[string]any{}
	}
	var out []map[string]any
	lines := strings.Split(string(b), "\n")
	// keep last 100
	start := 0
	if len(lines) > 100 {
		start = len(lines) - 100
	}
	for _, line := range lines[start:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var m map[string]any
		if json.Unmarshal([]byte(line), &m) == nil {
			out = append(out, m)
		}
	}
	return out
}

func CmdResult(payload map[string]any) {
	mu.Lock()
	defer mu.Unlock()
	_ = ensure()
	b, _ := json.Marshal(payload)
	f, err := os.OpenFile(resultsPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(b, '\n'))
}

func deviceDir(id string) string {
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, id)
	return filepath.Join(paths.EdgeDir(), safe)
}

func SaveBackup(deviceID string, r io.Reader) (string, error) {
	mu.Lock()
	defer mu.Unlock()
	dir := filepath.Join(deviceDir(deviceID), "backups")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	name := time.Now().UTC().Format("20060102-150405") + ".tar.gz"
	path := filepath.Join(dir, name)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	return path, nil
}

func SaveMetrics(deviceID string, payload map[string]any) error {
	mu.Lock()
	defer mu.Unlock()
	dir := deviceDir(deviceID)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	path := filepath.Join(dir, "metrics.jsonl")
	b, _ := json.Marshal(payload)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(b, '\n'))
	// trim large files roughly by size
	if st, e := os.Stat(path); e == nil && st.Size() > 5*1024*1024 {
		// keep simple: truncate not implemented fully
	}
	return err
}

func randomID() string {
	var b [6]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
