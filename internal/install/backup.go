package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	fmt.Fprintln(os.Stderr, "backup timer: daily → local (+ offsite if configured)")
	return nil
}

// Offsite config: /etc/netductor/backup.offsite
//   method=scp|http|rsync
//   target=user@host:/path   OR  https://example/upload
//   scp_opts=-i /root/.ssh/id_ed25519
func loadOffsite() (method, target, extra string) {
	b, err := os.ReadFile(filepath.Join(paths.EtcDir(), "backup.offsite"))
	if err != nil {
		return "", "", ""
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(k) {
		case "method":
			method = strings.TrimSpace(v)
		case "target":
			target = strings.TrimSpace(v)
		case "scp_opts", "extra":
			extra = strings.TrimSpace(v)
		}
	}
	return method, target, extra
}

func uploadOffsite(localPath string) error {
	method, target, extra := loadOffsite()
	if method == "" || target == "" {
		return nil
	}
	switch method {
	case "scp":
		args := []string{}
		if extra != "" {
			args = append(args, strings.Fields(extra)...)
		}
		args = append(args, localPath, target)
		return run("scp", args...)
	case "rsync":
		args := []string{"-az"}
		if extra != "" {
			args = append(args, strings.Fields(extra)...)
		}
		args = append(args, localPath, target)
		return run("rsync", args...)
	case "http", "https", "curl":
		// POST file as body
		return run("curl", "-fsS", "-X", "PUT", "--data-binary", "@"+localPath, target)
	default:
		return fmt.Errorf("unknown offsite method %s", method)
	}
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
	out := plain
	if key != "" {
		enc := plain + ".enc"
		err := run("openssl", "enc", "-aes-256-cbc", "-salt", "-pbkdf2",
			"-in", plain, "-out", enc, "-pass", "pass:"+key)
		_ = os.Remove(plain)
		if err != nil {
			return "", err
		}
		out = enc
	}
	_ = os.Chmod(out, 0o600)
	if err := uploadOffsite(out); err != nil {
		fmt.Fprintf(os.Stderr, "offsite upload: %v\n", err)
	}
	pruneBackups(dir, 14)
	return out, nil
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
