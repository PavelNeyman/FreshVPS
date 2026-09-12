package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/PavelNeyman/netductor/internal/nodes"
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
	case "id":
		fmt.Println(nodes.LocalStableID())
	default:
		fmt.Fprintln(os.Stderr, "usage: netductor nodes list|rename|sync-local|id")
		os.Exit(2)
	}
}
