package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/PavelNeyman/netductor/internal/vpn"
)

func runTUI(args []string) {
	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(`
╔══════════════════════════════════════╗
║           Netductor menu             ║
╚══════════════════════════════════════╝
  1) Status (services)
  2) Doctor
  3) VPN — list users
  4) VPN — add user
  5) VPN — session token
  6) Edge — list devices
  7) Probes (live)
  8) Collect metrics sample
  9) Install (legacy install.sh)
  0) Quit

> `)
		line, err := in.ReadString('\n')
		if err != nil {
			return
		}
		choice := strings.TrimSpace(line)
		switch choice {
		case "1", "status":
			runStatus()
		case "2", "doctor":
			_ = runDoctorNative()
		case "3", "list":
			users, err := vpn.List()
			if err != nil {
				fmt.Println("error:", err)
				continue
			}
			if len(users) == 0 {
				fmt.Println("(no users)")
			}
			for _, u := range users {
				en := "off"
				if u.Enabled {
					en = "on"
				}
				fmt.Printf("  %s  [%s]  %s  %s\n", u.Name, en, u.UUID, u.Note)
			}
		case "4", "add":
			fmt.Print("Name: ")
			name, _ := in.ReadString('\n')
			name = strings.TrimSpace(name)
			fmt.Print("Note (optional): ")
			note, _ := in.ReadString('\n')
			note = strings.TrimSpace(note)
			if name == "" {
				fmt.Println("cancelled")
				continue
			}
			out, err := vpn.Add(name, note)
			fmt.Println(out)
			if err != nil {
				fmt.Println("error:", err)
			}
		case "5", "session":
			fmt.Print("Hours [72]: ")
			h, _ := in.ReadString('\n')
			hours := 72
			if s := strings.TrimSpace(h); s != "" {
				if n, err := strconv.Atoi(s); err == nil {
					hours = n
				}
			}
			tok, exp, err := vpn.CreateSession(hours)
			if err != nil {
				fmt.Println("error:", err)
				continue
			}
			fmt.Println(tok)
			fmt.Fprintf(os.Stderr, "expires_unix=%d hours=%d\n", exp, hours)
		case "6", "edge":
			runEdgeList()
		case "7", "probe":
			runProbe(nil)
		case "8", "collect":
			_ = runCollect()
		case "9", "install":
			fmt.Println("Launching install.sh …")
			runInstall(nil)
		case "0", "q", "quit", "exit":
			return
		default:
			if choice != "" {
				fmt.Println("unknown:", choice)
			}
		}
	}
}
