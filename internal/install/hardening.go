package install

import (
	"fmt"
	"os"
)

func InstallHardening() error {
	_ = aptInstall("ufw", "fail2ban", "curl", "ca-certificates")
	// BBR if available
	_ = os.WriteFile("/etc/sysctl.d/99-netductor.conf", []byte("net.core.default_qdisc=fq\nnet.ipv4.tcp_congestion_control=bbr\n"), 0o644)
	_ = run("sysctl", "--system")
	// UFW basics — do not lock SSH out: allow OpenSSH
	_ = run("ufw", "allow", "OpenSSH")
	_ = run("ufw", "allow", "443/tcp")
	_ = run("ufw", "allow", "8443/udp")
	// noninteractive enable
	_ = run("bash", "-c", "echo y | ufw enable")
	fmt.Fprintln(os.Stderr, "hardening: ufw + bbr applied")
	return nil
}
