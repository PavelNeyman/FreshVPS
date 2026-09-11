package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/PavelNeyman/netductor/internal/install"
	"github.com/PavelNeyman/netductor/internal/probes"
)

func lookPath(names ...string) string {
	for _, n := range names {
		if p, err := exec.LookPath(n); err == nil {
			return p
		}
		for _, d := range []string{"/usr/local/bin", "/opt/netductor/bin"} {
			c := filepath.Join(d, n)
			if st, err := os.Stat(c); err == nil && !st.IsDir() {
				return c
			}
		}
	}
	return ""
}


func runBridge(bin string, args []string) {
	if bin == "" {
		fmt.Fprintln(os.Stderr, "binary not found")
		os.Exit(1)
	}
	c := exec.Command(bin, args...)
	c.Stdout, c.Stderr, c.Stdin = os.Stdout, os.Stderr, os.Stdin
	if err := c.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			os.Exit(ee.ExitCode())
		}
		os.Exit(1)
	}
}


func runStatus() {
	for _, u := range []string{"sing-box", "blocky", "netductor-api", "netductor-telegram-bot"} {
		out, _ := exec.Command("systemctl", "is-active", u).Output()
		st := strings.TrimSpace(string(out))
		if st == "" {
			st = "inactive"
		}
		fmt.Printf("  %s: %s\n", u, st)
	}
}


func runBackupCmd() {
	path, err := install.Backup()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(path)
}


func runSelfInstall() {
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	dest := "/usr/local/bin/netductor"
	data, err := os.ReadFile(exe)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	tmp := dest + ".new"
	if err := os.WriteFile(tmp, data, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.Rename(tmp, dest); err != nil {
		fmt.Fprintf(os.Stderr, "wrote %s (dest busy). restart api then: mv %s %s\n", tmp, tmp, dest)
		return
	}
	fmt.Println("updated", dest)
	fmt.Println("restart: systemctl restart netductor-api")
}



func runInstall(args []string) {
	comps := []string{}
	for _, a := range args {
		if a == "--help" || a == "-h" {
			fmt.Println("netductor install [--component name ...]   default: core stack")
			fmt.Println("components: dirs singbox blocky vpn-users api metrics telegram")
			return
		}
		if a == "--component" || a == "-c" {
			continue
		}
		if strings.HasPrefix(a, "-") {
			continue
		}
		comps = append(comps, a)
	}
	// allow: install singbox blocky
	if err := install.Run(install.Options{Components: comps}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}


func runProbe(args []string) {
	cfg := probes.Load()
	results := probes.Run(cfg)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(map[string]any{"probes": results})
	fail := 0
	for _, r := range results {
		if ok, _ := r["ok"].(bool); !ok {
			fail++
		}
	}
	if fail > 0 {
		os.Exit(1)
	}
}



