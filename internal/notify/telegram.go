package notify

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/PavelNeyman/netductor/internal/paths"
)

func secret(name string) string {
	b, err := os.ReadFile(filepath.Join(paths.EtcDir(), "secrets", name))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func htmlEsc(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// Telegram sends a message to the configured admin chat (HTML-safe).
func Telegram(msg string) error {
	tok := secret("telegram_bot_token")
	chat := secret("telegram_admin_id")
	if tok == "" || chat == "" {
		return fmt.Errorf("telegram secrets not configured")
	}
	u := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tok)
	resp, err := http.PostForm(u, url.Values{
		"chat_id": {chat}, "text": {msg}, "parse_mode": {"HTML"},
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		// fallback plain text without parse_mode
		resp2, err2 := http.PostForm(u, url.Values{"chat_id": {chat}, "text": {msg}})
		if err2 != nil {
			return fmt.Errorf("telegram HTTP %d", resp.StatusCode)
		}
		defer resp2.Body.Close()
		if resp2.StatusCode >= 300 {
			return fmt.Errorf("telegram HTTP %d", resp2.StatusCode)
		}
	}
	return nil
}

// TelegramCmd reports a finished node command.
func TelegramCmd(node, cmd string, ok bool, log string) error {
	mark := "✅"
	if !ok {
		mark = "❌"
	}
	if len(log) > 1200 {
		log = log[len(log)-1200:]
	}
	msg := fmt.Sprintf("%s <b>%s</b> · <code>%s</code>\n<pre>%s</pre>", mark, htmlEsc(cmd), htmlEsc(node), htmlEsc(log))
	return Telegram(msg)
}
