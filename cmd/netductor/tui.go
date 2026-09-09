package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/PavelNeyman/netductor/internal/vpn"
)

const (
	cReset  = "\033[0m"
	cBold   = "\033[1m"
	cDim    = "\033[2m"
	cCyan   = "\033[36m"
	cGreen  = "\033[32m"
	cYellow = "\033[33m"
	cRed    = "\033[31m"
	cWhite  = "\033[97m"
)

type runMode string

const (
	modeVPS         runMode = "vps"
	modeOpenWRT     runMode = "openwrt"
	modeWorkstation runMode = "workstation"
	modeOperator    runMode = "operator"
)

var tuiMode runMode

func colorOK() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	fi, err := os.Stdout.Stat()
	return err == nil && (fi.Mode()&os.ModeCharDevice) != 0
}

func paint(code, s string) string {
	if !colorOK() {
		return s
	}
	return code + s + cReset
}

func box(title string) {
	width := 54
	inner := width - 2
	top := "╭" + strings.Repeat("─", inner) + "╮"
	bot := "╰" + strings.Repeat("─", inner) + "╯"
	t := " " + title + " "
	r := []rune(t)
	if len(r) > inner {
		t = string(r[:inner])
		r = []rune(t)
	}
	mid := "│" + t + strings.Repeat(" ", inner-len(r)) + "│"
	fmt.Println(paint(cCyan+cBold, top))
	fmt.Println(paint(cCyan, mid))
	fmt.Println(paint(cCyan+cBold, bot))
}

func item(n, title, desc string) {
	fmt.Printf("  %s  %s\n", paint(cBold+cWhite, n+")"), paint(cBold, title))
	if desc != "" {
		fmt.Printf("      %s\n", paint(cDim, desc))
	}
}

func prompt(in *bufio.Reader, label string) string {
	fmt.Printf("%s ", paint(cGreen+cBold, label))
	line, _ := in.ReadString('\n')
	return strings.TrimSpace(line)
}

func pause(in *bufio.Reader) {
	fmt.Print(paint(cDim, "\n  [Enter] … "))
	_, _ = in.ReadString('\n')
}

func confirm(in *bufio.Reader, q string) bool {
	a := prompt(in, q+" [y/N]:")
	return strings.HasPrefix(strings.ToLower(a), "y")
}

func detectSuggestedMode() (runMode, string) {
	if _, err := os.Stat("/etc/openwrt_release"); err == nil {
		return modeOpenWRT, "found /etc/openwrt_release"
	}
	if _, err := os.Stat("/etc/config/network"); err == nil {
		if _, err2 := os.Stat("/sbin/uci"); err2 == nil {
			return modeOpenWRT, "UCI + /etc/config/network"
		}
	}
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		if os.Geteuid() == 0 && os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
			return modeVPS, "Debian, root, no GUI"
		}
		if os.Geteuid() == 0 && os.Getenv("SSH_CONNECTION") != "" {
			return modeVPS, "Debian root over SSH"
		}
	}
	if runtime.GOOS == "darwin" {
		return modeWorkstation, "macOS"
	}
	if os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != "" {
		return modeWorkstation, "GUI session"
	}
	return modeOperator, "no strong signal"
}

func describeMode(m runMode) (string, string) {
	switch m {
	case modeVPS:
		return "VPS setup", "Install & configure the stack on a server (VPN, DNS, API, bot…)"
	case modeOpenWRT:
		return "OpenWrt / edge", "Router agent, site network, tunnel toward your VPS"
	case modeWorkstation:
		return "Workstation (PC/Mac)", "Build binaries, bootstrap hints, remote operator helpers"
	default:
		return "Operator panel", "Day-2 ops only: users, sessions, edge, doctor, probes"
	}
}

func parseModeFlags(args []string) runMode {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--mode" && i+1 < len(args) {
			return runMode(args[i+1])
		}
		if strings.HasPrefix(a, "--mode=") {
			return runMode(strings.TrimPrefix(a, "--mode="))
		}
	}
	if v := os.Getenv("NETDUCTOR_MODE"); v != "" {
		return runMode(v)
	}
	return ""
}

func chooseMode(in *bufio.Reader, forced runMode) runMode {
	if forced == modeVPS || forced == modeOpenWRT || forced == modeWorkstation || forced == modeOperator {
		t, _ := describeMode(forced)
		fmt.Println(paint(cDim, "  mode forced: "+t))
		return forced
	}
	sug, why := detectSuggestedMode()
	st, _ := describeMode(sug)

	fmt.Println()
	box("Netductor — select mode")
	fmt.Println()
	fmt.Printf("  %s %s\n", paint(cYellow, "Suggested:"), paint(cBold, st))
	fmt.Printf("  %s %s\n\n", paint(cDim, "Why:"), paint(cDim, why))

	modes := []runMode{modeVPS, modeOpenWRT, modeWorkstation, modeOperator}
	def := "1"
	for i, m := range modes {
		t, d := describeMode(m)
		mark := ""
		if m == sug {
			mark = " " + paint(cYellow, "← suggested")
			def = strconv.Itoa(i + 1)
		}
		fmt.Printf("  %s  %s%s\n      %s\n",
			paint(cBold+cWhite, fmt.Sprintf("%d)", i+1)),
			paint(cBold, t), mark,
			paint(cDim, d),
		)
	}
	fmt.Printf("\n  %s\n\n", paint(cDim, "CLI: netductor tui --mode vps|openwrt|workstation|operator"))

	ans := prompt(in, fmt.Sprintf("Mode [%s]:", def))
	if ans == "" {
		ans = def
	}
	n, err := strconv.Atoi(ans)
	if err != nil || n < 1 || n > len(modes) {
		fmt.Println(paint(cRed, "  invalid — using suggestion"))
		return sug
	}
	chosen := modes[n-1]
	ct, _ := describeMode(chosen)
	if !confirm(in, "Start «"+ct+"»?") {
		fmt.Println(paint(cDim, "  cancelled"))
		os.Exit(0)
	}
	return chosen
}

func runTUI(args []string) {
	forced := parseModeFlags(args)
	in := bufio.NewReader(os.Stdin)
	tuiMode = chooseMode(in, forced)
	for {
		var cont bool
		switch tuiMode {
		case modeVPS:
			cont = menuVPS(in)
		case modeOpenWRT:
			cont = menuOpenWRT(in)
		case modeWorkstation:
			cont = menuWorkstation(in)
		default:
			cont = menuOperator(in)
		}
		if !cont {
			fmt.Println(paint(cDim, "bye"))
			return
		}
	}
}

func menuHeader(m runMode) {
	t, _ := describeMode(m)
	fmt.Println()
	box("Netductor · " + t)
	fmt.Println()
}

func menuOperator(in *bufio.Reader) bool {
	menuHeader(modeOperator)
	item("1", "Status", "systemd units")
	item("2", "Doctor", "health checks")
	item("3", "VPN — list users", "")
	item("4", "VPN — add user", "")
	item("5", "Session token", "admin API Bearer")
	item("6", "Edge — list devices", "")
	item("7", "Live probes", "")
	item("8", "Collect metrics", "")
	item("9", "Change mode…", "")
	item("0", "Quit", "")
	fmt.Println()
	switch prompt(in, ">") {
	case "1":
		runStatus()
		pause(in)
	case "2":
		_ = runDoctorNative()
		pause(in)
	case "3":
		listUsersTUI()
		pause(in)
	case "4":
		addUserTUI(in)
		pause(in)
	case "5":
		sessionTUI(in)
		pause(in)
	case "6":
		runEdgeList()
		pause(in)
	case "7":
		runProbe(nil)
		pause(in)
	case "8":
		_ = runCollect()
		pause(in)
	case "9":
		tuiMode = chooseMode(in, "")
	case "0", "q":
		return false
	}
	return true
}

func menuVPS(in *bufio.Reader) bool {
	menuHeader(modeVPS)
	item("1", "Full install / upgrade", "bash install.sh (needs root + repo)")
	item("2", "Prepare only", "install.sh --prepare")
	item("3", "Doctor", "")
	item("4", "Status", "")
	item("5", "Operator tools…", "users, edge, probes")
	item("9", "Change mode…", "")
	item("0", "Quit", "")
	fmt.Println()
	switch prompt(in, ">") {
	case "1":
		fmt.Println(paint(cYellow, "  → bash install.sh"))
		if confirm(in, "Run install?") {
			runInstall(nil)
		}
		pause(in)
	case "2":
		fmt.Println(paint(cYellow, "  → bash install.sh --prepare"))
		if confirm(in, "Run prepare?") {
			runInstall([]string{"--prepare"})
		}
		pause(in)
	case "3":
		_ = runDoctorNative()
		pause(in)
	case "4":
		runStatus()
		pause(in)
	case "5":
		tuiMode = modeOperator
	case "9":
		tuiMode = chooseMode(in, "")
	case "0", "q":
		return false
	}
	return true
}

func menuOpenWRT(in *bufio.Reader) bool {
	menuHeader(modeOpenWRT)
	item("1", "Agent install instructions", "outbound netductor-agent")
	item("2", "Run install-openwrt.sh", "if present on device")
	item("3", "Check agent config", "")
	item("9", "Change mode…", "")
	item("0", "Quit", "")
	fmt.Println()
	switch prompt(in, ">") {
	case "1":
		fmt.Println(paint(cCyan, `
  mkdir -p /etc/netductor-agent
  # SERVER= TOKEN= DEVICE_ID= INTERVAL=60  → /etc/netductor-agent/config
  # binary: releases …/netductor-agent-linux-arm64|mipsle|arm
  # docs: edge/openwrt/INSTALL.md`))
		pause(in)
	case "2":
		if !confirm(in, "Run OpenWrt installer scripts?") {
			return true
		}
		for _, c := range []string{"install-openwrt.sh", "/opt/freshvps/install-openwrt.sh", "/opt/netductor/install-openwrt.sh"} {
			if st, err := os.Stat(c); err == nil && !st.IsDir() {
				cmd := exec.Command("sh", c)
				cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
				_ = cmd.Run()
				pause(in)
				return true
			}
		}
		fmt.Println(paint(cRed, "  install-openwrt.sh not found"))
		pause(in)
	case "3":
		for _, p := range []string{"/etc/netductor-agent/config", "/etc/freshvps-agent/config"} {
			if _, err := os.Stat(p); err == nil {
				fmt.Println(paint(cGreen, "  "+p))
			}
		}
		pause(in)
	case "9":
		tuiMode = chooseMode(in, "")
	case "0", "q":
		return false
	}
	return true
}

func menuWorkstation(in *bufio.Reader) bool {
	menuHeader(modeWorkstation)
	item("1", "Bootstrap one-liner", "copy to VPS")
	item("2", "Build netductor (local Go)", "")
	item("3", "Operator tools…", "via SSH tunnel")
	item("9", "Change mode…", "")
	item("0", "Quit", "")
	fmt.Println()
	switch prompt(in, ">") {
	case "1":
		fmt.Println(`
  curl -fsSL -o /usr/local/bin/netductor \
    https://github.com/PavelNeyman/netductor/releases/download/v0.7.0-dev/netductor-linux-amd64
  chmod 755 /usr/local/bin/netductor
  netductor tui --mode vps
`)
		pause(in)
	case "2":
		if confirm(in, "go build ./cmd/netductor ?") {
			cmd := exec.Command("go", "build", "-o", "netductor", "./cmd/netductor")
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
			_ = cmd.Run()
		}
		pause(in)
	case "3":
		tuiMode = modeOperator
	case "9":
		tuiMode = chooseMode(in, "")
	case "0", "q":
		return false
	}
	return true
}

func listUsersTUI() {
	users, err := vpn.List()
	if err != nil {
		fmt.Println(paint(cRed, err.Error()))
		return
	}
	if len(users) == 0 {
		fmt.Println(paint(cDim, "  (no users)"))
		return
	}
	for _, u := range users {
		en := paint(cRed, "off")
		if u.Enabled {
			en = paint(cGreen, "on")
		}
		fmt.Printf("  %s  [%s]  %s  %s\n", paint(cBold, u.Name), en, paint(cDim, u.UUID), u.Note)
	}
}

func addUserTUI(in *bufio.Reader) {
	name := prompt(in, "Name:")
	if name == "" {
		return
	}
	note := prompt(in, "Note:")
	out, err := vpn.Add(name, note)
	fmt.Println(out)
	if err != nil {
		fmt.Println(paint(cRed, err.Error()))
	}
}

func sessionTUI(in *bufio.Reader) {
	h := prompt(in, "Hours [72]:")
	hours := 72
	if h != "" {
		if n, err := strconv.Atoi(h); err == nil {
			hours = n
		}
	}
	tok, exp, err := vpn.CreateSession(hours)
	if err != nil {
		fmt.Println(paint(cRed, err.Error()))
		return
	}
	fmt.Println(paint(cGreen, tok))
	fmt.Fprintf(os.Stderr, "expires_unix=%d hours=%d\n", exp, hours)
}
