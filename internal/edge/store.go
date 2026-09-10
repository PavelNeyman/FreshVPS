package edge

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/PavelNeyman/netductor/internal/paths"
)

var mu sync.Mutex

const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusDenied   = "denied"
	StatusRevoked  = "revoked"
)

func ensure() error {
	return os.MkdirAll(paths.EdgeDir(), 0o700)
}

// BootstrapToken — shared secret only for enroll (not full access).
func BootstrapToken() string {
	b, err := os.ReadFile(filepath.Join(paths.EtcDir(), "secrets", "edge_bootstrap_token"))
	if err != nil {
		// fallback to legacy edge_token for transition
		return Token()
	}
	return string(bytesTrim(b))
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

func constEq(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func bearerRaw(auth string) string {
	const p = "Bearer "
	if len(auth) > len(p) && auth[:len(p)] == p {
		return auth[len(p):]
	}
	return ""
}

// ValidBearer: legacy global edge_token OR any approved device token.
func ValidBearer(auth string) bool {
	raw := bearerRaw(auth)
	if raw == "" {
		return false
	}
	if tok := Token(); tok != "" && constEq(raw, tok) {
		return true
	}
	mu.Lock()
	defer mu.Unlock()
	m := loadDevices()
	for _, d := range m {
		dt, _ := d["device_token"].(string)
		st, _ := d["status"].(string)
		if dt != "" && constEq(raw, dt) && st == StatusApproved {
			return true
		}
	}
	return false
}

// DeviceIDFromAuth returns device_id if auth is a device token.
func DeviceIDFromAuth(auth string) string {
	raw := bearerRaw(auth)
	if raw == "" {
		return ""
	}
	mu.Lock()
	defer mu.Unlock()
	m := loadDevices()
	for id, d := range m {
		dt, _ := d["device_token"].(string)
		if dt != "" && constEq(raw, dt) {
			return id
		}
	}
	return ""
}

func ValidBootstrap(auth string) bool {
	raw := bearerRaw(auth)
	bt := BootstrapToken()
	return raw != "" && bt != "" && constEq(raw, bt)
}

// RequireApproved: device token must map to approved device_id matching claim.
func RequireApproved(auth, deviceID string) bool {
	raw := bearerRaw(auth)
	if raw == "" || deviceID == "" {
		return false
	}
	// global token (operator tooling) may act as superuser
	if tok := Token(); tok != "" && constEq(raw, tok) {
		return true
	}
	mu.Lock()
	defer mu.Unlock()
	m := loadDevices()
	d, ok := m[deviceID]
	if !ok {
		return false
	}
	st, _ := d["status"].(string)
	dt, _ := d["device_token"].(string)
	return st == StatusApproved && dt != "" && constEq(raw, dt)
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

// Enroll registers or refreshes pending device. Returns status + optional token if already approved.
func Enroll(payload map[string]any) (status, deviceToken string, isNew bool) {
	mu.Lock()
	defer mu.Unlock()
	m := loadDevices()
	did, _ := payload["device_id"].(string)
	if did == "" {
		did = "unknown"
	}
	existing, ok := m[did]
	now := time.Now().Unix()
	if ok {
		st, _ := existing["status"].(string)
		if st == StatusApproved {
			// merge last_seen / inventory, keep token
			for k, v := range payload {
				if k == "device_token" || k == "status" {
					continue
				}
				existing[k] = v
			}
			existing["last_seen"] = now
			m[did] = existing
			_ = saveDevices(m)
			dt, _ := existing["device_token"].(string)
			return StatusApproved, dt, false
		}
		if st == StatusDenied || st == StatusRevoked {
			existing["last_seen"] = now
			existing["last_enroll"] = payload
			m[did] = existing
			_ = saveDevices(m)
			return st, "", false
		}
		// pending — update facts
		for k, v := range payload {
			existing[k] = v
		}
		existing["status"] = StatusPending
		existing["last_seen"] = now
		m[did] = existing
		_ = saveDevices(m)
		return StatusPending, "", false
	}
	row := Device{}
	for k, v := range payload {
		row[k] = v
	}
	row["device_id"] = did
	row["status"] = StatusPending
	row["enrolled_at"] = now
	row["last_seen"] = now
	m[did] = row
	_ = saveDevices(m)
	return StatusPending, "", true
}

func Approve(deviceID string) (deviceToken string, err error) {
	mu.Lock()
	defer mu.Unlock()
	m := loadDevices()
	d, ok := m[deviceID]
	if !ok {
		return "", fmt.Errorf("unknown device")
	}
	st, _ := d["status"].(string)
	if st == StatusApproved {
		dt, _ := d["device_token"].(string)
		return dt, nil
	}
	if st == StatusRevoked {
		// re-approve allowed → new token
	}
	tok := randomID() + randomID()
	d["status"] = StatusApproved
	d["device_token"] = tok
	d["approved_at"] = time.Now().Unix()
	m[deviceID] = d
	_ = saveDevices(m)
	return tok, nil
}

func Deny(deviceID string) error {
	mu.Lock()
	defer mu.Unlock()
	m := loadDevices()
	d, ok := m[deviceID]
	if !ok {
		return fmt.Errorf("unknown device")
	}
	d["status"] = StatusDenied
	delete(d, "device_token")
	d["denied_at"] = time.Now().Unix()
	m[deviceID] = d
	return saveDevices(m)
}

func Revoke(deviceID string) error {
	mu.Lock()
	defer mu.Unlock()
	m := loadDevices()
	d, ok := m[deviceID]
	if !ok {
		return fmt.Errorf("unknown device")
	}
	d["status"] = StatusRevoked
	delete(d, "device_token")
	d["revoked_at"] = time.Now().Unix()
	m[deviceID] = d
	return saveDevices(m)
}

func StatusOf(deviceID string) string {
	mu.Lock()
	defer mu.Unlock()
	d, ok := loadDevices()[deviceID]
	if !ok {
		return ""
	}
	st, _ := d["status"].(string)
	return st
}

func Heartbeat(payload map[string]any) {
	mu.Lock()
	defer mu.Unlock()
	m := loadDevices()
	did, _ := payload["device_id"].(string)
	if did == "" {
		did = "unknown"
	}
	cur, ok := m[did]
	if !ok {
		// unknown — ignore or soft pending without token path
		return
	}
	st, _ := cur["status"].(string)
	if st != StatusApproved {
		cur["last_seen"] = time.Now().Unix()
		m[did] = cur
		_ = saveDevices(m)
		return
	}
	for k, v := range payload {
		if k == "device_token" || k == "status" {
			continue
		}
		cur[k] = v
	}
	cur["last_seen"] = time.Now().Unix()
	m[did] = cur
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
			if k == "device_token" {
				continue // never leak
			}
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

func ListPending() []Device {
	var out []Device
	for _, d := range ListDevices() {
		if st, _ := d["status"].(string); st == StatusPending {
			out = append(out, d)
		}
	}
	return out
}

func EnqueueCmd(deviceID, action, arg string) string {
	mu.Lock()
	defer mu.Unlock()
	_ = ensure()
	if st := statusUnlocked(deviceID); st != StatusApproved && st != "" {
		return ""
	}
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

func statusUnlocked(deviceID string) string {
	d, ok := loadDevices()[deviceID]
	if !ok {
		return ""
	}
	st, _ := d["status"].(string)
	return st
}

func PollCommands(deviceID string) []map[string]any {
	mu.Lock()
	defer mu.Unlock()
	if st := statusUnlocked(deviceID); st != StatusApproved {
		return []map[string]any{}
	}
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

func ListBackups(deviceID string) []map[string]any {
	dir := filepath.Join(deviceDir(deviceID), "backups")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []map[string]any{}
	}
	var out []map[string]any
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, map[string]any{
			"name": e.Name(),
			"size": info.Size(),
			"mod":  info.ModTime().Unix(),
		})
	}
	return out
}

func BackupPath(deviceID, name string) (string, error) {
	if name == "" || strings.Contains(name, "..") || strings.Contains(name, "/") {
		return "", fmt.Errorf("invalid name")
	}
	path := filepath.Join(deviceDir(deviceID), "backups", name)
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return path, nil
}

func ListMetricsTail(deviceID string, max int) []map[string]any {
	if max <= 0 {
		max = 50
	}
	path := filepath.Join(deviceDir(deviceID), "metrics.jsonl")
	b, err := os.ReadFile(path)
	if err != nil {
		return []map[string]any{}
	}
	lines := strings.Split(string(b), "\n")
	start := 0
	if len(lines) > max {
		start = len(lines) - max
	}
	var out []map[string]any
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
	return err
}

func randomID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
