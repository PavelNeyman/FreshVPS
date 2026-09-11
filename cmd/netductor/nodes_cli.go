package main

import (
	"fmt"
	"os"

	"github.com/PavelNeyman/netductor/internal/nodes"
)

func runNodes(args []string) {
	if len(args) == 0 {
		args = []string{"list"}
	}
	switch args[0] {
	case "list":
		list, err := nodes.List()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, n := range list {
			fmt.Printf("id=%s\thost=%s\trole=%s\tkind=%s\tip=%s\tdesired=%s\n",
				n.ID, n.Hostname, n.Role, n.Kind, n.PublicIP, n.DesiredHN)
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
