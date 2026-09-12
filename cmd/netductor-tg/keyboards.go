package main

import (
	"fmt"
	"os/exec"
	"strings"
)

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

func relayKeyboard() map[string]any {
	return map[string]any{"inline_keyboard": [][]map[string]any{
		{btn(T("relay_export"), "m:relay:export", "primary")},
		{btn("➕ Enroll relay (SSH)", "m:relay:enroll", "primary")},
		{btn(T("relay_list"), "m:relay:list", "")},
		{btn("🇷🇺 RU exit ON", "m:relay:exit:on", "success"), btn("RU exit OFF", "m:relay:exit:off", "danger")},
		{btn("🔄 Sync relays", "m:relay:sync", "primary")},
		{btn(T("main_menu"), "m:menu", "primary")},
	}}
}

func mainKeyboard() map[string]any {
	return map[string]any{
		"inline_keyboard": [][]map[string]any{
			{btn(T("status"), "m:status", "primary")},
			{btn(T("cat_vpn"), "m:cat:vpn", "primary"), btn(T("cat_routers"), "m:cat:routers", "primary")},
			{btn(T("session"), "m:session", ""), btn(T("admin"), "m:admin", "")},
			{btn(T("addons"), "m:addons", ""), btn(T("nodes"), "m:cat:nodes", "primary")},
			{btn(T("relay"), "m:cat:relay", "primary")},
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

func addonsKeyboard() map[string]any {
	return map[string]any{"inline_keyboard": [][]map[string]any{
		{btn("Lampac", "m:addon:lampac", "primary")},
		{btn(T("back"), "m:menu", "")},
	}}
}

func nodesKeyboard() map[string]any {
	return map[string]any{
		"inline_keyboard": [][]map[string]any{
			{btn("📋 List", "m:nodes_list", "primary"), btn("✏️ Rename", "m:node_rename", "")},
			{btn(T("main_menu"), "m:menu", "")},
		},
	}
}

type nodeRow struct {
	ID, Host, Role, Kind, IP, Desired string
}

func parseNodesList() []nodeRow {
	out, err := exec.Command(netductorBin(), "nodes", "list").CombinedOutput()
	if err != nil {
		return nil
	}
	var rows []nodeRow
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "(") {
			continue
		}
		// id\thost=...\trole=...
		parts := strings.Split(line, "	")
		if len(parts) == 0 {
			continue
		}
		r := nodeRow{}
		for i, p := range parts {
			switch {
			case strings.HasPrefix(p, "id="):
				r.ID = strings.TrimPrefix(p, "id=")
			case strings.HasPrefix(p, "host="):
				r.Host = strings.TrimPrefix(p, "host=")
			case strings.HasPrefix(p, "role="):
				r.Role = strings.TrimPrefix(p, "role=")
			case strings.HasPrefix(p, "kind="):
				r.Kind = strings.TrimPrefix(p, "kind=")
			case strings.HasPrefix(p, "ip="):
				r.IP = strings.TrimPrefix(p, "ip=")
			case strings.HasPrefix(p, "desired="):
				r.Desired = strings.TrimPrefix(p, "desired=")
			case i == 0 && !strings.Contains(p, "="):
				r.ID = p // legacy
			}
		}
		if r.Host == "" {
			r.Host = r.ID
		}
		rows = append(rows, r)
	}
	return rows
}

func formatNodesListHTML() string {
	rows := parseNodesList()
	if len(rows) == 0 {
		return T("nodes_empty")
	}
	var b strings.Builder
	for i, r := range rows {
		label := r.Host
		if label == "" {
			label = r.ID
		}
		b.WriteString(fmt.Sprintf("%d. <b>%s</b>", i+1, esc(label)))
		if r.ID != "" && r.ID != r.Host {
			b.WriteString(" <code>"+esc(r.ID)+"</code>")
		}
		if r.Role != "" {
			b.WriteString(" · "+esc(r.Role))
		}
		if r.Kind != "" {
			b.WriteString(" · "+esc(r.Kind))
		}
		if r.IP != "" {
			b.WriteString(" · "+esc(r.IP))
		}
		if r.Desired != "" {
			b.WriteString(" → <i>"+esc(r.Desired)+"</i>")
		}
		b.WriteByte(10)
	}
	return strings.TrimRight(b.String(), "\n")
}

func nodesRenameKeyboard() map[string]any {
	rows := parseNodesList()
	kb := [][]map[string]any{}
	for i, r := range rows {
		name := r.Host
		if name == "" {
			name = r.ID
		}
		label := fmt.Sprintf("%d. %s", i+1, name)
		if len(label) > 40 {
			label = label[:40]
		}
		// callback max 64 bytes
		data := "m:nr:" + r.ID
		if len(data) > 64 {
			data = data[:64]
		}
		kb = append(kb, []map[string]any{btn(label, data, "")})
	}
	kb = append(kb, []map[string]any{btn(T("main_menu"), "m:menu", "primary")})
	return map[string]any{"inline_keyboard": kb}
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
	return userCardKeyboardMode(name, "vless")
}

// mode: vless | hy2 — top button toggles QR in the same message
func userCardKeyboardMode(name, mode string) map[string]any {
	toggle := btn("📱 HY2 QR", "u:hy2qr:"+name, "primary")
	if mode == "hy2" {
		toggle = btn("📱 VLESS QR", "u:vlessqr:"+name, "primary")
	}
	rows := [][]map[string]any{
		{toggle},
		{btn(T("show_links"), "u:link:"+name, "")},
		{btn(T("vpn_enable"), "u:enable:"+name, "success"), btn(T("vpn_disable"), "u:disable:"+name, "danger")},
		{btn(T("vpn_revoke"), "u:revoke:"+name, "danger")},
		{btn(T("main_menu"), "m:menu", "primary"), btn(T("cat_vpn"), "m:cat:vpn", "")},
	}
	return map[string]any{"inline_keyboard": rows}
}

func vpnUsersKeyboard() map[string]any {
	return vpnUsersKeyboardFor("link")
}

// action: link | enable | disable | revoke
func vpnUsersKeyboardFor(action string) map[string]any {
	if action == "" {
		action = "link"
	}
	raw := runVPN("list")
	rows := [][]map[string]any{}
	n := 0
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 1 {
			parts = strings.Fields(line)
		}
		if len(parts) < 1 || parts[0] == "" {
			continue
		}
		name := parts[0]
		if name == "relay-uplink" {
			continue
		}
		en := ""
		if len(parts) > 1 {
			en = parts[1]
		}
		n++
		icon := "🔗"
		switch action {
		case "enable":
			icon = "✅"
		case "disable":
			icon = "🚫"
		case "revoke":
			icon = "🗑"
		}
		if en == "off" {
			icon = "🔴"
		} else if en == "on" && action == "link" {
			icon = "🟢"
		}
		label := icon + " " + fmt.Sprintf("%d. %s", n, name)
		rows = append(rows, []map[string]any{btn(label, "u:"+action+":"+name, "primary")})
		if n >= 30 {
			break
		}
	}
	if n == 0 {
		rows = append(rows, []map[string]any{btn("— empty —", "m:cat:vpn", "")})
	}
	rows = append(rows, []map[string]any{btn(T("main_menu"), "m:menu", "primary"), btn(T("cat_vpn"), "m:cat:vpn", "")})
	return map[string]any{"inline_keyboard": rows}
}
