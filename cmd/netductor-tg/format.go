package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	qrcode "github.com/skip2/go-qrcode"

	"github.com/PavelNeyman/netductor/internal/addons"
)

func yn(ok bool, good, bad string) string {
	if ok {
		return good
	}
	return bad
}

func formatLampacHTML() string {
	st := addons.CollectLampac()
	ru := getLang() != "en"
	var b strings.Builder
	b.WriteString("📺 <b>Lampac</b>\n\n")
	if !st.Installed {
		if ru {
			b.WriteString("⚪ Не установлен\n")
			b.WriteString("<i>Установка: <code>netductor install lampac</code></i>")
		} else {
			b.WriteString("⚪ Not installed\n")
			b.WriteString("<i>Install: <code>netductor install lampac</code></i>")
		}
		return b.String()
	}
	runIcon := yn(st.Running, "🟢", "🔴")
	healthIcon := yn(st.Healthy, "✅", "⚠️")
	if ru {
		b.WriteString(fmt.Sprintf("%s <b>Контейнер:</b> %s\n", runIcon, yn(st.Running, "запущен", "остановлен")))
		b.WriteString(fmt.Sprintf("%s <b>Health:</b> %s\n", healthIcon, yn(st.Healthy, "healthy", "unhealthy")))
	} else {
		b.WriteString(fmt.Sprintf("%s <b>Container:</b> %s\n", runIcon, yn(st.Running, "running", "stopped")))
		b.WriteString(fmt.Sprintf("%s <b>Health:</b> %s\n", healthIcon, yn(st.Healthy, "healthy", "unhealthy")))
	}
	if st.Image != "" {
		b.WriteString(fmt.Sprintf("📦 <b>Image:</b> <code>%s</code>\n", esc(st.Image)))
	}
	b.WriteString(fmt.Sprintf("🔌 <b>Bind:</b> <code>%s</code>\n", esc(st.Bind)))
	if st.VersionHash != "" {
		b.WriteString(fmt.Sprintf("🏷 <b>Version:</b> <code>%s</code>\n", esc(st.VersionHash)))
	}
	if st.CPU != "" || st.Mem != "" {
		b.WriteString(fmt.Sprintf("📊 <b>CPU / RAM:</b> %s · %s\n", esc(st.CPU), esc(st.Mem)))
	}
	b.WriteString(fmt.Sprintf("🏓 <b>Ping:</b> %s\n", yn(st.PingOK, "✅", "—")))
	b.WriteString(fmt.Sprintf("🧭 <b>Chromium:</b> %s\n", yn(st.ChromiumOK, "✅", "—")))
	b.WriteByte('\n')
	if ru {
		b.WriteString("🔗 UI: <code>" + esc(st.UIURL) + "</code>\n")
		b.WriteString("🛠 Admin: <code>" + esc(st.AdminURL) + "</code>\n")
		b.WriteString("\n<i>Только localhost — доступ через VPN/SSH-туннель</i>")
	} else {
		b.WriteString("🔗 UI: <code>" + esc(st.UIURL) + "</code>\n")
		b.WriteString("🛠 Admin: <code>" + esc(st.AdminURL) + "</code>\n")
		b.WriteString("\n<i>Loopback only — reach via VPN/SSH tunnel</i>")
	}
	return b.String()
}

func formatAddonsHTML() string {
	ru := getLang() != "en"
	var b strings.Builder
	if ru {
		b.WriteString("🧩 <b>Аддоны</b>\n\nВыберите компонент:")
	} else {
		b.WriteString("🧩 <b>Addons</b>\n\nPick a component:")
	}
	return b.String()
}

// formatEdgeListHTML turns tab-separated edge list into cards.
func formatEdgeListHTML(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if getLang() != "en" {
			return "📡 <i>Нет устройств</i>"
		}
		return "📡 <i>No devices</i>"
	}
	var b strings.Builder
	if getLang() != "en" {
		b.WriteString("📡 <b>Роутеры / edge</b>\n\n")
	} else {
		b.WriteString("📡 <b>Routers / edge</b>\n\n")
	}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		// id status online board host ip=… last=
		id := parts[0]
		rest := ""
		if len(parts) > 1 {
			rest = strings.Join(parts[1:], " · ")
		}
		icon := "⚪"
		low := strings.ToLower(line)
		if strings.Contains(low, "online") {
			icon = "🟢"
		} else if strings.Contains(low, "offline") {
			icon = "🔴"
		} else if strings.Contains(low, "pending") {
			icon = "⏳"
		}
		b.WriteString(fmt.Sprintf("%s <code>%s</code>\n   %s\n", icon, esc(id), esc(rest)))
	}
	return b.String()
}

func formatPendingHTML(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if getLang() != "en" {
			return "⏳ <b>Ожидают approve</b>\n\n<i>Список пуст</i>"
		}
		return "⏳ <b>Pending</b>\n\n<i>Empty</i>"
	}
	var b strings.Builder
	if getLang() != "en" {
		b.WriteString("⏳ <b>Ожидают approve</b>\n\n")
	} else {
		b.WriteString("⏳ <b>Pending enroll</b>\n\n")
	}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		id := parts[0]
		extra := ""
		if len(parts) > 1 {
			extra = strings.Join(parts[1:], " · ")
		}
		b.WriteString(fmt.Sprintf("• <code>%s</code>", esc(id)))
		if extra != "" {
			b.WriteString(" — " + esc(extra))
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func formatTemplatesHTML(raw string) string {
	raw = strings.TrimSpace(raw)
	var b strings.Builder
	if getLang() != "en" {
		b.WriteString("📋 <b>Шаблоны</b>\n\n")
	} else {
		b.WriteString("📋 <b>Templates</b>\n\n")
	}
	if raw == "" {
		b.WriteString("<i>—</i>")
		return b.String()
	}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		b.WriteString("• <code>" + esc(line) + "</code>\n")
	}
	return b.String()
}

func formatSessionHTML(raw string) string {
	raw = strings.TrimSpace(raw)
	// expect token line(s) from vpn session
	var token string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if len(line) >= 32 && isHexish(line) {
			token = line
			break
		}
	}
	var b strings.Builder
	if getLang() != "en" {
		b.WriteString("🔑 <b>Session token</b>\n\n")
	} else {
		b.WriteString("🔑 <b>Session token</b>\n\n")
	}
	if token != "" {
		b.WriteString("<code>" + esc(token) + "</code>\n\n")
		if getLang() != "en" {
			b.WriteString("<i>Для Admin SPA / API (Authorization: Bearer …)</i>")
		} else {
			b.WriteString("<i>For Admin SPA / API (Authorization: Bearer …)</i>")
		}
	} else {
		b.WriteString("<pre>" + esc(raw) + "</pre>")
	}
	return b.String()
}

func isHexish(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func formatDoctorHTML(raw string) string {
	raw = strings.TrimSpace(raw)
	var b strings.Builder
	b.WriteString("🩺 <b>Doctor</b>\n\n")
	ok, fail, warn := 0, 0, 0
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "OK"):
			ok++
			b.WriteString("✅ " + esc(strings.TrimSpace(strings.TrimPrefix(line, "OK"))) + "\n")
		case strings.HasPrefix(line, "FAIL"):
			fail++
			b.WriteString("❌ " + esc(strings.TrimSpace(strings.TrimPrefix(line, "FAIL"))) + "\n")
		case strings.HasPrefix(line, "WARN"):
			warn++
			b.WriteString("⚠️ " + esc(strings.TrimSpace(strings.TrimPrefix(line, "WARN"))) + "\n")
		case strings.HasPrefix(line, "INFO"):
			b.WriteString("ℹ️ " + esc(strings.TrimSpace(strings.TrimPrefix(line, "INFO"))) + "\n")
		case strings.HasPrefix(line, "Summary"):
			b.WriteString("\n<b>" + esc(line) + "</b>\n")
		default:
			if strings.HasPrefix(line, "Netductor doctor") {
				b.WriteString("<b>" + esc(line) + "</b>\n")
			}
		}
	}
	_ = ok
	_ = fail
	_ = warn
	return b.String()
}


func formatVPNListPretty(raw string) string {
	nl := string([]byte{10})
	var b strings.Builder
	if getLang() != "en" {
		b.WriteString("👥 <b>VPN пользователи</b>" + nl + nl)
	} else {
		b.WriteString("👥 <b>VPN users</b>" + nl + nl)
	}
	n := 0
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 1 {
			parts = strings.Fields(line)
		}
		name := parts[0]
		if name == "relay-uplink" {
			continue
		}
		en := ""
		if len(parts) > 1 {
			en = parts[1]
		}
		n++
		icon := "🟢"
		if en == "off" {
			icon = "🔴"
		}
		b.WriteString(fmt.Sprintf("%s <b>%d. %s</b>", icon, n, esc(name)))
		if en != "" {
			b.WriteString(" · <code>" + esc(en) + "</code>")
		}
		b.WriteString(nl)
	}
	if n == 0 {
		b.WriteString("<i>—</i>")
	}
	return strings.TrimRight(b.String(), nl)
}

func formatStatusPretty() string {
	return statusInline()
}

func formatVPNLinkHTML(name string) (caption string, vless, hy2, sub string) {
	vless = strings.TrimSpace(runVPN("link", name, "vless"))
	hy2 = strings.TrimSpace(runVPN("link", name, "hy2"))
	sub = strings.TrimSpace(runVPN("link", name))
	if sub != "" {
		first := strings.TrimSpace(strings.Split(sub, "\n")[0])
		if strings.HasPrefix(first, "vless://") {
			vless = first
		}
	}
	if strings.Contains(vless, "not found") || strings.Contains(vless, "exit status") {
		vless = ""
	}
	if strings.Contains(hy2, "not found") || strings.Contains(hy2, "exit status") {
		hy2 = ""
	}
	if strings.Contains(sub, "not found") || strings.Contains(sub, "exit status") {
		sub = ""
	}
	nl := string([]byte{10})
	var b strings.Builder
	b.WriteString("🔗 <b>VPN · " + esc(name) + "</b>" + nl + nl)
	if vless != "" {
		b.WriteString("<b>VLESS</b>" + nl + "<code>" + esc(vless) + "</code>" + nl + nl)
	}
	if hy2 != "" {
		b.WriteString("<b>HY2</b>" + nl + "<code>" + esc(hy2) + "</code>" + nl)
	}
	if vless == "" && hy2 == "" {
		b.WriteString("❌ no links")
	}
	return b.String(), vless, hy2, sub
}

func formatRelayListHTML() string {
	_ = runND("relay", "status")
	ex := strings.TrimSpace(runND("relay", "exit"))
	nl := string([]byte{10})
	body := formatNodesListHTML()
	return "📡 <b>Nodes / relay</b>" + nl + "RU exit: <code>" + esc(ex) + "</code>" + nl + nl + body
}

func formatNodeDetailHTML(id string) string {
	nl := string([]byte{10})
	out := runND("nodes", "list")
	var line string
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, "id="+id) || strings.HasPrefix(strings.TrimSpace(l), "id="+id) {
			line = l
			break
		}
	}
	// relay metrics from status
	st := runND("relay", "status")
	var b strings.Builder
	b.WriteString("🖥 <b>Node</b> <code>" + esc(id) + "</code>" + nl + nl)
	if line != "" {
		for _, p := range strings.Split(line, "\t") {
			b.WriteString("• <code>" + esc(p) + "</code>" + nl)
		}
	}
	if st != "" {
		b.WriteString(nl + "<b>relay status</b>" + nl)
		for _, l := range strings.Split(st, "\n") {
			if strings.Contains(l, id) {
				b.WriteString("<code>" + esc(l) + "</code>" + nl)
			}
		}
	}
	b.WriteString(nl + "<i>Reboot/Upgrade: queued for agent (relay) or local (core).</i>")
	return b.String()
}

func enqueueNodeCmd(id, cmd string) string {
	// relay queue
	out := runND("relay", "cmd", id, cmd)
	if out != "" && !strings.Contains(out, "unknown") && !strings.Contains(out, "usage") {
		return out
	}
	// core local
	if cmd == "reboot" {
		return runND("nodes", "local-cmd", "reboot")
	}
	if cmd == "upgrade" {
		return runND("nodes", "local-cmd", "upgrade")
	}
	return out
}

func ensureQRFile(path, payload string) string {
	if payload == "" {
		return ""
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o700)
	// Always regenerate — cached PNG may still encode core IP while link is relay.
	if err := qrcode.WriteFile(payload, qrcode.Medium, 512, path); err != nil {
		return ""
	}
	_ = os.Chmod(path, 0o600)
	return path
}

func deliverVPNLink(token string, chat int64, msgID int, name string) {
	// Replace the message that contained the user button (list / card).
	showVPNQR(token, chat, msgID, name, "vless", msgID > 0)
}

func vpnQRCaption(name, mode, vless, hy2 string) string {
	ru := getLang() != "en"
	nl := string([]byte{10})
	var b strings.Builder
	if mode == "hy2" {
		b.WriteString("📱 <b>Hysteria2</b> · " + esc(name) + nl + nl)
		if hy2 != "" {
			b.WriteString("<code>" + esc(hy2) + "</code>")
		}
	} else {
		b.WriteString("📱 <b>VLESS Reality</b> · " + esc(name) + nl + nl)
		if vless != "" {
			b.WriteString("<code>" + esc(vless) + "</code>")
		}
		if hy2 != "" {
			if ru {
				b.WriteString(nl + nl + "<i>HY2 — кнопка «HY2 QR» выше</i>")
			} else {
				b.WriteString(nl + nl + "<i>HY2 — use «HY2 QR» button</i>")
			}
		}
	}
	return b.String()
}

func showVPNQR(token string, chat int64, msgID int, name, mode string, edit bool) {
	_, vless, hy2, sub := formatVPNLinkHTML(name)
	if mode != "hy2" {
		mode = "vless"
	}
	kb := userCardKeyboardMode(name, mode)
	dir := filepath.Join("/etc/netductor/clients", name)
	var path, payload string
	if mode == "hy2" {
		payload = hy2
		path = ensureQRFile(filepath.Join(dir, "qr-hy2.png"), hy2)
	} else {
		payload = vless
		if payload == "" {
			payload = sub
		}
		path = ensureQRFile(filepath.Join(dir, "qr-vless.png"), vless)
		if path == "" {
			path = ensureQRFile(filepath.Join(dir, "qr.png"), vless)
		}
	}
	cap := vpnQRCaption(name, mode, vless, hy2)
	if payload == "" && path == "" {
		sendHTML(token, chat, cap, kb)
		return
	}
	if path == "" {
		sendHTML(token, chat, cap, kb)
		return
	}
	if edit && msgID > 0 {
		if err := editPhotoFile(token, chat, msgID, path, cap, kb); err != nil {
			// Text list message cannot become a photo — delete and put QR in its place.
			_ = deleteMessage(token, chat, msgID)
			if err2 := sendPhotoFile(token, chat, path, cap, kb); err2 != nil {
				sendHTML(token, chat, cap+string([]byte{10})+"⚠️ <code>"+esc(err2.Error())+"</code>", kb)
			}
		}
		return
	}
	if err := sendPhotoFile(token, chat, path, cap, kb); err != nil {
		sendHTML(token, chat, cap+string([]byte{10})+"⚠️ <code>"+esc(err.Error())+"</code>", kb)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func formatRelayHelp() string {
	nl := string([]byte{10})
	if getLang() != "en" {
		return "Управление промежуточным VPS в РФ." + nl + nl + "1) <b>Export</b> / <b>One-liner</b>" + nl + "2) На RU выполнить команду" + nl + "3) Мобильным — relay links"
	}
	return "RU intermediate VPS." + nl + nl + "1) Export / One-liner" + nl + "2) Run on RU" + nl + "3) Mobile uses relay links"
}

func formatRelayOneline() string {
	_ = runND("relay", "export", "-o", "/tmp/nd-relay-bundle.json", "--sni", "ya.ru")
	b, err := os.ReadFile("/tmp/nd-relay-bundle.json")
	if err != nil {
		return "❌ export failed: " + esc(err.Error())
	}
	enc := base64.StdEncoding.EncodeToString(b)
	cmd := "wget -qO /usr/local/bin/netductor https://github.com/PavelNeyman/netductor/releases/download/v0.7.0-dev/netductor-linux-amd64 && chmod 755 /usr/local/bin/netductor && echo " + enc + " | base64 -d > /root/bundle.json && netductor relay join /root/bundle.json"
	nl := string([]byte{10})
	if getLang() != "en" {
		return "🧾 <b>Одна команда на RU VPS</b>" + nl + nl + "<code>" + esc(cmd) + "</code>"
	}
	return "🧾 <b>One command on RU VPS</b>" + nl + nl + "<code>" + esc(cmd) + "</code>"
}

