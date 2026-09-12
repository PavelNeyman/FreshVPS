package main

import (
	"fmt"
	"os"
	"time"

	"github.com/PavelNeyman/netductor/internal/nodes"
	"github.com/PavelNeyman/netductor/internal/relay"
)

func runNodes(args []string) {
	if len(args) == 0 {
		args = []string{"list"}
	}
	switch args[0] {
	case "list":
		_ = relay.PruneDuplicates()
		list, err := nodes.List()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		// Merge relay devices as role=relay (virtual rows for UI)
		seenIP := map[string]bool{}
		for _, n := range list {
			if n.PublicIP != "" {
				seenIP[n.PublicIP] = true
			}
			fmt.Printf("id=%s\thost=%s\trole=%s\tkind=%s\tip=%s\tstatus=%s\tdesired=%s\n",
				n.ID, n.Hostname, n.Role, n.Kind, n.PublicIP, n.Status, n.DesiredHN)
		}
		for _, d := range relay.List() {
			if d.PublicIP != "" && seenIP[d.PublicIP] {
				// already a node with same IP — still show relay metrics line if no role relay
			}
			st := "offline"
			if relay.Online(d, 2*time.Minute) {
				st = "online"
			}
			host := d.Name
			if host == "" || host == "relay" {
				host = "nd-relay-" + d.PublicIP
			}
			fmt.Printf("id=%s\thost=%s\trole=relay\tkind=vps\tip=%s\tstatus=%s\tcpu=%.0f\tmem=%d/%d\tload=%.2f\tdesired=\n",
				d.ID, host, d.PublicIP, st, d.CPUPercent, d.MemUsedMB, d.MemTotalMB, d.Load1)
		}
	case "rename":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: netductor nodes rename <id> <hostname>")
			fmt.Fprintln(os.Stderr, "  hostname format: nd-<role>-<marker>  e.g. nd-core-nl01")
			os.Exit(2)
		}
		n, err := nodes.SetDesiredHostname(args[1], args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("desired_hostname=%s for %s\n", n.DesiredHN, n.ID)
	case "sync-local":
		if err := nodes.SyncLocalHostname(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "id":
		fmt.Println(nodes.LocalStableID())
	default:
		fmt.Fprintln(os.Stderr, "usage: netductor nodes list|rename|sync-local|id")
		os.Exit(2)
	}
}
