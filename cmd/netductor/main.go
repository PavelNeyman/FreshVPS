// Netductor — network control plane CLI (orchestrator).
// During migration this binary bridges to legacy FreshVPS tools where needed.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Set via -ldflags "-X main.version=..."
var version = "0.7.0-dev"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	switch cmd {
	case "version", "-v", "--version":
		fmt.Printf("netductor %s\n", version)
	case "help", "-h", "--help":
		printHelp()
	case "doctor":
		runDoctor(args)
	case "vpn":
		runVPN(args)
	case "edge":
		runEdge(args)
	case "status":
		runStatus()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`netductor — personal network control plane

Usage:
  netductor <command> [args]

Commands:
  version          Print version
  doctor           Run health checks (bridges to freshvps-doctor if present)
  status           Short local status (systemd units if available)
  vpn list         List VPN users
  vpn <args...>    Pass through to freshvps-vpn
  edge list        List edge agents
  edge cmd ...     Enqueue edge command
  help             This help

Legacy FreshVPS CLIs remain supported during migration.
Docs: https://github.com/PavelNeyman/FreshVPS
`)
}

func lookPath(names ...string) string {
	for _, n := range names {
		if p, err := exec.LookPath(n); err == nil {
			return p
		}
		for _, dir := range []string{"/usr/local/bin", "/opt/freshvps/bin", "/opt/netductor/bin"} {
			cand := filepath.Join(dir, n)
			if st, err := os.Stat(cand); err == nil && !st.IsDir() {
				return cand
			}
		}
	}
	return ""
}

func runDoctor(args []string) {
	bin := lookPath("freshvps-doctor", "netductor-doctor")
	if bin == "" {
		fmt.Fprintln(os.Stderr, "doctor: freshvps-doctor not found; install stack first or wait for native doctor")
		os.Exit(1)
	}
	c := exec.Command(bin, args...)
	c.Stdout, c.Stderr, c.Stdin = os.Stdout, os.Stderr, os.Stdin
	if err := c.Run(); err != nil {
		os.Exit(exitCode(err))
	}
}

func runVPN(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: netductor vpn list|add|link|...")
		os.Exit(2)
	}
	bin := lookPath("freshvps-vpn")
	if bin == "" {
		fmt.Fprintln(os.Stderr, "vpn: freshvps-vpn not found")
		os.Exit(1)
	}
	c := exec.Command(bin, args...)
	c.Stdout, c.Stderr, c.Stdin = os.Stdout, os.Stderr, os.Stdin
	if err := c.Run(); err != nil {
		os.Exit(exitCode(err))
	}
}

func runEdge(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: netductor edge list|cmd <device_id> <action> [arg]")
		os.Exit(2)
	}
	bin := lookPath("freshvps-vpn")
	if bin == "" {
		fmt.Fprintln(os.Stderr, "edge: freshvps-vpn not found (edge-list/cmd bridge)")
		os.Exit(1)
	}
	var vpnArgs []string
	switch args[0] {
	case "list":
		vpnArgs = []string{"edge-list"}
	case "cmd":
		vpnArgs = append([]string{"edge-cmd"}, args[1:]...)
	default:
		fmt.Fprintf(os.Stderr, "edge: unknown subcommand %q\n", args[0])
		os.Exit(2)
	}
	c := exec.Command(bin, vpnArgs...)
	c.Stdout, c.Stderr, c.Stdin = os.Stdout, os.Stderr, os.Stdin
	if err := c.Run(); err != nil {
		os.Exit(exitCode(err))
	}
}

func runStatus() {
	units := []string{"sing-box", "blocky", "freshvps-api", "freshvps-telegram-bot"}
	fmt.Println("netductor status")
	for _, u := range units {
		out, err := exec.Command("systemctl", "is-active", u).Output()
		st := strings.TrimSpace(string(out))
		if err != nil && st == "" {
			st = "inactive"
		}
		fmt.Printf("  %s: %s\n", u, st)
	}
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode()
	}
	return 1
}
