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
	tokenFile = "/etc/freshvps/secrets/telegram_bot_token"
	chatFile  = "/etc/freshvps/secrets/telegram_admin_id"
	langFile  = "/etc/freshvps/telegram_lang"
	statusSh  = "/opt/freshvps/runtime/telegram/status.sh"
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
		"menu_title":    "🛡 <b>FreshVPS operator panel</b>",
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
		"lang":          "🌐 Language",
		"main_menu":     "🏠 Main menu",
		"status_title":  "📊 <b>Status</b>",
		"help_title":    "❓ <b>Help</b>",
		"help_body":     "Operator-only bot.\n• VPN users: add / link / enable / disable / revoke\n• Session token for Admin UI &amp; Shortcuts\n• Status &amp; READY\n\nCommands: /menu /status /vpn_list /lang /session /admin /help",
		"add_prompt":    "➕ <b>Add VPN user</b>\n\nSend a short <b>name</b> (e.g. <code>alice</code>).",
		"name_prompt":   "✏️ <b>%s</b>\n\nSend the VPN <b>user name</b>:",
		"session_prompt": "🔑 <b>API session</b>\n\nSend lifetime in <b>hours</b> (e.g. <code>72</code>), or /cancel:",
		"admin_body":    "🖥 <b>Admin UI</b>\n\nURL: <code>http://127.0.0.1:8787/admin/</code>\n\n1) SSH tunnel:\n<code>ssh -L 8787:127.0.0.1:8787 root@VPS</code>\n2) Open URL in browser\n3) Paste session token\n\nOr: <code>freshvps-vpn api-bind detect</code>",
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
	},
	"ru": {
		"menu_title":    "🛡 <b>Панель оператора FreshVPS</b>",
		"menu_hint":     "Кнопки ниже. Слеш-команды тоже работают.",
		"status":        "📊 Статус",
		"ready":         "📄 READY",
		"vpn_list":      "👥 Список VPN",
		"vpn_add":       "➕ Добавить",
		"vpn_link":      "🔗 Ссылка / QR",
		"vpn_disable":   "🚫 Отключить",
		"vpn_enable":    "✅ Включить",
		"vpn_revoke":    "🗑 Отозвать",
		"session":       "🔑 API session",
		"admin":         "🖥 Админка",
		"help":          "❓ Справка",
		"lang":          "🌐 Язык",
		"main_menu":     "🏠 Меню",
		"status_title":  "📊 <b>Статус</b>",
		"help_title":    "❓ <b>Справка</b>",
		"help_body":     "Бот только для оператора.\n• VPN: добавить / ссылка / вкл / выкл / отозвать\n• Session-токен для админки и Shortcuts\n• Статус и READY\n\nКоманды: /menu /status /vpn_list /lang /session /admin /help",
		"add_prompt":    "➕ <b>Новый VPN-пользователь</b>\n\nПришлите короткое <b>имя</b> (напр. <code>alice</code>).",
		"name_prompt":   "✏️ <b>%s</b>\n\nПришлите <b>имя пользователя VPN</b>:",
		"session_prompt": "🔑 <b>API session</b>\n\nПришлите срок в <b>часах</b> (напр. <code>72</code>) или /cancel:",
		"admin_body":    "🖥 <b>Админка</b>\n\nURL: <code>http://127.0.0.1:8787/admin/</code>\n\n1) SSH-туннель:\n<code>ssh -L 8787:127.0.0.1:8787 root@VPS</code>\n2) Откройте URL в браузере\n3) Вставьте session-токен\n\nИли: <code>freshvps-vpn api-bind detect</code>",
		"unknown":       "Неизвестное действие.",
		"unknown_cmd":   "Неизвестно. Откройте меню:",
		"no_ready":      "⚠️ Нет READY.txt",
		"ready_title":   "📄 <b>Готово (READY)</b>",
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

func mustRead(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", path, err)
		os.Exit(1)
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
	return io.ReadAll(resp.Body)
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
			{btn(T("status"), "m:status", "primary"), btn(T("ready"), "m:ready", "")},
			{btn(T("vpn_list"), "m:vpn_list", ""), btn(T("vpn_add"), "m:vpn_add", "success")},
			{btn(T("vpn_link"), "m:vpn_link", "primary"), btn(T("vpn_disable"), "m:vpn_disable", "danger")},
			{btn(T("vpn_enable"), "m:vpn_enable", "success"), btn(T("vpn_revoke"), "m:vpn_revoke", "danger")},
			{btn(T("session"), "m:session", ""), btn(T("admin"), "m:admin", "primary")},
			{btn(T("lang"), "m:lang", ""), btn(T("help"), "m:help", "")},
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

func runVPN(args ...string) string {
	cmd := exec.Command("/usr/local/bin/freshvps-vpn", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out) + "\n" + err.Error()
	}
	return string(out)
}



func statusInline() string {
	var b strings.Builder
	host, _ := os.Hostname()
	ru := getLang() != "en"
	act, inact := "active", "inactive"
	if ru {
		b.WriteString("Status FreshVPS host ")
		b.WriteString(host)
		b.WriteByte(10)
		act, inact = "active-ru", "inactive-ru"
	} else {
		b.WriteString("FreshVPS status on ")
		b.WriteString(host)
		b.WriteByte(10)
	}
	// localized labels
	if ru {
		act, inact = "активен", "неактивен"
		// rewrite header properly
		var hdr strings.Builder
		hdr.WriteString("Статус FreshVPS на ")
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
	for _, u := range []string{"sing-box", "blocky", "freshvps-api", "freshvps-telegram-bot"} {
		out, _ := exec.Command("systemctl", "is-active", u).Output()
		b.WriteString(u + ": " + mapSt(string(out)))
		b.WriteByte(10)
	}
	if out, err := exec.Command("systemctl", "is-active", "freshvps-metrics.timer").Output(); err == nil {
		b.WriteString("metrics-timer: " + mapSt(string(out)))
		b.WriteByte(10)
	}
	if out, err := exec.Command("docker", "ps", "--format", "{{.Names}}").Output(); err == nil {
		names := strings.TrimSpace(string(out))
		names = strings.ReplaceAll(names, string([]byte{10}), " ")
		b.WriteString("docker: " + names)
		b.WriteByte(10)
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
	b, err := os.ReadFile("/etc/freshvps/READY.txt")
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

func menuText() string {
	return T("menu_title") + "\n\n" + T("menu_hint")
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
	data := cq.Data

	if strings.HasPrefix(data, "u:") {
		parts := strings.SplitN(data, ":", 3)
		if len(parts) == 3 {
			action, name := parts[1], parts[2]
			switch action {
			case "link":
				sub := runVPN("link", name)
				sendHTML(token, chat, "🔗 <b>"+esc(name)+"</b>\n\n<pre>"+esc(sub)+"</pre>", userCardKeyboard(name, strings.TrimSpace(sub)))
			case "enable":
				sendHTML(token, chat, "✅ <pre>"+esc(runVPN("enable", name))+"</pre>", backKeyboard())
			case "disable":
				sendHTML(token, chat, "🚫 <pre>"+esc(runVPN("disable", name))+"</pre>", backKeyboard())
			case "revoke":
				sendHTML(token, chat, "🗑 <pre>"+esc(runVPN("revoke", name))+"</pre>", backKeyboard())
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
		sendHTML(token, chat, Tf("lang_set", name), mainKeyboard())
		return
	}

	switch data {
	case "m:menu", "m:help":
		setState(chat, "", "")
		if data == "m:help" {
			sendHTML(token, chat, helpText(), mainKeyboard())
		} else {
			sendHTML(token, chat, menuText(), mainKeyboard())
		}
	case "m:lang":
		cur := getLang()
		label := "Русский"
		if cur == "en" {
			label = "English"
		}
		sendHTML(token, chat, Tf("lang_now", label), langKeyboard())
	case "m:status":
		sendHTML(token, chat, T("status_title")+"\n\n<pre>"+esc(statusText())+"</pre>", backKeyboard())
	case "m:ready":
		t := readyText()
		if t == "" {
			sendHTML(token, chat, T("no_ready"), backKeyboard())
		} else {
			if len(t) > 3500 {
				t = t[:3500] + "\n…"
			}
			sendHTML(token, chat, T("ready_title")+"\n\n<pre>"+esc(t)+"</pre>", backKeyboard())
		}
	case "m:vpn_list":
		sendHTML(token, chat, T("vpn_users")+"\n\n<pre>"+esc(formatVPNList(runVPN("list")))+"</pre>", backKeyboard())
	case "m:vpn_add":
		setState(chat, "wait_vpn_add_name", "")
		sendHTML(token, chat, T("add_prompt"), backKeyboard())
	case "m:vpn_link", "m:vpn_disable", "m:vpn_enable", "m:vpn_revoke":
		action := strings.TrimPrefix(data, "m:")
		setState(chat, "wait_vpn_name:"+action, "")
		label := map[string]string{
			"vpn_link": T("label_link"), "vpn_disable": T("label_disable"),
			"vpn_enable": T("label_enable"), "vpn_revoke": T("label_revoke"),
		}[action]
		sendHTML(token, chat, Tf("name_prompt", label), backKeyboard())
	case "m:admin":
		sendHTML(token, chat, T("admin_body"), backKeyboard())
	case "m:session":
		setState(chat, "wait_session_hours", "")
		sendHTML(token, chat, T("session_prompt"), backKeyboard())
	default:
		sendHTML(token, chat, T("unknown"), mainKeyboard())
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
	case "/ready":
		t := readyText()
		if t == "" {
			sendHTML(token, chat, T("no_ready"), backKeyboard())
		} else {
			if len(t) > 3500 {
				t = t[:3500] + "…"
			}
			sendHTML(token, chat, T("ready_title")+string([]byte{10, 10})+"<pre>"+esc(t)+"</pre>", backKeyboard())
		}
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
	adminStr := mustRead(chatFile)
	admin, err := strconv.ParseInt(adminStr, 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bad admin id: %v\n", err)
		os.Exit(1)
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
				handleCallback(token, u.CallbackQuery, admin)
				continue
			}
			if u.Message != nil {
				handleMessage(token, u.Message, admin)
			}
		}
	}
}
