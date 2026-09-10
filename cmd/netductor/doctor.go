package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/PavelNeyman/netductor/internal/paths"
)

func runDoctorNative() int {
	ok, fail, warn := 0, 0, 0
	check := func(name string, good bool) {
		if good {
			fmt.Printf("OK   %s\n", name)
			ok++
		} else {
			fmt.Printf("FAIL %s\n", name)
			fail++
		}
	}
	warnCheck := func(name string, good bool) {
		if good {
			fmt.Printf("OK   %s\n", name)
			ok++
		} else {
			fmt.Printf("WARN %s\n", name)
			warn++
		}
	}
	exists := func(p string) bool {
		_, err := os.Stat(p)
		return err == nil
	}
	active := func(unit string) bool {
		out, _ := exec.Command("systemctl", "is-active", unit).Output()
		return strings.TrimSpace(string(out)) == "active"
	}
	curlOK := func(url string) bool {
		return exec.Command("curl", "-fsS", "--max-time", "2", "-o", "/dev/null", url).Run() == nil
	}

	host, _ := os.Hostname()
	fmt.Printf("Netductor doctor — %s\n", host)
	etc := paths.EtcDir()
	opt := paths.OptDir()
	state := paths.StateDir()

	if b, err := os.ReadFile(filepath.Join(state, "installed_version")); err == nil {
		fmt.Printf("INFO installed_version=%s\n", strings.TrimSpace(string(b)))
	}
	check("debian", exists("/etc/debian_version"))
	check("secrets", exists(filepath.Join(etc, "secrets")))
	check("READY.txt", exists(filepath.Join(etc, "READY.txt")))
	check("vpn-users.json", exists(filepath.Join(etc, "vpn-users.json")))
	check("operator subscription", exists(filepath.Join(etc, "clients", "operator", "subscription.txt")))
	check("netductor", lookPath("netductor") != "")
	check("sing-box binary", exists("/usr/local/bin/sing-box"))
	check("sing-box unit", active("sing-box"))
	if exists("/usr/local/bin/sing-box") && exists("/usr/local/etc/sing-box/config.json") {
		check("sing-box config", exec.Command("/usr/local/bin/sing-box", "check", "-c", "/usr/local/etc/sing-box/config.json").Run() == nil)
	} else {
		check("sing-box config", false)
	}
	check("blocky unit", active("blocky"))
	warnCheck("netductor-api", active("netductor-api"))
	warnCheck("netductor-telegram-bot", active("netductor-telegram-bot"))

	if active("netductor-api") {
		warnCheck("api health :8787", curlOK("http://127.0.0.1:8787/health"))
	}
	if exists("/etc/systemd/system/netductor-metrics.timer") {
		warnCheck("metrics timer", active("netductor-metrics.timer"))
		warnCheck("metrics latest.json", exists(filepath.Join(state, "metrics", "latest.json")))
	}
	_ = opt
	fmt.Printf("\nSummary: ok=%d fail=%d warn=%d\n", ok, fail, warn)
	if fail > 0 {
		return 1
	}
	return 0
}
