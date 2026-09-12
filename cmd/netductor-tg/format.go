package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
	// name\ton\tuuid\tnote\tcreated
	raw = strings.TrimSpace(raw)
	var b strings.Builder
	if getLang() != "en" {
		b.WriteString("👥 <b>VPN пользователи</b>\n\n")
	} else {
		b.WriteString("👥 <b>VPN users</b>\n\n")
	}
	if raw == "" {
		b.WriteString("<i>—</i>")
		return b.String()
	}
	n := 0
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			parts = strings.Fields(line)
		}
		if len(parts) < 2 {
			continue
		}
		n++
		name, en := parts[0], parts[1]
		icon := yn(en == "on", "🟢", "🔴")
		note := ""
		if len(parts) >= 4 {
			note = parts[3]
		}
		b.WriteString(fmt.Sprintf("%s <b>%s</b>", icon, esc(name)))
		if note != "" {
			b.WriteString(" — <i>" + esc(note) + "</i>")
		}
		b.WriteByte('\n')
	}
	if n == 0 {
		b.WriteString("<pre>" + esc(raw) + "</pre>")
	}
	return b.String()
}

// formatJSONPretty is last resort for unknown JSON blobs.
func formatJSONPretty(title, raw string) string {
	raw = strings.TrimSpace(raw)
	var v any
	if json.Unmarshal([]byte(raw), &v) == nil {
		b, _ := json.MarshalIndent(v, "", "  ")
		return title + "\n<pre>" + esc(string(b)) + "</pre>"
	}
	return title + "\n<pre>" + esc(raw) + "</pre>"
}

func formatStatusPretty() string {
	// Prefer structured inline HTML over raw shell dump.
	return statusInline()
}

func formatVPNLinkHTML(name string) (caption string, vless, hy2, sub string) {
	vless = strings.TrimSpace(runVPN("link", name, "vless"))
	hy2 = strings.TrimSpace(runVPN("link", name, "hy2"))
	sub = strings.TrimSpace(runVPN("link", name))
	// strip errors
	if strings.Contains(vless, "not found") || strings.Contains(vless, "exit status") {
		vless = ""
	}
	if strings.Contains(hy2, "not found") || strings.Contains(hy2, "exit status") {
		hy2 = ""
	}
	if strings.Contains(sub, "not found") || strings.Contains(sub, "exit status") {
		sub = ""
	}
	ru := getLang() != "en"
	var b strings.Builder
	if ru {
		b.WriteString("🔗 <b>VPN · " + esc(name) + "</b>\n\n")
	} else {
		b.WriteString("🔗 <b>VPN · " + esc(name) + "</b>\n\n")
	}
	if vless != "" {
		b.WriteString("<b>VLESS Reality</b>\n<code>" + esc(vless) + "</code>\n\n")
	}
	if hy2 != "" {
		b.WriteString("<b>Hysteria2</b>\n<code>" + esc(hy2) + "</code>\n\n")
	}
	if vless == "" && hy2 == "" && sub != "" {
		b.WriteString("<code>" + esc(sub) + "</code>\n\n")
	}
	if vless == "" && hy2 == "" && sub == "" {
		if ru {
			b.WriteString("❌ Пользователь не найден или нет ссылок.\nПроверьте: VPN → список.")
		} else {
			b.WriteString("❌ User not found or no links.\nCheck: VPN → list.")
		}
	} else if ru {
		b.WriteString("<i>QR ниже (VLESS). Скопируйте ссылку длинным нажатием на mono-текст.</i>")
	} else {
		b.WriteString("<i>QR below (VLESS). Long-press mono text to copy.</i>")
	}
	return b.String(), vless, hy2, sub
}

func vpnQRPath(name string) string {
	candidates := []string{
		"/etc/netductor/clients/" + name + "/qr.png",
		"/etc/netductor/clients/" + name + "/qr-subscription.png",
		"/var/lib/netductor/clients/" + name + "/qr.png",
	}
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}


func deliverVPNLink(token string, chat int64, replyTo int, name string) {
	caption, _, _, _ := formatVPNLinkHTML(name)
	kb := userCardKeyboard(name, "")
	qr := vpnQRPath(name)
	if qr != "" {
		sendPhotoFile(token, chat, qr, caption, kb)
		return
	}
	if replyTo > 0 {
		reply(token, chat, replyTo, caption, kb)
	} else {
		sendHTML(token, chat, caption, kb)
	}
}
