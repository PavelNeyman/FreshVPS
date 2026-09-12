package main

import (
	"fmt"
	"strings"
)

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

	if strings.HasPrefix(data, "m:nr:") {
		id := strings.TrimPrefix(data, "m:nr:")
		setState(chat, "wait_node_newname", id)
		reply(token, chat, msgID, fmt.Sprintf(T("nodes_pick"), esc(id)), backKeyboard())
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
		reply(token, chat, msgID, T("nodes_title")+string([]byte{10, 10})+T("nodes_hint"), nodesKeyboard())
	case "m:nodes_list":
		reply(token, chat, msgID, T("nodes_title")+string([]byte{10, 10})+formatNodesListHTML()+string([]byte{10, 10})+"<i>"+T("nodes_hint")+"</i>", nodesKeyboard())
	case "m:node_rename":
		setState(chat, "", "")
		reply(token, chat, msgID, T("nodes_rename")+string([]byte{10, 10})+formatNodesListHTML(), nodesRenameKeyboard())
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

