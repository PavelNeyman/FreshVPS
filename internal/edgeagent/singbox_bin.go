package edgeagent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// EnsureSingBox places sing-box at dest (default /usr/sbin/sing-box).
func EnsureSingBox(dest string) (string, error) {
	if dest == "" {
		dest = "/usr/sbin/sing-box"
	}
	if p, err := exec.LookPath("sing-box"); err == nil {
		return p, nil
	}
	if st, err := os.Stat(dest); err == nil && !st.IsDir() {
		return dest, nil
	}
	arch := runtime.GOARCH
	switch arch {
	case "arm":
		arch = "armv7"
	}
	tag, err := latestSingBoxTag()
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://github.com/SagerNet/sing-box/releases/download/%s/sing-box-%s-linux-%s.tar.gz",
		tag, strings.TrimPrefix(tag, "v"), arch)
	tmp, err := os.MkdirTemp("", "sb-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)
	tgz := filepath.Join(tmp, "sb.tgz")
	if err := download(url, tgz); err != nil {
		return "", err
	}
	if out, err := exec.Command("tar", "-xzf", tgz, "-C", tmp).CombinedOutput(); err != nil {
		return "", fmt.Errorf("tar: %s %v", out, err)
	}
	var bin string
	_ = filepath.Walk(tmp, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && info.Name() == "sing-box" {
			bin = path
		}
		return nil
	})
	if bin == "" {
		return "", fmt.Errorf("sing-box not in archive")
	}
	_ = os.MkdirAll(filepath.Dir(dest), 0o755)
	data, err := os.ReadFile(bin)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(dest, data, 0o755); err != nil {
		return "", err
	}
	return dest, nil
}

func latestSingBoxTag() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/SagerNet/sing-box/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var m struct {
		Tag string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return "", err
	}
	if m.Tag == "" {
		return "", fmt.Errorf("no tag")
	}
	return m.Tag, nil
}

func download(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}
