package edge

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	mu       sync.Mutex
	EdgeDir  = env("FRESHVPS_EDGE_DIR", "/var/lib/freshvps/edge")
	TokenFile = env("NETDUCTOR_EDGE_TOKEN_FILE", "/etc/freshvps/secrets/edge_token")
)

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func ensure() error {
	return os.MkdirAll(EdgeDir, 0o700)
}

func Token() string {
	b, err := os.ReadFile(TokenFile)
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

func devicesPath() string { return filepath.Join(EdgeDir, "devices.json") }
func commandsPath() string { return filepath.Join(EdgeDir, "commands.json") }
func resultsPath() string { return filepath.Join(EdgeDir, "results.jsonl") }

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

func randomID() string {
	var b [6]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
