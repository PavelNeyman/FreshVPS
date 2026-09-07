#!/bin/sh
# Install sing-box client + split routing on OpenWrt
# Expects vless-parse.sh sourced (vless_parse)

install_singbox_binary() {
    ARCH="$(uname -m)"
    case "$ARCH" in
        aarch64|arm64) SB_ARCH=arm64 ;;
        armv7l|armhf) SB_ARCH=armv7 ;;
        x86_64|amd64) SB_ARCH=amd64 ;;
        mips) SB_ARCH=mipsle ;;
        *) echo "[ERROR] Unsupported arch: $ARCH"; return 1 ;;
    esac
    TAG="$(wget -qO- https://api.github.com/repos/SagerNet/sing-box/releases/latest 2>/dev/null | sed -n 's/.*"tag_name": "\(v[^"]*\)".*/\1/p' | head -1)"
    [ -n "$TAG" ] || TAG="v1.12.0"
    VER="${TAG#v}"
    URL="https://github.com/SagerNet/sing-box/releases/download/${TAG}/sing-box-${VER}-linux-${SB_ARCH}.tar.gz"
    echo "[INFO] Fetch sing-box $TAG ($SB_ARCH)"
    mkdir -p /usr/local/bin /usr/local/etc/sing-box /tmp/sb-inst
    wget -qO /tmp/sb-inst/sb.tgz "$URL" || { echo "[ERROR] download failed"; return 1; }
    tar -xzf /tmp/sb-inst/sb.tgz -C /tmp/sb-inst
    BIN="$(find /tmp/sb-inst -type f -name sing-box | head -1)"
    [ -n "$BIN" ] || { echo "[ERROR] binary missing in archive"; return 1; }
    cp "$BIN" /usr/local/bin/sing-box
    chmod 755 /usr/local/bin/sing-box
    rm -rf /tmp/sb-inst
}

write_singbox_config() {
    ROUTE_MODE="${ROUTE_MODE:-ru-direct}"
    vless_parse "$EDGE_UPSTREAM_VLESS" || { echo "[ERROR] bad EDGE_UPSTREAM_VLESS"; return 1; }

    case "$ROUTE_MODE" in
        global)
            ROUTE_RULES='"rules": [ { "action": "sniff" }, { "protocol": "dns", "action": "hijack-dns" } ],
    "final": "proxy"'
            ;;
        blocked-only)
            ROUTE_RULES='"rules": [
      { "action": "sniff" },
      { "protocol": "dns", "action": "hijack-dns" },
      { "domain_suffix": ["instagram.com","cdninstagram.com","facebook.com","fbcdn.net","twitter.com","x.com","t.co","youtube.com","googlevideo.com","ytimg.com","google.com","googleapis.com","gstatic.com","telegram.org","t.me"], "outbound": "proxy" }
    ],
    "final": "direct"'
            ;;
        *)
            ROUTE_RULES='"rules": [
      { "action": "sniff" },
      { "protocol": "dns", "action": "hijack-dns" },
      { "domain_suffix": [".ru",".рф",".su",".рус"], "outbound": "direct" },
      { "domain_keyword": ["yandex","vk.com","mail.ru","wildberries","ozon.ru","avito","gosuslugi","sberbank","tinkoff"], "outbound": "direct" }
    ],
    "final": "proxy"'
            ;;
    esac

    cat > /usr/local/etc/sing-box/config.json <<EOF
{
  "log": { "level": "info", "timestamp": true },
  "dns": {
    "servers": [
      { "type": "udp", "tag": "remote", "server": "1.1.1.1", "detour": "proxy" },
      { "type": "udp", "tag": "local-ru", "server": "8.8.8.8", "detour": "direct" },
      { "type": "local", "tag": "local" }
    ],
    "rules": [
      { "domain_suffix": [".ru",".рф",".su"], "server": "local-ru" }
    ],
    "final": "remote",
    "strategy": "ipv4_only"
  },
  "inbounds": [
    {
      "type": "tun",
      "tag": "tun-in",
      "interface_name": "singtun0",
      "address": ["172.19.0.1/30"],
      "mtu": 1400,
      "auto_route": true,
      "strict_route": true,
      "stack": "system",
      "sniff": true
    }
  ],
  "outbounds": [
    {
      "type": "vless",
      "tag": "proxy",
      "server": "${VLESS_HOST}",
      "server_port": ${VLESS_PORT},
      "uuid": "${VLESS_UUID}",
      "flow": "${VLESS_FLOW}",
      "tls": {
        "enabled": true,
        "server_name": "${VLESS_SNI}",
        "utls": { "enabled": true, "fingerprint": "${VLESS_FP}" },
        "reality": {
          "enabled": true,
          "public_key": "${VLESS_PBK}",
          "short_id": "${VLESS_SID}"
        }
      },
      "packet_encoding": "xudp"
    },
    { "type": "direct", "tag": "direct" },
    { "type": "block", "tag": "block" }
  ],
  "route": {
    ${ROUTE_RULES},
    "auto_detect_interface": true
  }
}
EOF
}

install_procd_service() {
    cat > /etc/init.d/sing-box <<'INIT'
#!/bin/sh /etc/rc.common
START=99
STOP=10
USE_PROCD=1

start_service() {
    procd_open_instance
    procd_set_param command /usr/local/bin/sing-box run -c /usr/local/etc/sing-box/config.json
    procd_set_param respawn
    procd_set_param stdout 1
    procd_set_param stderr 1
    procd_close_instance
}
INIT
    chmod +x /etc/init.d/sing-box
    /etc/init.d/sing-box enable
    /etc/init.d/sing-box restart
}

apply_vpn_client() {
    [ "${ENABLE_VPN:-0}" = "1" ] || { echo "[SKIP] VPN (ENABLE_VPN!=1)"; return 0; }
    [ -n "$EDGE_UPSTREAM_VLESS" ] || { echo "[ERROR] EDGE_UPSTREAM_VLESS empty"; return 1; }
    command -v wget >/dev/null || opkg install wget-ssl 2>/dev/null || true
    install_singbox_binary || return 1
    write_singbox_config || return 1
    /usr/local/bin/sing-box check -c /usr/local/etc/sing-box/config.json || {
        echo "[ERROR] sing-box config invalid"
        return 1
    }
    install_procd_service
    echo "[OK] VPN client up (ROUTE_MODE=${ROUTE_MODE:-ru-direct})"
}
