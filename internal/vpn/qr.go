package vpn

import (
	"fmt"
	"os"
	"path/filepath"

	qrcode "github.com/skip2/go-qrcode"
)

func WriteQR(name, payload string) (string, error) {
	dir := filepath.Join(Clients(), name)
	_ = os.MkdirAll(dir, 0o700)
	path := filepath.Join(dir, "qr.png")
	if err := qrcode.WriteFile(payload, qrcode.Medium, 256, path); err != nil {
		return "", err
	}
	_ = os.Chmod(path, 0o600)
	return path, nil
}

func ASCIIQR(payload string) string {
	q, err := qrcode.New(payload, qrcode.Medium)
	if err != nil {
		return ""
	}
	return q.ToSmallString(false)
}

func PrintASCIIQR(payload string) {
	s := ASCIIQR(payload)
	if s != "" {
		fmt.Print(s)
	}
}
