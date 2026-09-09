package session

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PavelNeyman/netductor/internal/paths"
)

func Dir() string { return paths.SessionsDir() }

func TokenFromAuth(auth, cookie string) string {
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimSpace(auth[7:])
	}
	for _, p := range strings.Split(cookie, ";") {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "fv_session=") || strings.HasPrefix(p, "nd_session=") {
			if i := strings.Index(p, "="); i >= 0 {
				return strings.TrimSpace(p[i+1:])
			}
		}
	}
	return ""
}

func Expiry(token string) (int64, bool) {
	if token == "" || strings.Contains(token, "/") || strings.Contains(token, "..") || len(token) > 128 {
		return 0, false
	}
	b, err := os.ReadFile(filepath.Join(Dir(), token))
	if err != nil {
		return 0, false
	}
	exp, err := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
	if err != nil {
		return 0, false
	}
	if exp < time.Now().Unix() {
		_ = os.Remove(filepath.Join(Dir(), token))
		return 0, false
	}
	return exp, true
}

func Valid(token string) bool {
	_, ok := Expiry(token)
	return ok
}
