#!/usr/bin/env bash
# Module: home edge VPN client (Armbian/Debian SBC)
# shellcheck disable=SC2154

module_edge_client_install() {
  require_root
  if [[ ! -f /etc/debian_version ]]; then
    die "edge-client supports Debian/Armbian only"
  fi

  # shellcheck source=/dev/null
  source "${FRESHVPS_ROOT}/lib/vless-parse.sh"
  # shellcheck source=/dev/null
  source "${FRESHVPS_ROOT}/lib/install-singbox.sh"

  local conf_dir=/usr/local/etc/sing-box
  local bin_dir=/usr/local/bin
  local edge_etc="${FRESHVPS_ETC}/edge"
  mkdir -p "${conf_dir}" "${edge_etc}" /var/lib/sing-box "${FRESHVPS_STATE_DIR}"

  pkg_install curl tar jq openssl iptables iproute2 ca-certificates
  install_singbox_binary "${bin_dir}"

  local vless_url="${EDGE_UPSTREAM_VLESS:-}"
  if [[ -z "${vless_url}" && -f "${edge_etc}/upstream.vless" ]]; then
    vless_url="$(tr -d '\r\n' < "${edge_etc}/upstream.vless")"
  fi
  if [[ -z "${vless_url}" ]]; then
    warn "EDGE_UPSTREAM_VLESS not set — write ${edge_etc}/upstream.vless"
  else
    printf '%s' "${vless_url}" > "${edge_etc}/upstream.vless"
    chmod 600 "${edge_etc}/upstream.vless"
  fi

  local lan_if="${EDGE_LAN_IF:-}"
  if [[ -z "${lan_if}" ]]; then
    lan_if="$(ip -4 route show default 2>/dev/null | awk '{print $5; exit}')"
    lan_if="${lan_if:-eth0}"
  fi
  local mode="${EDGE_MODE:-gateway}"
  local route_mode="${EDGE_ROUTE_MODE:-global}"

  local host port uuid sni pbk sid fp flow
  if [[ -n "${vless_url}" ]] && vless_parse "${vless_url}"; then
    host="${VLESS_HOST}"; port="${VLESS_PORT}"; uuid="${VLESS_UUID}"
    sni="${VLESS_SNI}"; pbk="${VLESS_PBK}"; sid="${VLESS_SID}"
    fp="${VLESS_FP}"; flow="${VLESS_FLOW}"
  else
    host="127.0.0.1"; port="443"; uuid="00000000-0000-0000-0000-000000000000"
    sni="www.microsoft.com"; pbk=""; sid=""; fp="chrome"; flow="xtls-rprx-vision"
  fi

  local route_rules final_out
  case "${route_mode}" in
    ru-direct)
      route_rules='{ "action": "sniff" },
      { "protocol": "dns", "action": "hijack-dns" },
      { "domain_suffix": [".ru",".xn--p1ai",".su"], "outbound": "direct" }'
      final_out=proxy
      ;;
    *)
      route_rules='{ "action": "sniff" }, { "protocol": "dns", "action": "hijack-dns" }'
      final_out=proxy
      ;;
  esac

  cat > "${conf_dir}/config.json" <<EOF
{
  "log": { "level": "info", "timestamp": true },
  "dns": {
    "servers": [
      { "type": "udp", "tag": "remote", "server": "1.1.1.1" },
      { "type": "local", "tag": "local" }
    ],
    "final": "remote",
    "strategy": "ipv4_only"
  },
  "inbounds": [
    {
      "type": "tun",
      "tag": "tun-in",
      "interface_name": "singbox",
      "address": ["172.19.0.1/30"],
      "mtu": 1400,
      "auto_route": true,
      "strict_route": true,
      "stack": "system"
    }
  ],
  "outbounds": [
    {
      "type": "vless",
      "tag": "proxy",
      "server": "${host}",
      "server_port": ${port},
      "uuid": "${uuid}",
      "flow": "${flow}",
      "tls": {
        "enabled": true,
        "server_name": "${sni}",
        "utls": { "enabled": true, "fingerprint": "${fp}" },
        "reality": {
          "enabled": true,
          "public_key": "${pbk}",
          "short_id": "${sid}"
        }
      },
      "packet_encoding": "xudp"
    },
    { "type": "direct", "tag": "direct" },
    { "type": "block", "tag": "block" }
  ],
  "route": {
    "rules": [ ${route_rules} ],
    "final": "${final_out}",
    "auto_detect_interface": true
  }
}
EOF

  if [[ "${mode}" == "gateway" ]]; then
    cat > /etc/sysctl.d/99-freshvps-edge.conf <<EOF
net.ipv4.ip_forward=1
net.ipv6.conf.all.forwarding=0
EOF
    sysctl -p /etc/sysctl.d/99-freshvps-edge.conf >/dev/null 2>&1 || true
    local wan_if
    wan_if="$(ip -4 route show default 2>/dev/null | awk '{print $5; exit}')"
    wan_if="${wan_if:-${lan_if}}"
    if command -v iptables >/dev/null 2>&1; then
      iptables -t nat -C POSTROUTING -o "${wan_if}" -j MASQUERADE 2>/dev/null || \
        iptables -t nat -A POSTROUTING -o "${wan_if}" -j MASQUERADE || true
    fi
  fi

  cat > /etc/systemd/system/sing-box.service <<EOF
[Unit]
Description=sing-box edge client (FreshVPS)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=${bin_dir}/sing-box run -c ${conf_dir}/config.json
Restart=on-failure
RestartSec=3
LimitNOFILE=1048576
AmbientCapabilities=CAP_NET_ADMIN CAP_NET_BIND_SERVICE
CapabilityBoundingSet=CAP_NET_ADMIN CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
EOF

  systemctl daemon-reload
  if [[ -n "${pbk}" ]]; then
    systemctl enable --now sing-box
    info "sing-box edge client started (mode=${mode}, route=${route_mode})"
  else
    systemctl enable sing-box
    warn "sing-box enabled but not started — set upstream VLESS first"
  fi

  cat > "${edge_etc}/README.txt" <<EOF
FreshVPS edge-client
Upstream: ${edge_etc}/upstream.vless
Config:   ${conf_dir}/config.json
Mode:     ${mode}
Route:    ${route_mode}
EOF
}

module_edge_client_uninstall() {
  systemctl disable --now sing-box 2>/dev/null || true
  rm -f /etc/systemd/system/sing-box.service
  systemctl daemon-reload
}
