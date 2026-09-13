package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/PavelNeyman/netductor/internal/nodes"
	"github.com/PavelNeyman/netductor/internal/notify"
	"github.com/PavelNeyman/netductor/internal/relay"
)

func syncRelaysIntoNodes() {
	_ = relay.PruneDuplicates()
	for _, d := range relay.List() {
		host := d.Name
		if host == "" || host == "relay" {
			host = "nd-relay-" + strings.ReplaceAll(d.PublicIP, ".", "-")
		}
		st := "offline"
		if relay.Online(d, 2*time.Minute) {
			st = "online"
		}
		ls := d.LastSeen.Unix()
		if ls <= 0 {
			ls = time.Now().Unix()
		}
		_, _ = nodes.UpsertFromDevice(nodes.Node{
			ID: d.ID, Hostname: host, Role: "relay", Kind: "vps",
			PublicIP: d.PublicIP, Status: st, LastSeen: ls,
		})
	}
}

func runNodes(args []string) {
	if len(args) == 0 {
		args = []string{"list"}
	}
	switch args[0] {
	case "list":
		syncRelaysIntoNodes()
		list, err := nodes.List()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, n := range list {
			fmt.Printf("id=%s\thost=%s\trole=%s\tkind=%s\tip=%s\tstatus=%s\tdesired=%s\n",
				n.ID, n.Hostname, n.Role, n.Kind, n.PublicIP, n.Status, n.DesiredHN)
		}
	case "rename":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: netductor nodes rename <id> <hostname>")
			fmt.Fprintln(os.Stderr, "  hostname format: nd-<role>-<marker>  e.g. nd-relay-msk01")
			os.Exit(2)
		}
		syncRelaysIntoNodes()
		n, err := nodes.SetDesiredHostname(args[1], args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		_ = relay.Rename(args[1], args[2])
		fmt.Printf("desired_hostname=%s for %s\n", n.DesiredHN, n.ID)
	case "sync-local":
		if err := nodes.SyncLocalHostname(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "local-cmd":
		if len(args) < 2 {
			os.Exit(2)
		}
		switch args[1] {
		case "reboot":
			fmt.Println("core reboot scheduled in 3s")
			go func() {
				time.Sleep(3 * time.Second)
				_ = exec.Command("systemctl", "reboot").Run()
			}()
		case "upgrade":
			logf := "/var/lib/netductor/core-upgrade.log"
			_ = os.MkdirAll("/var/lib/netductor", 0o755)
			_ = os.WriteFile(logf, append([]byte("started"), 10), 0o600)
			go func() {
				script := "export DEBIAN_FRONTEND=noninteractive; apt-get update -qq 2>&1 | tail -5; apt-get -y -o Dpkg::Options::=--force-confdef -o Dpkg::Options::=--force-confold upgrade 2>&1 | tail -40; wget -qO /tmp/nd.bin https://github.com/PavelNeyman/netductor/releases/download/v0.7.0-dev/netductor-linux-amd64; wget -qO /tmp/nd-tg.bin https://github.com/PavelNeyman/netductor/releases/download/v0.7.0-dev/netductor-tg-linux-amd64; cp /tmp/nd.bin /usr/local/bin/netductor; install -m 755 /tmp/nd-tg.bin /opt/netductor/bin/netductor-tg; install -m 755 /tmp/nd-tg.bin /usr/local/bin/netductor-tg; systemctl restart sing-box netductor-api 2>&1 || true; echo CORE_UPGRADE_DONE"
				out, err := exec.Command("bash", "-c", script).CombinedOutput()
				msg := string(out)
				ok := err == nil
				if err != nil {
					msg += string([]byte{10}) + "ERR: " + err.Error()
				}
				_ = os.WriteFile(logf, []byte(msg), 0o600)
				_ = notify.TelegramCmd("nd-core", "upgrade", ok, msg)
				_ = exec.Command("systemctl", "restart", "netductor-telegram-bot").Run()
			}()
			fmt.Println("core upgrade started in background")
			fmt.Println("log:", logf)
		default:
			fmt.Println("unknown", args[1])
		}
	case "id":
		fmt.Println(nodes.LocalStableID())
	default:
		fmt.Fprintln(os.Stderr, "usage: netductor nodes list|rename|sync-local|id")
		os.Exit(2)
	}
}
