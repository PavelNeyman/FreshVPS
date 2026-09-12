package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/PavelNeyman/netductor/internal/install"
	"github.com/PavelNeyman/netductor/internal/paths"
	"github.com/PavelNeyman/netductor/internal/vpn"
)

func runRelay(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: netductor relay export|join|links|status")
		os.Exit(2)
	}
	switch args[0] {
	case "export":
		out := "bundle.json"
		sni := "ya.ru"
		for i := 1; i < len(args); i++ {
			switch args[i] {
			case "-o", "--output":
				if i+1 < len(args) {
					i++
					out = args[i]
				}
			case "--sni":
				if i+1 < len(args) {
					i++
					sni = args[i]
				}
			}
		}
		b, err := vpn.ExportRelayBundle(sni)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		raw, _ := json.MarshalIndent(b, "", "  ")
		if err := os.WriteFile(out, append(raw, '\n'), 0o600); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("wrote", out)
		fmt.Println("Copy to RU VPS and run: netductor relay join", out)
		fmt.Println("uplink user:", vpn.RelayUplinkName, "uuid="+b.UplinkUUID)
	case "join":
		path := "bundle.json"
		if len(args) > 1 {
			path = args[1]
		}
		if err := install.InstallRelay(path); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "links":
		dir := filepath.Join(paths.StateDir(), "relay", "clients")
		ents, err := os.ReadDir(dir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "no relay client links — run join on RU VPS first")
			os.Exit(1)
		}
		for _, e := range ents {
			if e.IsDir() {
				continue
			}
			b, _ := os.ReadFile(filepath.Join(dir, e.Name()))
			fmt.Printf("## %s\n%s\n", e.Name(), string(b))
		}
	case "status":
		b, err := os.ReadFile(filepath.Join(paths.StateDir(), "relay", "bundle.json"))
		if err != nil {
			fmt.Println("role: not a relay (no bundle)")
			return
		}
		fmt.Println(string(b))
	default:
		fmt.Fprintln(os.Stderr, "unknown relay subcommand")
		os.Exit(2)
	}
}
