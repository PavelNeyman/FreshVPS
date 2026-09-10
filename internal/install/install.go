package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/PavelNeyman/netductor/internal/paths"
)

type Options struct {
	Components []string // empty = default core set
	Force      bool
}

func DefaultComponents() []string {
	return []string{"dirs", "hardening", "singbox", "blocky", "vpn-users", "api", "metrics", "telegram", "backup"}
}

func Run(opts Options) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("install requires root")
	}
	if _, err := os.Stat("/etc/debian_version"); err != nil {
		return fmt.Errorf("only Debian-like hosts supported for full install")
	}
	comps := opts.Components
	if len(comps) == 0 {
		comps = DefaultComponents()
	}
	if err := paths.EnsureLayout(); err != nil {
		return err
	}
	_ = copySelfToLocalBin()
	for _, c := range comps {
		fmt.Fprintf(os.Stderr, "==> %s\n", c)
		var err error
		switch c {
		case "dirs":
			err = paths.EnsureLayout()
		case "hardening":
			err = InstallHardening()
		case "singbox", "sing-box":
			err = InstallSingBox()
		case "blocky":
			err = InstallBlocky()
		case "vpn-users", "vpn":
			err = InstallVPNUsers()
		case "api", "netductor-api":
			err = InstallAPI()
		case "metrics":
			err = InstallMetrics()
		case "telegram", "tg":
			err = InstallTelegram()
		case "backup":
			err = InstallBackup()
		default:
			fmt.Fprintf(os.Stderr, "skip unknown component %s\n", c)
		}
		if err != nil {
			return fmt.Errorf("%s: %w", c, err)
		}
	}
	_ = os.WriteFile(filepath.Join(paths.StateDir(), "installed_version"), []byte("0.7.0-dev\n"), 0o644)
	writeReady()
	fmt.Fprintln(os.Stderr, "install finished")
	return nil
}

func arch() string {
	switch runtime.GOARCH {
	case "amd64", "arm64":
		return runtime.GOARCH
	case "arm":
		return "armv7"
	default:
		return runtime.GOARCH
	}
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func runOut(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return string(out), err
}

func aptInstall(pkgs ...string) error {
	_ = run("bash", "-c", "export DEBIAN_FRONTEND=noninteractive; apt-get update -y")
	args := append([]string{"install", "-y"}, pkgs...)
	return run("apt-get", args...)
}

func writeSecret(name, val string) error {
	dir := filepath.Join(paths.EtcDir(), "secrets")
	_ = os.MkdirAll(dir, 0o700)
	return os.WriteFile(filepath.Join(dir, name), []byte(val+"\n"), 0o600)
}

func readSecret(name string) string {
	b, err := os.ReadFile(filepath.Join(paths.EtcDir(), "secrets", name))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func writeUnit(name, content string) error {
	path := filepath.Join("/etc/systemd/system", name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	_ = run("systemctl", "daemon-reload")
	return nil
}

func enableStart(unit string) error {
	_ = run("systemctl", "enable", unit)
	return run("systemctl", "restart", unit)
}

func writeReady() {
	ip := ""
	if b, err := os.ReadFile(filepath.Join(paths.EtcDir(), "public_ip")); err == nil {
		ip = strings.TrimSpace(string(b))
	}
	txt := fmt.Sprintf("Netductor ready\nETC=%s\nSTATE=%s\nOPT=%s\nIP=%s\n",
		paths.EtcDir(), paths.StateDir(), paths.OptDir(), ip)
	_ = os.WriteFile(filepath.Join(paths.EtcDir(), "READY.txt"), []byte(txt), 0o644)
}


func copySelfToLocalBin() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	dest := "/usr/local/bin/netductor"
	if exe == dest {
		return nil
	}
	in, err := os.ReadFile(exe)
	if err != nil {
		return err
	}
	tmp := dest + ".tmp"
	if err := os.WriteFile(tmp, in, 0o755); err != nil {
		return err
	}
	return os.Rename(tmp, dest)
}
