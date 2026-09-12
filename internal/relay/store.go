package relay

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/PavelNeyman/netductor/internal/paths"
)

type Device struct {
	ID         string    `json:"id"`
	Token      string    `json:"token"` // agent auth
	Name       string    `json:"name"`
	PublicIP   string    `json:"public_ip"`
	PBK        string    `json:"pbk"`
	SID        string    `json:"sid"`
	SNI        string    `json:"sni"`
	Version    string    `json:"version"`
	SingBoxOK  bool      `json:"singbox_ok"`
	ConfigVer  int       `json:"config_ver"` // last applied
	LastSeen   time.Time `json:"last_seen"`
	CreatedAt  time.Time `json:"created_at"`
}

type registry struct {
	ConfigVer   int      `json:"config_ver"`
	ExitEnabled bool     `json:"exit_enabled"` // abroad → exit via RU
	Devices     []Device `json:"devices"`
}

var mu sync.Mutex

func path() string {
	return filepath.Join(paths.StateDir(), "relay", "devices.json")
}

func load() (*registry, error) {
	_ = os.MkdirAll(filepath.Dir(path()), 0o700)
	b, err := os.ReadFile(path())
	if err != nil {
		if os.IsNotExist(err) {
			return &registry{ConfigVer: 1}, nil
		}
		return nil, err
	}
	var r registry
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func save(r *registry) error {
	_ = os.MkdirAll(filepath.Dir(path()), 0o700)
	raw, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	tmp := path() + ".tmp"
	if err := os.WriteFile(tmp, append(raw, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path())
}

func BumpConfigVer() int {
	mu.Lock()
	defer mu.Unlock()
	r, err := load()
	if err != nil {
		return 0
	}
	r.ConfigVer++
	_ = save(r)
	return r.ConfigVer
}

func ConfigVer() int {
	mu.Lock()
	defer mu.Unlock()
	r, _ := load()
	if r == nil {
		return 1
	}
	return r.ConfigVer
}

func IssueToken(name string) (id, token string, err error) {
	mu.Lock()
	defer mu.Unlock()
	r, err := load()
	if err != nil {
		return "", "", err
	}
	var tb, ib [16]byte
	_, _ = rand.Read(tb[:])
	_, _ = rand.Read(ib[:])
	token = hex.EncodeToString(tb[:])
	id = "relay-" + hex.EncodeToString(ib[:8])
	if name == "" {
		name = id
	}
	r.Devices = append(r.Devices, Device{
		ID: id, Token: token, Name: name, CreatedAt: time.Now().UTC(), SNI: "ya.ru",
	})
	if r.ConfigVer < 1 {
		r.ConfigVer = 1
	}
	return id, token, save(r)
}

func FindByToken(token string) *Device {
	mu.Lock()
	defer mu.Unlock()
	r, err := load()
	if err != nil {
		return nil
	}
	for i := range r.Devices {
		if r.Devices[i].Token == token {
			d := r.Devices[i]
			return &d
		}
	}
	return nil
}

type HeartbeatIn struct {
	PublicIP  string `json:"public_ip"`
	PBK       string `json:"pbk"`
	SID       string `json:"sid"`
	SNI       string `json:"sni"`
	Version   string `json:"version"`
	SingBoxOK bool   `json:"singbox_ok"`
	ConfigVer int    `json:"config_ver"`
}

func Heartbeat(token string, in HeartbeatIn) (*Device, int, error) {
	mu.Lock()
	defer mu.Unlock()
	r, err := load()
	if err != nil {
		return nil, 0, err
	}
	for i := range r.Devices {
		if r.Devices[i].Token != token {
			continue
		}
		r.Devices[i].PublicIP = in.PublicIP
		r.Devices[i].PBK = in.PBK
		r.Devices[i].SID = in.SID
		if in.SNI != "" {
			r.Devices[i].SNI = in.SNI
		}
		r.Devices[i].Version = in.Version
		r.Devices[i].SingBoxOK = in.SingBoxOK
		r.Devices[i].ConfigVer = in.ConfigVer
		r.Devices[i].LastSeen = time.Now().UTC()
		_ = save(r)
		d := r.Devices[i]
		return &d, r.ConfigVer, nil
	}
	return nil, 0, os.ErrNotExist
}

func List() []Device {
	mu.Lock()
	defer mu.Unlock()
	r, err := load()
	if err != nil {
		return nil
	}
	out := make([]Device, len(r.Devices))
	copy(out, r.Devices)
	return out
}

func Online(d Device, within time.Duration) bool {
	if d.LastSeen.IsZero() {
		return false
	}
	return time.Since(d.LastSeen) < within
}

func ExitEnabled() bool {
	mu.Lock()
	defer mu.Unlock()
	r, err := load()
	if err != nil || r == nil {
		return false
	}
	return r.ExitEnabled
}

func SetExitEnabled(on bool) error {
	mu.Lock()
	defer mu.Unlock()
	r, err := load()
	if err != nil {
		return err
	}
	r.ExitEnabled = on
	r.ConfigVer++
	return save(r)
}
