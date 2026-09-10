package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/PavelNeyman/netductor/internal/vpn"
)

// ── modes ──────────────────────────────────────────────────

type runMode string

const (
	modeVPS         runMode = "vps"
	modeOpenWRT     runMode = "openwrt"
	modeWorkstation runMode = "workstation"
	modeOperator    runMode = "operator"
)

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
		return "VPS setup", "Install & configure stack on a server (VPN, DNS, API, bot…)"
	case modeOpenWRT:
		return "OpenWrt / edge", "Router agent, site network, tunnel toward your VPS"
	case modeWorkstation:
		return "Workstation (PC/Mac)", "Build binaries, bootstrap hints, remote helpers"
	default:
		return "Operator panel", "Day-2 ops: users, sessions, edge, doctor, probes"
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

// ── styles ─────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginLeft(1)
	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginLeft(1)
	docStyle = lipgloss.NewStyle().Margin(1, 2)
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	okStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

// ── list items ─────────────────────────────────────────────

type menuItem struct {
	title, desc, id string
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.desc }
func (i menuItem) FilterValue() string { return i.title }

// ── model ──────────────────────────────────────────────────

type screen int

const (
	screenMode screen = iota
	screenMenu
	screenOutput
)

type model struct {
	screen   screen
	mode     runMode
	list     list.Model
	output   string
	quitting bool
	width    int
	height   int
}

func modeItems(sug runMode) []list.Item {
	order := []runMode{modeVPS, modeOpenWRT, modeWorkstation, modeOperator}
	items := make([]list.Item, 0, 4)
	for _, m := range order {
		t, d := describeMode(m)
		if m == sug {
			t = t + "  ← suggested"
		}
		items = append(items, menuItem{title: t, desc: d, id: string(m)})
	}
	return items
}

func menuItemsFor(mode runMode) []list.Item {
	switch mode {
	case modeVPS:
		return []list.Item{
			menuItem{"Full install / upgrade", "bash install.sh (root + repo)", "install"},
			menuItem{"Prepare only", "install.sh --prepare", "prepare"},
			menuItem{"Doctor", "health checks", "doctor"},
			menuItem{"Status", "systemd units", "status"},
			menuItem{"Operator tools…", "VPN, edge, probes", "to-operator"},
			menuItem{"Change mode…", "", "change-mode"},
			menuItem{"Quit", "", "quit"},
		}
	case modeOpenWRT:
		return []list.Item{
			menuItem{"Agent install instructions", "outbound netductor-agent", "agent-help"},
			menuItem{"Run install-openwrt.sh", "if present on device", "owrt-install"},
			menuItem{"Check agent config", "", "agent-cfg"},
			menuItem{"Change mode…", "", "change-mode"},
			menuItem{"Quit", "", "quit"},
		}
	case modeWorkstation:
		return []list.Item{
			menuItem{"Bootstrap one-liner", "copy-paste for VPS", "bootstrap"},
			menuItem{"Build netductor (local Go)", "", "build"},
			menuItem{"Operator tools…", "via SSH tunnel", "to-operator"},
			menuItem{"Change mode…", "", "change-mode"},
			menuItem{"Quit", "", "quit"},
		}
	default:
		return []list.Item{
			menuItem{"Status", "systemd units", "status"},
			menuItem{"Doctor", "health checks", "doctor"},
			menuItem{"VPN — list users", "", "vpn-list"},
			menuItem{"VPN — add user", "prompt name", "vpn-add"},
			menuItem{"Session token", "admin API Bearer", "session"},
			menuItem{"Edge — list devices", "", "edge-list"},
			menuItem{"Live probes", "", "probe"},
			menuItem{"Collect metrics", "", "collect"},
			menuItem{"Change mode…", "", "change-mode"},
			menuItem{"Quit", "", "quit"},
		}
	}
}

func newList(title string, items []list.Item, w, h int) list.Model {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(lipgloss.Color("205")).BorderForeground(lipgloss.Color("205"))
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.Foreground(lipgloss.Color("218"))
	l := list.New(items, d, w, h)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.SetShowHelp(true)
	return l
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-6)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			if m.screen == screenOutput {
				m.screen = screenMenu
				m.output = ""
				return m, nil
			}
			if m.screen == screenMenu {
				m.screen = screenMode
				sug, _ := detectSuggestedMode()
				m.list = newList("Select mode", modeItems(sug), m.width-4, m.height-6)
				return m, nil
			}
		case "enter":
			if m.screen == screenOutput {
				m.screen = screenMenu
				m.output = ""
				return m, nil
			}
			it, ok := m.list.SelectedItem().(menuItem)
			if !ok {
				return m, nil
			}
			if m.screen == screenMode {
				m.mode = runMode(it.id)
				t, _ := describeMode(m.mode)
				m.list = newList("Netductor · "+t, menuItemsFor(m.mode), m.width-4, m.height-6)
				m.screen = screenMenu
				return m, nil
			}
			// screenMenu action
			return m.runAction(it.id)
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) runAction(id string) (tea.Model, tea.Cmd) {
	switch id {
	case "quit":
		m.quitting = true
		return m, tea.Quit
	case "change-mode":
		sug, _ := detectSuggestedMode()
		m.list = newList("Select mode", modeItems(sug), m.width-4, m.height-6)
		m.screen = screenMode
		return m, nil
	case "to-operator":
		m.mode = modeOperator
		t, _ := describeMode(m.mode)
		m.list = newList("Netductor · "+t, menuItemsFor(m.mode), m.width-4, m.height-6)
		return m, nil
	case "status":
		m.output = capture(func() { runStatus() })
		m.screen = screenOutput
		return m, nil
	case "doctor":
		m.output = capture(func() { _ = runDoctorNative() })
		m.screen = screenOutput
		return m, nil
	case "probe":
		m.output = capture(func() { runProbe(nil) })
		m.screen = screenOutput
		return m, nil
	case "collect":
		m.output = capture(func() { _ = runCollect() })
		m.screen = screenOutput
		return m, nil
	case "edge-list":
		m.output = capture(runEdgeList)
		m.screen = screenOutput
		return m, nil
	case "vpn-list":
		m.output = capture(func() {
			users, err := vpn.List()
			if err != nil {
				fmt.Println(err)
				return
			}
			if len(users) == 0 {
				fmt.Println("(no users)")
			}
			for _, u := range users {
				en := "off"
				if u.Enabled {
					en = "on"
				}
				fmt.Printf("%s\t%s\t%s\t%s\n", u.Name, en, u.UUID, u.Note)
			}
		})
		m.screen = screenOutput
		return m, nil
	case "vpn-add":
		m.output = "VPN add requires a name.\n\nUse:  netductor vpn add <name> [note]\n\nInteractive prompt will be added later."
		m.screen = screenOutput
		return m, nil
	case "session":
		tok, exp, err := vpn.CreateSession(72)
		if err != nil {
			m.output = err.Error()
		} else {
			m.output = fmt.Sprintf("%s\n\nexpires_unix=%d hours=72", tok, exp)
		}
		m.screen = screenOutput
		return m, nil
	case "install":
		m.output = "Will run: bash install.sh\nExit TUI and run:\n  netductor install\nor confirm from shell as root."
		m.screen = screenOutput
		return m, nil
	case "prepare":
		m.output = "Will run: bash install.sh --prepare\nUse: netductor install --prepare"
		m.screen = screenOutput
		return m, nil
	case "bootstrap":
		m.output = `curl -fsSL -o /usr/local/bin/netductor \
  https://github.com/PavelNeyman/netductor/releases/download/v0.7.0-dev/netductor-linux-amd64
chmod 755 /usr/local/bin/netductor
netductor tui --mode vps`
		m.screen = screenOutput
		return m, nil
	case "build":
		m.output = capture(func() {
			cmd := exec.Command("go", "build", "-o", "netductor", "./cmd/netductor")
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Println("error:", err)
			} else {
				fmt.Println("ok: ./netductor")
			}
		})
		m.screen = screenOutput
		return m, nil
	case "agent-help":
		m.output = `mkdir -p /etc/netductor-agent
# SERVER= TOKEN= DEVICE_ID= INTERVAL=60 → config
# binary: netductor-agent-linux-arm64|mipsle|arm from releases
# docs: edge/openwrt/INSTALL.md`
		m.screen = screenOutput
		return m, nil
	case "owrt-install":
		m.output = capture(func() {
			for _, c := range []string{"install-openwrt.sh", "/opt/freshvps/install-openwrt.sh", "/opt/netductor/install-openwrt.sh"} {
				if st, err := os.Stat(c); err == nil && !st.IsDir() {
					cmd := exec.Command("sh", c)
					cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
					_ = cmd.Run()
					return
				}
			}
			fmt.Println("install-openwrt.sh not found")
		})
		m.screen = screenOutput
		return m, nil
	case "agent-cfg":
		m.output = capture(func() {
			for _, p := range []string{"/etc/netductor-agent/config", "/etc/freshvps-agent/config"} {
				if _, err := os.Stat(p); err == nil {
					fmt.Println("found:", p)
				}
			}
		})
		m.screen = screenOutput
		return m, nil
	}
	return m, nil
}

func capture(fn func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		fn()
		return ""
	}
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = w, w
	fn()
	_ = w.Close()
	os.Stdout, os.Stderr = oldOut, oldErr
	var buf strings.Builder
	tmp := make([]byte, 4096)
	for {
		n, er := r.Read(tmp)
		if n > 0 {
			buf.Write(tmp[:n])
		}
		if er != nil {
			break
		}
	}
	_ = r.Close()
	return buf.String()
}

func (m model) View() string {
	if m.quitting {
		return subtitleStyle.Render("bye") + "\n"
	}
	if m.screen == screenOutput {
		body := lipgloss.NewStyle().Width(m.width - 4).Render(m.output)
		return docStyle.Render(
			titleStyle.Render("Output") + "\n\n" + body + "\n\n" +
				helpStyle.Render("enter/esc back · ctrl+c quit"),
		)
	}
	sug, why := detectSuggestedMode()
	header := ""
	if m.screen == screenMode {
		st, _ := describeMode(sug)
		header = subtitleStyle.Render(fmt.Sprintf("Suggested: %s (%s)", st, why)) + "\n" +
			subtitleStyle.Render("↑↓ select · enter confirm · nothing runs until you choose") + "\n\n"
	}
	return docStyle.Render(header + m.list.View())
}

// runTUI — Bubble Tea UI for VPS / PC / operator. Agent stays separate & light.
func runTUI(args []string) {
	forced := parseModeFlags(args)
	w, h := 80, 24
	sug, _ := detectSuggestedMode()

	var m model
	m.width, m.height = w, h
	if forced == modeVPS || forced == modeOpenWRT || forced == modeWorkstation || forced == modeOperator {
		m.mode = forced
		m.screen = screenMenu
		t, _ := describeMode(forced)
		m.list = newList("Netductor · "+t, menuItemsFor(forced), w-4, h-6)
	} else {
		m.screen = screenMode
		m.list = newList("Select mode", modeItems(sug), w-4, h-6)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
