package install

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/PavelNeyman/netductor/internal/paths"
)

func InstallBackup() error {
	bin := "/usr/local/bin/netductor"
	if readSecret("backup_key") == "" {
		_ = writeSecret("backup_key", randomHex(32))
	}
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
	fmt.Fprintln(os.Stderr, "backup timer: daily encrypted → /var/lib/netductor/backups/")
	return nil
}

func Backup() (string, error) {
	dir := filepath.Join(paths.StateDir(), "backups")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	stamp := time.Now().UTC().Format("20060102-150405")
	plain := filepath.Join(dir, "netductor-"+stamp+".tar.gz")
	etc := paths.EtcDir()
	_ = run("tar", "-czf", plain, "-C", filepath.Dir(etc), filepath.Base(etc))
	if st, err := os.Stat(plain); err != nil || st.Size() == 0 {
		return "", fmt.Errorf("tar failed")
	}
	key := readSecret("backup_key")
	out := plain + ".enc"
	if key != "" {
		// AES-256-CBC via openssl (portable, no extra Go crypto deps for stream)
		err := run("openssl", "enc", "-aes-256-cbc", "-salt", "-pbkdf2",
			"-in", plain, "-out", out, "-pass", "pass:"+key)
		_ = os.Remove(plain)
		if err != nil {
			return "", err
		}
		_ = os.Chmod(out, 0o600)
		pruneBackups(dir, 14)
		return out, nil
	}
	_ = os.Chmod(plain, 0o600)
	pruneBackups(dir, 14)
	return plain, nil
}

func pruneBackups(dir string, keep int) {
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) <= keep {
		return
	}
	for i := 0; i < len(entries)-keep; i++ {
		_ = os.Remove(filepath.Join(dir, entries[i].Name()))
	}
}
