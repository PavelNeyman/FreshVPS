package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	tokenFile = "/etc/freshvps/secrets/telegram_bot_token"
	chatFile  = "/etc/freshvps/secrets/telegram_admin_id"
	vpnBin    = "/usr/local/bin/freshvps-vpn"
)

var (
	stateMu sync.Mutex
	state   = map[int64]string{}
	pending = map[int64]string{}
)

type update struct {
	UpdateID      int            `json:"update_id"`
	Message       *message       `json:"message"`
	CallbackQuery *callbackQuery `json:"callback_query"`
}

type message struct {
	MessageID     int    `json:"message_id"`
	Chat          chat   `json:"chat"`
	From          *user  `json:"from"`
	Text          string `json:"text"`
	ForwardFrom   *user  `json:"forward_from"`
	ForwardOrigin *struct {
		Type       string `json:"type"`
		SenderUser *user  `json:"sender_user"`
	} `json:"forward_origin"`
}

type user struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

type chat struct {
	ID int64 `json:"id"`
}

type callbackQuery struct {
	ID      string   `json:"id"`
	From    user     `json:"from"`
	Message *message `json:"message"`
	Data    string   `json:"data"`
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
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	u := "https://api.telegram.org/bot" + token + "/" + method
	resp, err := http.Post(u, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func apiGet(token, method string, vals url.Values) ([]byte, error) {
	u := "https://api.telegram.org/bot" + token + "/" + method
	if vals != nil {
		u += "?" + vals.Encode()
	}
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

func btn(text, data string) map[string]string {
	return map[string]string{"text": text, "callback_data": data}
}

func mainKeyboard() map[string]any {
	return map[string]any{
		"inline_keyboard": [][]map[string]string{
			{btn("📊 Status", "m:status"), btn("📄 READY", "m:ready")},
			{btn("👥 VPN list", "m:vpn_list"), btn("➕ Add user", "m:vpn_add")},
			{btn("🔗 Link / QR", "m:vpn_link"), btn("🚫 Disable", "m:vpn_disable")},
			{btn("✅ Enable", "m:vpn_enable"), btn("🗑 Revoke", "m:vpn_revoke")},
			{btn("🔑 API session", "m:session"), btn("❓ Help", "m:help")},
		},
	}
}

func backKeyboard() map[string]any {
	return map[string]any{
		"inline_keyboard": [][]map[string]string{
			{btn("🏠 Main menu", "m:menu")},
		},
	}
}

func sendHTML(token string, chat int64, text string, keyboard any) {
	payload := map[string]any{
		"chat_id":                  chat,
		"text":                     text,
		"parse_mode":               "HTML",
		"disable_web_page_preview": true,
	}
	if keyboard != nil {
		payload["reply_markup"] = keyboard
	}
	_, _ = apiPost(token, "sendMessage", payload)
}

func answerCallback(token, id, text string) {
	p := map[string]any{"callback_query_id": id}
	if text != "" {
		p["text"] = text
	}
	_, _ = apiPost(token, "answerCallbackQuery", p)
}

func sendPhotoFile(token string, chat int64, path, caption string) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("chat_id", strconv.FormatInt(chat, 10))
	_ = w.WriteField("caption", caption)
	_ = w.WriteField("parse_mode", "HTML")
	part, err := w.CreateFormFile("photo", filepath.Base(path))
	if err != nil {
		return
	}
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = io.Copy(part, f)
	_ = w.Close()
	req, err := http.NewRequest(http.MethodPost, "https://api.telegram.org/bot"+token+"/sendPhoto", &buf)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)
}

func runVPN(args ...string) string {
	out, err := exec.Command(vpnBin, args...).CombinedOutput()
	s := strings.TrimSpace(string(out))
	if err != nil && s == "" {
		return err.Error()
	}
	return s
}

func statusText() string {
	out, _ := exec.Command("/opt/freshvps/runtime/telegram/status.sh").CombinedOutput()
	if len(out) > 0 {
		return string(out)
	}
	return "status script missing"
}

func menuText() string {
	return "🛡 <b>FreshVPS Operator</b>\n" +
		"────────────────────\n" +
		"Choose an action below.\n" +
		"Only <b>you</b> (admin) can use this bot."
}

func helpText() string {
	return "❓ <b>Help</b>\n\n" +
		"• <b>Status</b> — services snapshot\n" +
		"• <b>READY</b> — install summary + operator links\n" +
		"• <b>VPN</b> — list / add / link / disable / enable / revoke\n" +
		"• <b>API session</b> — token for Shortcuts\n\n" +
		"Commands still work: <code>/status</code> <code>/vpn_add name</code> …\n\n" +
		"Forward someone’s message here to see their Telegram <b>id</b>.\n" +
		"Cancel a step with <code>/cancel</code>."
}

func setState(chat int64, s, extra string) {
	stateMu.Lock()
	defer stateMu.Unlock()
	if s == "" {
		delete(state, chat)
		delete(pending, chat)
		return
	}
	state[chat] = s
	if extra != "" {
		pending[chat] = extra
	} else {
		delete(pending, chat)
	}
}

func getState(chat int64) (string, string) {
	stateMu.Lock()
	defer stateMu.Unlock()
	return state[chat], pending[chat]
}

func deliverUserLinks(token string, chat int64, name string) {
	sub := filepath.Join("/etc/freshvps/clients", name, "subscription.txt")
	if b, err := os.ReadFile(sub); err == nil {
		sendHTML(token, chat, "📦 <b>Subscription</b> (<code>"+esc(name)+"</code>)\n\n<pre>"+esc(strings.TrimSpace(string(b)))+"</pre>", backKeyboard())
	} else if b, err := os.ReadFile(filepath.Join("/etc/freshvps/clients", name, "link.txt")); err == nil {
		sendHTML(token, chat, "🔗 <b>Link</b>\n\n<pre>"+esc(strings.TrimSpace(string(b)))+"</pre>", backKeyboard())
	} else {
		sendHTML(token, chat, "⚠️ User <code>"+esc(name)+"</code> not found.", backKeyboard())
		return
	}
	qr := filepath.Join("/etc/freshvps/clients", name, "qr.png")
	if _, err := os.Stat(qr); err == nil {
		sendPhotoFile(token, chat, qr, "📷 QR · "+name)
	}
}

func handleCallback(token string, cq *callbackQuery, admin int64) {
	if cq.From.ID != admin {
		answerCallback(token, cq.ID, "⛔ Not authorized")
		return
	}
	chat := cq.Message.Chat.ID
	data := cq.Data
	answerCallback(token, cq.ID, "")

	switch data {
	case "m:menu", "m:help":
		setState(chat, "", "")
		if data == "m:help" {
			sendHTML(token, chat, helpText(), mainKeyboard())
		} else {
			sendHTML(token, chat, menuText(), mainKeyboard())
		}
	case "m:status":
		sendHTML(token, chat, "📊 <b>Status</b>\n\n<pre>"+esc(statusText())+"</pre>", backKeyboard())
	case "m:ready":
		if b, err := os.ReadFile("/etc/freshvps/READY.txt"); err == nil {
			t := string(b)
			if len(t) > 3500 {
				t = t[:3500] + "\n…"
			}
			sendHTML(token, chat, "📄 <b>READY</b>\n\n<pre>"+esc(t)+"</pre>", backKeyboard())
		} else {
			sendHTML(token, chat, "⚠️ No READY.txt", backKeyboard())
		}
	case "m:vpn_list":
		sendHTML(token, chat, "👥 <b>VPN users</b>\n\n<pre>"+esc(runVPN("list"))+"</pre>", backKeyboard())
	case "m:vpn_add":
		setState(chat, "wait_vpn_add_name", "")
		sendHTML(token, chat,
			"➕ <b>Add VPN user</b>\n\n"+
				"Send a <b>short name</b> (e.g. <code>alice</code>).",
			backKeyboard())
	case "m:vpn_link", "m:vpn_disable", "m:vpn_enable", "m:vpn_revoke":
		action := strings.TrimPrefix(data, "m:")
		setState(chat, "wait_vpn_name:"+action, "")
		label := map[string]string{
			"vpn_link": "Link / QR", "vpn_disable": "Disable", "vpn_enable": "Enable", "vpn_revoke": "Revoke",
		}[action]
		sendHTML(token, chat, "✏️ <b>"+label+"</b>\n\nSend the VPN <b>user name</b>:", backKeyboard())
	case "m:session":
		setState(chat, "wait_session_hours", "")
		sendHTML(token, chat, "🔑 <b>API session</b>\n\nSend lifetime in <b>hours</b> (e.g. <code>72</code>), or /cancel:", backKeyboard())
	default:
		sendHTML(token, chat, "Unknown action.", mainKeyboard())
	}
}

func handleMessage(token string, m *message, admin int64) {
	if m.Chat.ID != admin {
		return
	}
	chat := m.Chat.ID
	text := strings.TrimSpace(m.Text)

	if m.ForwardFrom != nil || (m.ForwardOrigin != nil && m.ForwardOrigin.SenderUser != nil) {
		u := m.ForwardFrom
		if u == nil {
			u = m.ForwardOrigin.SenderUser
		}
		sendHTML(token, chat,
			fmt.Sprintf("👤 Forwarded user\nID: <code>%d</code>\nName: %s\nUsername: @%s",
				u.ID, esc(u.FirstName), esc(u.Username)),
			backKeyboard())
		return
	}

	st, extra := getState(chat)
	if text == "/cancel" {
		setState(chat, "", "")
		sendHTML(token, chat, "Cancelled.", mainKeyboard())
		return
	}

	switch {
	case st == "wait_vpn_add_name":
		if text == "" || strings.HasPrefix(text, "/") {
			sendHTML(token, chat, "Send a plain name, or /cancel", backKeyboard())
			return
		}
		name := strings.Fields(text)[0]
		setState(chat, "wait_vpn_add_note", name)
		sendHTML(token, chat,
			"📝 Optional <b>note</b> for <code>"+esc(name)+"</code>\n"+
				"(phone, who, …). Send <code>-</code> to skip.",
			backKeyboard())
		return
	case st == "wait_vpn_add_note":
		name := extra
		note := text
		if note == "-" {
			note = ""
		}
		setState(chat, "", "")
		out := runVPN("add", name, note)
		sendHTML(token, chat, "✅ <b>Created</b> <code>"+esc(name)+"</code>\n\n<pre>"+esc(out)+"</pre>", backKeyboard())
		deliverUserLinks(token, chat, name)
		return
	case strings.HasPrefix(st, "wait_vpn_name:"):
		action := strings.TrimPrefix(st, "wait_vpn_name:")
		if text == "" || strings.HasPrefix(text, "/") {
			sendHTML(token, chat, "Send user name, or /cancel", backKeyboard())
			return
		}
		name := strings.Fields(text)[0]
		setState(chat, "", "")
		switch action {
		case "vpn_link":
			deliverUserLinks(token, chat, name)
		case "vpn_disable":
			sendHTML(token, chat, "🚫 <pre>"+esc(runVPN("disable", name))+"</pre>", backKeyboard())
		case "vpn_enable":
			sendHTML(token, chat, "✅ <pre>"+esc(runVPN("enable", name))+"</pre>", backKeyboard())
		case "vpn_revoke":
			sendHTML(token, chat, "🗑 <pre>"+esc(runVPN("revoke", name))+"</pre>", backKeyboard())
		}
		return
	case st == "wait_session_hours":
		h := text
		if h == "" {
			h = "72"
		}
		setState(chat, "", "")
		tok := runVPN("session", h)
		sendHTML(token, chat, "🔑 <b>Session</b> ("+esc(h)+"h)\n\n<code>"+esc(tok)+"</code>", backKeyboard())
		return
	}

	if text == "" {
		return
	}
	fields := strings.Fields(text)
	cmd := fields[0]
	arg1, note := "", ""
	if len(fields) > 1 {
		arg1 = fields[1]
	}
	if len(fields) > 2 {
		note = strings.Join(fields[2:], " ")
	}
	switch cmd {
	case "/start", "/menu":
		sendHTML(token, chat, menuText(), mainKeyboard())
	case "/help":
		sendHTML(token, chat, helpText(), mainKeyboard())
	case "/status":
		sendHTML(token, chat, "📊 <b>Status</b>\n\n<pre>"+esc(statusText())+"</pre>", backKeyboard())
	case "/vpn_list":
		sendHTML(token, chat, "👥 <b>VPN users</b>\n\n<pre>"+esc(runVPN("list"))+"</pre>", backKeyboard())
	case "/vpn_add":
		if arg1 == "" {
			setState(chat, "wait_vpn_add_name", "")
			sendHTML(token, chat, "➕ Send VPN user <b>name</b>:", backKeyboard())
			return
		}
		out := runVPN("add", arg1, note)
		sendHTML(token, chat, "✅ <pre>"+esc(out)+"</pre>", backKeyboard())
		deliverUserLinks(token, chat, arg1)
	case "/vpn_link":
		if arg1 == "" {
			setState(chat, "wait_vpn_name:vpn_link", "")
			sendHTML(token, chat, "Send user name:", backKeyboard())
			return
		}
		deliverUserLinks(token, chat, arg1)
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
		sendHTML(token, chat, "🔑 <code>"+esc(runVPN("session", h))+"</code>", backKeyboard())
	case "/ready":
		if b, err := os.ReadFile("/etc/freshvps/READY.txt"); err == nil {
			t := string(b)
			if len(t) > 3500 {
				t = t[:3500] + "\n…"
			}
			sendHTML(token, chat, "📄 <pre>"+esc(t)+"</pre>", backKeyboard())
		} else {
			sendHTML(token, chat, "⚠️ no READY.txt", backKeyboard())
		}
	default:
		sendHTML(token, chat, "Unknown. Open the menu:", mainKeyboard())
	}
}

func main() {
	token := mustRead(tokenFile)
	adminStr := mustRead(chatFile)
	admin, err := strconv.ParseInt(adminStr, 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bad admin id: %v\n", err)
		os.Exit(1)
	}
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
