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

// Telegram sends a message to the configured admin chat.
func Telegram(msg string) error {
	tok := secret("telegram_bot_token")
	chat := secret("telegram_admin_id")
	if tok == "" || chat == "" {
		return fmt.Errorf("telegram secrets not configured")
	}
	u := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tok)
	resp, err := http.PostForm(u, url.Values{"chat_id": {chat}, "text": {msg}, "parse_mode": {"HTML"}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram HTTP %d", resp.StatusCode)
	}
	return nil
}
