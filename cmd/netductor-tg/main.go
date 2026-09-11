package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	tokenFile = "/etc/netductor/secrets/telegram_bot_token"
	chatFile  = "/etc/netductor/secrets/telegram_admin_id"
	langFile  = "/etc/netductor/telegram_lang"
	statusSh  = "/opt/netductor/runtime/telegram/status.sh"
)

type update struct {
	UpdateID      int            `json:"update_id"`
	Message       *message       `json:"message"`
	CallbackQuery *callbackQuery `json:"callback_query"`
}

type message struct {
	MessageID    int      `json:"message_id"`
	Text         string   `json:"text"`
	Chat         chat     `json:"chat"`
	From         *user    `json:"from"`
	ForwardFrom  *user    `json:"forward_from"`
	ForwardOrigin *struct {
		SenderUser *user `json:"sender_user"`
	} `json:"forward_origin"`
}

type callbackQuery struct {
	ID      string   `json:"id"`
	From    user     `json:"from"`
	Message *message `json:"message"`
	Data    string   `json:"data"`
}

type chat struct {
	ID int64 `json:"id"`
}

type user struct {
	ID int64 `json:"id"`
}

// --- i18n ---

var dict = map[string]map[string]string{
	"en": {
		"menu_title":    "🛡 <b>Netductor operator panel</b>",
		"menu_hint":     "Use buttons below. Slash commands still work.",
		"status":        "📊 Status",
		"ready":         "📄 READY",
		"vpn_list":      "👥 VPN list",
		"vpn_add":       "➕ Add user",
		"vpn_link":      "🔗 Link / QR",
		"vpn_disable":   "🚫 Disable",
		"vpn_enable":    "✅ Enable",
		"vpn_revoke":    "🗑 Revoke",
		"session":       "🔑 API session",
		"admin":         "🖥 Admin UI",
		"help":          "❓ Help",
		"routers":       "📡 Routers",
		"routers_title": "📡 <b>Routers</b>",
		"routers_empty": "No edge agents yet.",
		"lang":          "🌐 Language",
		"main_menu":     "🏠 Main menu",
		"status_title":  "📊 <b>Status</b>",
		"help_title":    "❓ <b>Help</b>",
		"help_body":     "Operator-only bot.\n• VPN users: add / link / enable / disable / revoke\n• Session token for Admin UI &amp; Shortcuts\n• Status &amp; READY\n\nCommands: /menu /status /vpn_list /lang /session /admin /help",
		"add_prompt":    "➕ <b>Add VPN user</b>\n\nSend a short <b>name</b> (e.g. <code>alice</code>).",
		"name_prompt":   "✏️ <b>%s</b>\n\nSend the VPN <b>user name</b>:",
		"session_prompt": "🔑 <b>API session</b>\n\nSend lifetime in <b>hours</b> (e.g. <code>72</code>), or /cancel:",
		"admin_body":    "🖥 <b>Admin UI</b>\n\nURL: <code>http://127.0.0.1:8787/admin/</code>\n\n1) SSH tunnel:\n<code>ssh -L 8787:127.0.0.1:8787 root@VPS</code>\n2) Open URL in browser\n3) Paste session token\n\nOr: <code>netductor vpn api-bind detect</code>",
		"unknown":       "Unknown action.",
		"unknown_cmd":   "Unknown. Open the menu:",
		"no_ready":      "⚠️ No READY.txt",
		"ready_title":   "📄 <b>READY</b>",
		"session_token": "🔑 Session token",
		"lang_now":      "🌐 Language: <b>%s</b>\n\nChoose:",
		"lang_en":       "English",
		"lang_ru":       "Русский",
		"lang_set":      "✅ Language set to <b>%s</b>",
		"copy_token":    "📋 Copy token",
		"show_links":    "🔗 Show links again",
		"cancelled":     "Cancelled.",
		"vpn_users":     "👥 <b>VPN users</b>",
		"label_link":    "Link / QR",
		"label_disable": "Disable",
		"label_enable":  "Enable",
		"label_revoke":  "Revoke",
		"cat_vpn":       "🔐 VPN",
		"cat_routers":   "📡 Routers",
		"cat_vpn_title": "🔐 <b>VPN</b>\nManage users, links, access.",
		"cat_routers_title": "📡 <b>Routers</b>\nEdge devices, templates, enroll.",
		"pending":       "⏳ Pending",
		"templates":     "📋 Templates",
		"bind_tmpl":     "🔗 Bind template",
		"apply_tmpl":    "⚙️ Apply template",
		"devices":       "📡 Devices",
		"back_vpn":      "⬅️ VPN",
		"back_routers":  "⬅️ Routers",
		"pending_title": "⏳ <b>Pending</b>",
		"pending_empty": "No pending devices.",
		"templates_title": "📋 <b>Templates</b>",
		"bind_prompt":   "Device id to bind template:",
		"apply_prompt":  "Device id to enqueue <code>apply_template</code>:",
		"menu_hint2":    "Choose a category — VPN or Routers open a submenu.",
	},
	"ru": {
		"menu_title":    "🛡 <b>Панель оператора Netductor</b>",
		"menu_hint":     "Кнопки ниже. Слеш-команды тоже работают.",
		"status":        "📊 Статус",
		"ready":         "📄 Готово",
		"vpn_list":      "👥 Список VPN",
		"vpn_add":       "➕ Добавить",
		"vpn_link":      "🔗 Ссылка / QR",
		"vpn_disable":   "🚫 Отключить",
		"vpn_enable":    "✅ Включить",
		"vpn_revoke":    "🗑 Отозвать",
		"session":       "🔑 API session",
		"admin":         "🖥 Админка",
		"help":          "❓ Справка",
		"routers":       "📡 Роутеры",
		"routers_title": "📡 <b>Роутеры</b>",
		"routers_empty": "Агентов пока нет.",
		"lang":          "🌐 Язык",
		"main_menu":     "🏠 Меню",
		"status_title":  "📊 <b>Статус</b>",
		"help_title":    "❓ <b>Справка</b>",
		"help_body":     "Бот только для оператора.\n• VPN: добавить / ссылка / вкл / выкл / отозвать\n• Session-токен для админки и Shortcuts\n• Статус и READY\n\nКоманды: /menu /status /vpn_list /lang /session /admin /help",
		"add_prompt":    "➕ <b>Новый VPN-пользователь</b>\n\nПришлите короткое <b>имя</b> (напр. <code>alice</code>).",
		"name_prompt":   "✏️ <b>%s</b>\n\nПришлите <b>имя пользователя VPN</b>:",
		"session_prompt": "🔑 <b>API session</b>\n\nПришлите срок в <b>часах</b> (напр. <code>72</code>) или /cancel:",
		"admin_body":    "🖥 <b>Админка</b>\n\nURL: <code>http://127.0.0.1:8787/admin/</code>\n\n1) SSH-туннель:\n<code>ssh -L 8787:127.0.0.1:8787 root@VPS</code>\n2) Откройте URL в браузере\n3) Вставьте session-токен\n\nИли: <code>netductor vpn api-bind detect</code>",
		"unknown":       "Неизвестное действие.",
		"unknown_cmd":   "Неизвестно. Откройте меню:",
		"no_ready":      "⚠️ Нет READY.txt",
		"ready_title":   "📄 <b>Готово</b>",
		"session_token": "🔑 Session-токен",
		"lang_now":      "🌐 Язык: <b>%s</b>\n\nВыберите:",
		"lang_en":       "English",
		"lang_ru":       "Русский",
		"lang_set":      "✅ Язык: <b>%s</b>",
		"copy_token":    "📋 Копировать токен",
		"show_links":    "🔗 Ссылки ещё раз",
		"cancelled":     "Отменено.",
		"vpn_users":     "👥 <b>Пользователи VPN</b>",
		"label_link":    "Ссылка / QR",
		"label_disable": "Отключить",
		"label_enable":  "Включить",
		"label_revoke":  "Отозвать",
		"cat_vpn":       "🔐 VPN",
		"cat_routers":   "📡 Роутеры",
		"cat_vpn_title": "🔐 <b>VPN</b>\nПользователи, ссылки, доступ.",
		"cat_routers_title": "📡 <b>Роутеры</b>\nУстройства, шаблоны, enroll.",
		"pending":       "⏳ Ожидают",
		"templates":     "📋 Шаблоны",
		"bind_tmpl":     "🔗 Привязать шаблон",
		"apply_tmpl":    "⚙️ Применить шаблон",
		"devices":       "📡 Устройства",
		"back_vpn":      "⬅️ VPN",
		"back_routers":  "⬅️ Роутеры",
		"pending_title": "⏳ <b>Ожидают</b>",
		"pending_empty": "Нет устройств в ожидании.",
		"templates_title": "📋 <b>Шаблоны</b>",
		"bind_prompt":   "ID устройства для привязки шаблона:",
		"apply_prompt":  "ID устройства для <code>apply_template</code>:",
		"menu_hint2":    "Выберите раздел — VPN или Роутеры откроют подменю.",
	},
}

func getLang() string {
	b, err := os.ReadFile(langFile)
	if err != nil {
		return "ru"
	}
	s := strings.TrimSpace(string(b))
	if s == "en" || s == "ru" {
		return s
	}
	return "ru"
}

func setLang(l string) {
	if l != "en" && l != "ru" {
		return
	}
	_ = os.WriteFile(langFile, []byte(l+"\n"), 0644)
}

func T(key string) string {
	l := getLang()
	if m, ok := dict[l]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if v, ok := dict["en"][key]; ok {
		return v
	}
	return key
}

func Tf(key string, args ...any) string {
	return fmt.Sprintf(T(key), args...)
}

func claimAdmin(chatID int64) {
	_ = os.MkdirAll("/etc/netductor/secrets", 0o700)
	_ = os.WriteFile(chatFile, []byte(fmt.Sprintf("%d\n", chatID)), 0o600)
}

func mustRead(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", path, err)
		os.Exit(1)
	}
	return strings.TrimSpace(string(b))
}

func readOptional(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func apiPost(token, method string, payload any) ([]byte, error) {
	body, _ := json.Marshal(payload)
	resp, err := http.Post("https://api.telegram.org/bot"+token+"/"+method, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		msg := string(data)
		if len(msg) > 200 {
			msg = msg[:200] + "…"
		}
		return data, fmt.Errorf("telegram %s HTTP %d: %s", method, resp.StatusCode, msg)
	}
	return data, nil
}

func apiGet(token, method string, v url.Values) ([]byte, error) {
	u := "https://api.telegram.org/bot" + token + "/" + method + "?" + v.Encode()
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func esc(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}

func btn(text, data, style string) map[string]any {
	b := map[string]any{"text": text, "callback_data": data}
	if style != "" {
		b["style"] = style
	}
	return b
}

func btnCopy(text, copyPayload string) map[string]any {
	return map[string]any{
		"text":      text,
		"copy_text": map[string]string{"text": copyPayload},
	}
}

func mainKeyboard() map[string]any {
	return map[string]any{
		"inline_keyboard": [][]map[string]any{
			{btn(T("status"), "m:status", "primary")},
			{btn(T("cat_vpn"), "m:cat:vpn", "primary"), btn(T("cat_routers"), "m:cat:routers", "primary")},
			{btn(T("session"), "m:session", ""), btn(T("admin"), "m:admin", "")},
			{btn("🗂 Nodes", "m:cat:nodes", "primary")},
			{btn(T("lang"), "m:lang", ""), btn(T("help"), "m:help", "")},
		},
	}
}

func vpnKeyboard() map[string]any {
	return map[string]any{
		"inline_keyboard": [][]map[string]any{
			{btn(T("vpn_list"), "m:vpn_list", "primary"), btn(T("vpn_add"), "m:vpn_add", "success")},
			{btn(T("vpn_link"), "m:vpn_link", "primary")},
			{btn(T("vpn_enable"), "m:vpn_enable", "success"), btn(T("vpn_disable"), "m:vpn_disable", "danger")},
			{btn(T("vpn_revoke"), "m:vpn_revoke", "danger")},
			{btn(T("main_menu"), "m:menu", "")},
		},
	}
}

func nodesKeyboard() map[string]any {
	return map[string]any{
		"inline_keyboard": [][]map[string]any{
			{btn("📋 List", "m:nodes_list", "primary"), btn("✏️ Rename", "m:node_rename", "")},
			{btn(T("main_menu"), "m:menu", "")},
		},
	}
}

func nodesText() string {
	out, err := exec.Command(netductorBin(), "nodes", "list").CombinedOutput()
	if err != nil {
		// fallback API-less: try reading via edge not available
		return strings.TrimSpace(string(out) + " " + err.Error())
	}
	return strings.TrimSpace(string(out))
}

func routersKeyboard() map[string]any {
	return map[string]any{
		"inline_keyboard": [][]map[string]any{
			{btn(T("devices"), "m:routers", "primary"), btn(T("pending"), "m:pending", "primary")},
			{btn(T("templates"), "m:templates", ""), btn(T("bind_tmpl"), "m:edge_bind", "")},
			{btn(T("apply_tmpl"), "m:edge_apply", "primary")},
			{btn(T("main_menu"), "m:menu", "")},
		},
	}
}

func backKeyboard() map[string]any {
	return map[string]any{
		"inline_keyboard": [][]map[string]any{
			{btn(T("main_menu"), "m:menu", "primary")},
		},
	}
}

func backTo(cat string) map[string]any {
	up := "m:menu"
	label := T("main_menu")
	switch cat {
	case "vpn":
		up, label = "m:cat:vpn", T("back_vpn")
	case "routers":
		up, label = "m:cat:routers", T("back_routers")
	}
	return map[string]any{
		"inline_keyboard": [][]map[string]any{
			{btn(label, up, "primary"), btn(T("main_menu"), "m:menu", "")},
		},
	}
}

func langKeyboard() map[string]any {
	return map[string]any{
		"inline_keyboard": [][]map[string]any{
			{btn(T("lang_ru"), "m:lang:ru", "primary"), btn(T("lang_en"), "m:lang:en", "")},
			{btn(T("main_menu"), "m:menu", "")},
		},
	}
}

func userCardKeyboard(name, subText string) map[string]any {
	rows := [][]map[string]any{
		{btn(T("show_links"), "u:link:"+name, "primary")},
		{btn(T("vpn_enable"), "u:enable:"+name, "success"), btn(T("vpn_disable"), "u:disable:"+name, "danger")},
		{btn(T("vpn_revoke"), "u:revoke:"+name, "danger")},
		{btn(T("main_menu"), "m:menu", "primary")},
	}
	if subText != "" {
		rows = append([][]map[string]any{{btnCopy(T("copy_token"), subText)}}, rows...)
	}
	return map[string]any{"inline_keyboard": rows}
}

func sendHTML(token string, chat int64, text string, kb map[string]any) {
	payload := map[string]any{
		"chat_id":    chat,
		"text":       text,
		"parse_mode": "HTML",
	}
	if kb != nil {
		payload["reply_markup"] = kb
	}
	_, _ = apiPost(token, "sendMessage", payload)
}

func editHTML(token string, chat int64, msgID int, text string, kb map[string]any) error {
	payload := map[string]any{
		"chat_id":    chat,
		"message_id": msgID,
		"text":       text,
		"parse_mode": "HTML",
	}
	if kb != nil {
		payload["reply_markup"] = kb
	}
	_, err := apiPost(token, "editMessageText", payload)
	return err
}

func reply(token string, chat int64, msgID int, text string, kb map[string]any) {
	if msgID > 0 {
		if err := editHTML(token, chat, msgID, text, kb); err == nil {
			return
		}
	}
	sendHTML(token, chat, text, kb)
}

func answerCallback(token, id string) {
	_, _ = apiPost(token, "answerCallbackQuery", map[string]any{"callback_query_id": id})
}

func formatVPNList(s string) string {
	if getLang() == "en" {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		p := strings.Split(line, "\t")
		if len(p) >= 2 {
			if p[1] == "on" {
				p[1] = "вкл"
			} else if p[1] == "off" {
				p[1] = "выкл"
			}
			lines[i] = strings.Join(p, "\t")
		}
	}
	return strings.Join(lines, "\n")
}

func netductorBin() string {
	for _, p := range []string{"/usr/local/bin/netductor", "/usr/bin/netductor"} {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return "netductor"
}

func runVPN(args ...string) string {
	full := append([]string{"vpn"}, args...)
	cmd := exec.Command(netductorBin(), full...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := string(out)
		if msg != "" {
			msg += string([]byte{10})
		}
		return strings.TrimSpace(msg + err.Error())
	}
	return string(out)
}

func runND(args ...string) string {
	cmd := exec.Command(netductorBin(), args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := string(out)
		if msg != "" {
			msg += string([]byte{10})
		}
		return strings.TrimSpace(msg + err.Error())
	}
	return string(out)
}

func statusInline() string {
	var b strings.Builder
	host, _ := os.Hostname()
	ru := getLang() != "en"
	act, inact := "active", "inactive"
	if ru {
		b.WriteString("Status Netductor host ")
		b.WriteString(host)
		b.WriteByte(10)
		act, inact = "active-ru", "inactive-ru"
	} else {
		b.WriteString("Netductor status on ")
		b.WriteString(host)
		b.WriteByte(10)
	}
	// localized labels
	if ru {
		act, inact = "активен", "неактивен"
		// rewrite header properly
		var hdr strings.Builder
		hdr.WriteString("Статус Netductor на ")
		hdr.WriteString(host)
		hdr.WriteByte(10)
		b.Reset()
		b.WriteString(hdr.String())
	}
	mapSt := func(s string) string {
		s = strings.TrimSpace(s)
		if s == "active" {
			return act
		}
		if s == "" || (ru && s != "active") {
			return inact
		}
		return s
	}
	for _, u := range []string{"sing-box", "blocky", "netductor-api", "netductor-api", "netductor-telegram-bot", "netductor-telegram-bot"} {
		out, _ := exec.Command("systemctl", "is-active", u).Output()
		b.WriteString(u + ": " + mapSt(string(out)))
		b.WriteByte(10)
	}
	if out, err := exec.Command("systemctl", "is-active", "netductor-metrics.timer").Output(); err == nil {
		b.WriteString("metrics-timer: " + mapSt(string(out)))
		b.WriteByte(10)
	}
	if out, err := exec.Command("docker", "ps", "--format", "{{.Names}}").Output(); err == nil {
		names := strings.TrimSpace(string(out))
		names = strings.ReplaceAll(names, string([]byte{10}), " ")
		b.WriteString("docker: " + names)
		b.WriteByte(10)
	}
	// identity
	if ipb, err := os.ReadFile("/etc/netductor/public_ip"); err == nil {
		ip := strings.TrimSpace(string(ipb))
		if ip != "" {
			if ru {
				b.WriteString("IP: ")
			} else {
				b.WriteString("IP: ")
			}
			b.WriteString(ip)
			b.WriteByte(10)
		}
	}
	if idb, err := os.ReadFile("/etc/netductor/node_id"); err == nil {
		id := strings.TrimSpace(string(idb))
		if id != "" {
			b.WriteString("node: ")
			b.WriteString(id)
			b.WriteByte(10)
		}
	}
	b.WriteByte(10)
	if ru {
		b.WriteString("Пользователи VPN:")
	} else {
		b.WriteString("VPN users:")
	}
	b.WriteByte(10)
	b.WriteString(formatVPNList(runVPN("list")))
	return b.String()
}

func statusText() string {
	env := append(os.Environ(), "FRESHVPS_LANG="+getLang())
	if st, err := os.Stat(statusSh); err == nil && st.Mode()&0111 != 0 {
		cmd := exec.Command(statusSh)
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if err == nil && len(bytes.TrimSpace(out)) > 0 {
			return string(out)
		}
	}
	if _, err := os.Stat(statusSh); err == nil {
		cmd := exec.Command("bash", statusSh)
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if err == nil && len(bytes.TrimSpace(out)) > 0 {
			return string(out)
		}
	}
	return statusInline()
}


func readyText() string {
	b, err := os.ReadFile("/etc/netductor/READY.txt")
	if err != nil {
		return ""
	}
	text := string(b)
	ru := getLang() != "en"
	// Prefer language section if bilingual file present
	if strings.Contains(text, "=== RU ===") && strings.Contains(text, "=== EN ===") {
		var part string
		if ru {
			i := strings.Index(text, "=== RU ===")
			j := strings.Index(text, "=== EN ===")
			if i >= 0 && j > i {
				part = strings.TrimSpace(text[i+len("=== RU ==="):j])
			}
		} else {
			i := strings.Index(text, "=== EN ===")
			if i >= 0 {
				part = strings.TrimSpace(text[i+len("=== EN ==="):])
			}
		}
		if part != "" {
			return part
		}
	}
	return strings.TrimSpace(text)
}


func routersText() string {
	out, err := exec.Command(netductorBin(), "edge", "list").CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(out) + " " + err.Error())
	}
	return strings.TrimSpace(string(out))
}

func pendingText() string {
	out, err := exec.Command(netductorBin(), "edge", "pending").CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(out))
	}
	return strings.TrimSpace(string(out))
}

func templatesText() string {
	out, err := exec.Command(netductorBin(), "edge", "templates").CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(out))
	}
	return strings.TrimSpace(string(out))
}

func pendingKeyboard(lines string) map[string]any {
	rows := [][]map[string]any{}
	for _, line := range strings.Split(lines, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		id := strings.Fields(line)[0]
		rows = append(rows, []map[string]any{
			btn("✅ "+id, "e:appr:"+id, "success"),
			btn("🚫 "+id, "e:deny:"+id, "danger"),
		})
	}
	rows = append(rows, []map[string]any{btn(T("back_routers"), "m:cat:routers", "primary"), btn(T("main_menu"), "m:menu", "")})
	return map[string]any{"inline_keyboard": rows}
}

func menuText() string {
	return T("menu_title") + "\n\n" + T("menu_hint2")
}

func helpText() string {
	return T("help_title") + "\n\n" + T("help_body")
}

var chatState = map[int64]string{}
var chatExtra = map[int64]string{}

func setState(chat int64, st, extra string) {
	if st == "" {
		delete(chatState, chat)
		delete(chatExtra, chat)
		return
	}
	chatState[chat] = st
	chatExtra[chat] = extra
}

func handleCallback(token string, cq *callbackQuery, admin int64) {
	if cq.From.ID != admin {
		return
	}
	answerCallback(token, cq.ID)
	chat := cq.Message.Chat.ID
	msgID := 0
	if cq.Message != nil {
		msgID = cq.Message.MessageID
	}
	data := cq.Data

	if strings.HasPrefix(data, "u:") {
		parts := strings.SplitN(data, ":", 3)
		if len(parts) == 3 {
			action, name := parts[1], parts[2]
			switch action {
			case "link":
				sub := runVPN("link", name)
				reply(token, chat, msgID, "🔗 <b>"+esc(name)+"</b>\n\n<pre>"+esc(sub)+"</pre>", userCardKeyboard(name, strings.TrimSpace(sub)))
			case "enable":
				reply(token, chat, msgID, "✅ <pre>"+esc(runVPN("enable", name))+"</pre>", backKeyboard())
			case "disable":
				reply(token, chat, msgID, "🚫 <pre>"+esc(runVPN("disable", name))+"</pre>", backKeyboard())
			case "revoke":
				reply(token, chat, msgID, "🗑 <pre>"+esc(runVPN("revoke", name))+"</pre>", backKeyboard())
			}
		}
		return
	}

	if strings.HasPrefix(data, "m:lang:") {
		l := strings.TrimPrefix(data, "m:lang:")
		setLang(l)
		name := "Русский"
		if l == "en" {
			name = "English"
		}
		reply(token, chat, msgID, Tf("lang_set", name), mainKeyboard())
		return
	}

	switch data {
	case "m:menu", "m:help":
		setState(chat, "", "")
		if data == "m:help" {
			reply(token, chat, msgID, helpText(), mainKeyboard())
		} else {
			reply(token, chat, msgID, menuText(), mainKeyboard())
		}
	case "m:cat:vpn":
		setState(chat, "", "")
		reply(token, chat, msgID, T("cat_vpn_title"), vpnKeyboard())
	case "m:cat:routers":
		setState(chat, "", "")
		reply(token, chat, msgID, T("cat_routers_title"), routersKeyboard())
	case "m:cat:nodes":
		setState(chat, "", "")
		reply(token, chat, msgID, "🗂 <b>Nodes</b>", nodesKeyboard())
	case "m:nodes_list":
		reply(token, chat, msgID, "🗂 <b>Nodes</b>\n\n<pre>"+esc(nodesText())+"</pre>", nodesKeyboard())
	case "m:node_rename":
		setState(chat, "wait_node_rename", "")
		reply(token, chat, msgID, "Send: <code>id new-hostname</code>\nExample: <code>nd-core-11870 nd-core-nl01</code>", backKeyboard())
	case "m:lang":
		cur := getLang()
		label := "Русский"
		if cur == "en" {
			label = "English"
		}
		reply(token, chat, msgID, Tf("lang_now", label), langKeyboard())
	case "m:routers":
		t := strings.TrimSpace(routersText())
		if t == "" {
			reply(token, chat, msgID, T("routers_title")+"\n\n"+T("routers_empty"), backTo("routers"))
		} else {
			if len(t) > 3500 {
				t = t[:3500] + "…"
			}
			reply(token, chat, msgID, T("routers_title")+"\n\n<pre>"+esc(t)+"</pre>", backTo("routers"))
		}
	case "m:pending":
		t := pendingText()
		if t == "" {
			reply(token, chat, msgID, T("pending_title")+"\n\n"+T("pending_empty"), backTo("routers"))
		} else {
			reply(token, chat, msgID, T("pending_title")+"\n\n<pre>"+esc(t)+"</pre>", pendingKeyboard(t))
		}
	case "m:templates":
		t := templatesText()
		if t == "" {
			t = "(none)"
		}
		reply(token, chat, msgID, T("templates_title")+"\n\n<pre>"+esc(t)+"</pre>", backTo("routers"))
	case "m:edge_apply":
		setState(chat, "wait_edge_apply", "")
		reply(token, chat, msgID, T("apply_prompt"), backTo("routers"))
	case "m:edge_bind":
		setState(chat, "wait_edge_bind_dev", "")
		reply(token, chat, msgID, T("bind_prompt"), backTo("routers"))
	case "m:status":
		reply(token, chat, msgID, T("status_title")+"\n\n<pre>"+esc(statusText())+"</pre>", backKeyboard())
	case "m:vpn_list":
		reply(token, chat, msgID, T("vpn_users")+"\n\n<pre>"+esc(formatVPNList(runVPN("list")))+"</pre>", backTo("vpn"))
	case "m:vpn_add":
		setState(chat, "wait_vpn_add_name", "")
		reply(token, chat, msgID, T("add_prompt"), backTo("vpn"))
	case "m:vpn_link", "m:vpn_disable", "m:vpn_enable", "m:vpn_revoke":
		action := strings.TrimPrefix(data, "m:")
		setState(chat, "wait_vpn_name:"+action, "")
		label := map[string]string{
			"vpn_link": T("label_link"), "vpn_disable": T("label_disable"),
			"vpn_enable": T("label_enable"), "vpn_revoke": T("label_revoke"),
		}[action]
		reply(token, chat, msgID, Tf("name_prompt", label), backTo("vpn"))
	case "m:admin":
		reply(token, chat, msgID, T("admin_body"), backKeyboard())
	case "m:session":
		setState(chat, "wait_session_hours", "")
		reply(token, chat, msgID, T("session_prompt"), backKeyboard())
	default:
		reply(token, chat, msgID, T("unknown"), mainKeyboard())
	}
}

func handleMessage(token string, m *message, admin int64) {
	if m.Chat.ID != admin {
		return
	}
	chat := m.Chat.ID
	text := strings.TrimSpace(m.Text)

	if text == "/cancel" {
		setState(chat, "", "")
		sendHTML(token, chat, T("cancelled"), mainKeyboard())
		return
	}

	st := chatState[chat]
	if st == "wait_vpn_add_name" {
		name := strings.Fields(text)[0]
		out := runVPN("add", name)
		setState(chat, "", "")
		sub := runVPN("link", name)
		sendHTML(token, chat, "✅ <pre>"+esc(out)+"</pre>\n\n<pre>"+esc(sub)+"</pre>", userCardKeyboard(name, strings.TrimSpace(sub)))
		return
	}
	if strings.HasPrefix(st, "wait_vpn_name:") {
		action := strings.TrimPrefix(st, "wait_vpn_name:")
		name := strings.Fields(text)[0]
		setState(chat, "", "")
		switch action {
		case "vpn_link":
			sub := runVPN("link", name)
			sendHTML(token, chat, "🔗 <b>"+esc(name)+"</b>\n\n<pre>"+esc(sub)+"</pre>", userCardKeyboard(name, strings.TrimSpace(sub)))
		case "vpn_disable":
			sendHTML(token, chat, "🚫 <pre>"+esc(runVPN("disable", name))+"</pre>", backKeyboard())
		case "vpn_enable":
			sendHTML(token, chat, "✅ <pre>"+esc(runVPN("enable", name))+"</pre>", backKeyboard())
		case "vpn_revoke":
			sendHTML(token, chat, "🗑 <pre>"+esc(runVPN("revoke", name))+"</pre>", backKeyboard())
		}
		return
	}
	if st == "wait_session_hours" {
		h := text
		if h == "" {
			h = "72"
		}
		setState(chat, "", "")
		tok := runVPN("session", h)
		fields := strings.Fields(tok)
		copyVal := tok
		if len(fields) > 0 {
			copyVal = fields[0]
		}
		kb := map[string]any{
			"inline_keyboard": [][]map[string]any{
				{btnCopy(T("copy_token"), copyVal)},
				{btn(T("main_menu"), "m:menu", "primary")},
			},
		}
		sendHTML(token, chat, "🔑 <code>"+esc(tok)+"</code>", kb)
		return
	}

	parts := strings.Fields(text)
	if len(parts) == 0 {
		sendHTML(token, chat, menuText(), mainKeyboard())
		return
	}
	cmd := parts[0]
	arg1 := ""
	if len(parts) > 1 {
		arg1 = parts[1]
	}

	switch cmd {
	case "/start", "/menu":
		sendHTML(token, chat, menuText(), mainKeyboard())
	case "/help":
		sendHTML(token, chat, helpText(), mainKeyboard())
	case "/lang":
		if arg1 == "en" || arg1 == "ru" {
			setLang(arg1)
			name := "Русский"
			if arg1 == "en" {
				name = "English"
			}
			sendHTML(token, chat, Tf("lang_set", name), mainKeyboard())
		} else {
			cur := getLang()
			label := "Русский"
			if cur == "en" {
				label = "English"
			}
			sendHTML(token, chat, Tf("lang_now", label), langKeyboard())
		}
	case "/status":
		sendHTML(token, chat, T("status_title")+"\n\n<pre>"+esc(statusText())+"</pre>", backKeyboard())
	case "/vpn_list":
		sendHTML(token, chat, T("vpn_users")+"\n\n<pre>"+esc(formatVPNList(runVPN("list")))+"</pre>", backKeyboard())
	case "/vpn_add":
		if arg1 == "" {
			setState(chat, "wait_vpn_add_name", "")
			sendHTML(token, chat, T("add_prompt"), backKeyboard())
		} else {
			out := runVPN("add", arg1)
			sub := runVPN("link", arg1)
			sendHTML(token, chat, "✅ <pre>"+esc(out)+"</pre>\n\n<pre>"+esc(sub)+"</pre>", userCardKeyboard(arg1, strings.TrimSpace(sub)))
		}
	case "/vpn_disable":
		sendHTML(token, chat, "🚫 <pre>"+esc(runVPN("disable", arg1))+"</pre>", backKeyboard())
	case "/vpn_enable":
		sendHTML(token, chat, "✅ <pre>"+esc(runVPN("enable", arg1))+"</pre>", backKeyboard())
	case "/vpn_revoke":
		sendHTML(token, chat, "🗑 <pre>"+esc(runVPN("revoke", arg1))+"</pre>", backKeyboard())
	case "/session":
		h := "72"
		if arg1 != "" {
			h = arg1
		}
		tok := runVPN("session", h)
		fields := strings.Fields(tok)
		copyVal := tok
		if len(fields) > 0 {
			copyVal = fields[0]
		}
		kb := map[string]any{
			"inline_keyboard": [][]map[string]any{
				{btnCopy(T("copy_token"), copyVal)},
				{btn(T("main_menu"), "m:menu", "primary")},
			},
		}
		sendHTML(token, chat, "🔑 <code>"+esc(tok)+"</code>", kb)
	case "/admin":
		sendHTML(token, chat, T("admin_body"), mainKeyboard())
	default:
		sendHTML(token, chat, T("unknown_cmd"), mainKeyboard())
	}
}

func setBotCommands(token string) {
	// descriptions in both langs is limited; use English short + user can /lang
	cmds := []map[string]string{
		{"command": "menu", "description": "Menu / Меню"},
		{"command": "status", "description": "Status / Статус"},
		{"command": "vpn_list", "description": "VPN users"},
		{"command": "lang", "description": "Language / Язык"},
		{"command": "help", "description": "Help / Справка"},
	}
	_, _ = apiPost(token, "setMyCommands", map[string]any{"commands": cmds})
}

func main() {
	token := mustRead(tokenFile)
	if token == "" {
		fmt.Fprintln(os.Stderr, "telegram_bot_token empty")
		os.Exit(1)
	}
	adminStr := readOptional(chatFile)
	admin, _ := strconv.ParseInt(adminStr, 10, 64)
	if admin == 0 {
		fmt.Fprintln(os.Stderr, "telegram_admin_id empty — waiting for first /start to claim admin")
	}
	// default language ru for this project
	if _, err := os.Stat(langFile); err != nil {
		setLang("ru")
	}
	setBotCommands(token)
	offset := 0
	for {
		v := url.Values{}
		v.Set("timeout", "30")
		v.Set("offset", strconv.Itoa(offset))
		v.Set("allowed_updates", `["message","callback_query"]`)
		body, err := apiGet(token, "getUpdates", v)
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}
		var wrap struct {
			OK     bool              `json:"ok"`
			Result []json.RawMessage `json:"result"`
		}
		if json.Unmarshal(body, &wrap) != nil || !wrap.OK {
			time.Sleep(2 * time.Second)
			continue
		}
		for _, raw := range wrap.Result {
			var u update
			if json.Unmarshal(raw, &u) != nil {
				continue
			}
			offset = u.UpdateID + 1
			if u.CallbackQuery != nil {
				if admin == 0 {
					continue
				}
				handleCallback(token, u.CallbackQuery, admin)
				continue
			}
			if u.Message != nil {
				if admin == 0 {
					txt := strings.TrimSpace(u.Message.Text)
					if txt == "/start" || strings.HasPrefix(txt, "/start ") {
						admin = u.Message.Chat.ID
						claimAdmin(admin)
						sendHTML(token, admin, "✅ Admin claimed for this chat. Use /menu", mainKeyboard())
						continue
					}
					sendHTML(token, u.Message.Chat.ID, "Send /start once to claim operator admin.", nil)
					continue
				}
				handleMessage(token, u.Message, admin)
			}
		}
	}
}
