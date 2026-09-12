package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PavelNeyman/netductor/internal/install"
	"github.com/PavelNeyman/netductor/internal/paths"
	"github.com/PavelNeyman/netductor/internal/relay"
	"github.com/PavelNeyman/netductor/internal/vpn"
)

func runRelay(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: netductor relay export|join|links|status|exit on|exit off|exit")
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
		if id, tok, err := relay.IssueToken("relay"); err == nil {
			b.AgentID = id
			b.AgentToken = tok
			b.CoreAgentURL = "http://" + b.CoreIP + ":8788"
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
	case "agent":
		tokB, _ := os.ReadFile("/etc/netductor/secrets/relay_agent_token")
		urlB, _ := os.ReadFile("/etc/netductor/secrets/relay_core_url")
		tok := strings.TrimSpace(string(tokB))
		url := strings.TrimSpace(string(urlB))
		if len(args) > 1 && args[1] != "" {
			// optional overrides
		}
		if tok == "" || url == "" {
			fmt.Fprintln(os.Stderr, "missing relay_agent_token or relay_core_url")
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "relay agent →", url)
		relay.AgentLoop(url, tok, 30*time.Second)
	case "exit":
		if len(args) < 2 || args[1] == "status" {
			fmt.Println("exit_enabled", relay.ExitEnabled())
			return
		}
		on := args[1] == "on" || args[1] == "1" || args[1] == "true"
		if args[1] == "off" || args[1] == "0" || args[1] == "false" {
			on = false
		} else if args[1] != "on" && args[1] != "1" && args[1] != "true" {
			fmt.Fprintln(os.Stderr, "usage: netductor relay exit on|off")
			os.Exit(2)
		}
		if err := relay.SetExitEnabled(on); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		_ = vpn.ApplyConfig()
		fmt.Println("exit_enabled", on)
		fmt.Println("sing-box re-applied on core; relays will sync on next agent poll")
	case "status":
		devs := relay.List()
		if len(devs) == 0 {
			b, err := os.ReadFile(filepath.Join(paths.StateDir(), "relay", "bundle.json"))
			if err != nil {
				fmt.Println("no relays registered")
				return
			}
			fmt.Println(string(b))
			return
		}
		for _, d := range devs {
			on := "offline"
			if relay.Online(d, 2*time.Minute) {
				on = "online"
			}
			fmt.Printf("%s	%s	%s	ip=%s	sb=%v	ver=%d\n", d.ID, d.Name, on, d.PublicIP, d.SingBoxOK, d.ConfigVer)
		}
		fmt.Println("config_ver", relay.ConfigVer())
	default:
		fmt.Fprintln(os.Stderr, "unknown relay subcommand")
		os.Exit(2)
	}
}
