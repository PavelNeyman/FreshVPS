package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	tokenFile = "/etc/freshvps/secrets/telegram_bot_token"
	chatFile  = "/etc/freshvps/secrets/telegram_admin_id"
	vpnBin    = "/usr/local/bin/freshvps-vpn"
)

type apiResp struct {
	OK     bool              `json:"ok"`
	Result []json.RawMessage `json:"result"`
}

type update struct {
	UpdateID int `json:"update_id"`
	Message  *struct {
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
}

func mustRead(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", path, err)
		os.Exit(1)
	}
	return strings.TrimSpace(string(b))
}

func api(token, method string, vals url.Values) ([]byte, error) {
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

func send(token string, chat int64, text string) {
	v := url.Values{}
	v.Set("chat_id", strconv.FormatInt(chat, 10))
	v.Set("text", text)
	_, _ = api(token, "sendMessage", v)
}

func sendFile(token string, chat int64, path, caption string) {
	// send as document via multipart would need more code; use photo URL file attach via curl fallback
	cmd := exec.Command("curl", "-fsS", "-X", "POST",
		"https://api.telegram.org/bot"+token+"/sendPhoto",
		"-F", "chat_id="+strconv.FormatInt(chat, 10),
		"-F", "photo=@"+path,
		"-F", "caption="+caption,
	)
	_ = cmd.Run()
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

func handle(token string, chat int64, text string) {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return
	}
	cmd := fields[0]
	arg1, note := "", ""
	if len(fields) > 1 {
		arg1 = fields[1]
	}
	if len(fields) > 2 {
		note = strings.Join(fields[2:], " ")
	}
	switch cmd {
	case "/start", "/help":
		send(token, chat, "FreshVPS operator bot\n/status /ready /vpn_list\n/vpn_add <name> [note]\n/vpn_link <name>\n/vpn_disable|/vpn_enable|/vpn_revoke <name>\n/session [hours]")
	case "/status":
		send(token, chat, statusText())
	case "/vpn_list":
		send(token, chat, runVPN("list"))
	case "/vpn_add":
		if arg1 == "" {
			send(token, chat, "usage: /vpn_add name [note]")
			return
		}
		out := runVPN("add", arg1, note)
		send(token, chat, out)
		link := filepath.Join("/etc/freshvps/clients", arg1, "link.txt")
		if b, err := os.ReadFile(link); err == nil {
			send(token, chat, strings.TrimSpace(string(b)))
		}
		qr := filepath.Join("/etc/freshvps/clients", arg1, "qr.png")
		if _, err := os.Stat(qr); err == nil {
			sendFile(token, chat, qr, arg1)
		}
		sub := filepath.Join("/etc/freshvps/clients", arg1, "subscription.txt")
		if b, err := os.ReadFile(sub); err == nil {
			send(token, chat, "Subscription (VLESS+HY2):\n"+strings.TrimSpace(string(b)))
		}
	case "/vpn_link":
		if arg1 == "" {
			send(token, chat, "usage: /vpn_link name")
			return
		}
		link := filepath.Join("/etc/freshvps/clients", arg1, "subscription.txt")
		if b, err := os.ReadFile(link); err == nil {
			send(token, chat, strings.TrimSpace(string(b)))
		} else if b, err := os.ReadFile(filepath.Join("/etc/freshvps/clients", arg1, "link.txt")); err == nil {
			send(token, chat, strings.TrimSpace(string(b)))
		} else {
			send(token, chat, "not found")
		}
		qr := filepath.Join("/etc/freshvps/clients", arg1, "qr.png")
		if _, err := os.Stat(qr); err == nil {
			sendFile(token, chat, qr, arg1)
		}
	case "/vpn_disable":
		send(token, chat, runVPN("disable", arg1))
	case "/vpn_enable":
		send(token, chat, runVPN("enable", arg1))
	case "/vpn_revoke":
		send(token, chat, runVPN("revoke", arg1))
	case "/session":
		h := "72"
		if arg1 != "" {
			h = arg1
		}
		send(token, chat, "Session token:\n"+runVPN("session", h))
	case "/ready":
		if b, err := os.ReadFile("/etc/freshvps/READY.txt"); err == nil {
			t := string(b)
			if len(t) > 3500 {
				t = t[:3500]
			}
			send(token, chat, t)
		} else {
			send(token, chat, "no READY.txt")
		}
	default:
		send(token, chat, "Unknown. /help")
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
		body, err := api(token, "getUpdates", v)
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}
		var ar apiResp
		if json.Unmarshal(body, &ar) != nil || !ar.OK {
			time.Sleep(2 * time.Second)
			continue
		}
		for _, raw := range ar.Result {
			var u update
			if json.Unmarshal(raw, &u) != nil {
				continue
			}
	offset = u.UpdateID + 1
			if u.Message == nil || u.Message.Chat.ID != admin {
				continue
			}
			handle(token, u.Message.Chat.ID, u.Message.Text)
		}
	}
}
