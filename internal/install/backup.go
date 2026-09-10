package install

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/PavelNeyman/netductor/internal/paths"
)

func InstallBackup() error {
	// timer unit that runs netductor backup
	bin := "/usr/local/bin/netductor"
	unit := fmt.Sprintf(`[Unit]
Description=Netductor backup once
After=network-online.target

[Service]
Type=oneshot
ExecStart=%s backup
`, bin)
	timer := `[Unit]
Description=Netductor daily backup

[Timer]
OnCalendar=daily
Persistent=true
RandomizedDelaySec=30m

[Install]
WantedBy=timers.target
`
	if err := writeUnit("netductor-backup.service", unit); err != nil {
		return err
	}
	if err := writeUnit("netductor-backup.timer", timer); err != nil {
		return err
	}
	_ = run("systemctl", "enable", "netductor-backup.timer")
	_ = run("systemctl", "start", "netductor-backup.timer")
	fmt.Fprintln(os.Stderr, "backup timer: daily → /var/lib/netductor/backups/")
	return nil
}

// Backup creates tar.gz of etc + selected state (not huge metrics history by default).
func Backup() (string, error) {
	dir := filepath.Join(paths.StateDir(), "backups")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	name := fmt.Sprintf("netductor-%s.tar.gz", time.Now().UTC().Format("20060102-150405"))
	out := filepath.Join(dir, name)
	etc := paths.EtcDir()
	// pack etc + sing-box conf + blocky + secrets already under etc
	args := []string{"-czf", out, "-C", "/", 
		etc[1:], // strip leading /
		"usr/local/etc/sing-box",
		"etc/blocky",
	}
	// ignore missing
	_ = run("tar", args...)
	// verify file exists
	if st, err := os.Stat(out); err != nil || st.Size() == 0 {
		// fallback: only etc
		out2 := filepath.Join(dir, name)
		_ = run("tar", "-czf", out2, "-C", filepath.Dir(etc), filepath.Base(etc))
		out = out2
	}
	// prune keep last 14
	entries, _ := os.ReadDir(dir)
	if len(entries) > 14 {
		for i := 0; i < len(entries)-14; i++ {
			_ = os.Remove(filepath.Join(dir, entries[i].Name()))
		}
	}
	_ = os.Chmod(out, 0o600)
	return out, nil
}
