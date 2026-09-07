#!/usr/bin/env bash
# Module: home edge VPN client (Armbian/Debian SBC) — NOT for VPS server role
# shellcheck disable=SC2154

module_edge_client_install() {
  require_root
  if [[ ! -f /etc/debian_version ]]; then
    die "edge-client supports Debian/Armbian only"
  fi

  local conf_dir=/usr/local/etc/sing-box
  local bin_dir=/usr/local/bin
  local edge_etc="${FRESHVPS_ETC}/edge"
  mkdir -p "${conf_dir}" "${edge_etc}" /var/lib/sing-box

  pkg_install curl tar jq openssl iptables iproute2 ca-certificates

  local arch tag url tmp
  arch="$(arch_go)"
  tag="$(curl -fsSL https://api.github.com/repos/SagerNet/sing-box/releases/latest | jq -r .tag_name)"
  [[ -n "${tag}" && "${tag}" != null ]] || die "Cannot resolve sing-box release"
  url="https://github.com/SagerNet/sing-box/releases/download/${tag}/sing-box-${tag#v}-linux-${arch}.tar.gz"
  tmp="$(mktemp -d)"
  info "Downloading sing-box ${tag} (${arch})"
  curl -fsSL "${url}" -o "${tmp}/sb.tgz"
  tar -xzf "${tmp}/sb.tgz" -C "${tmp}"
  install -m 755 "${tmp}/sing-box-${tag#v}-linux-${arch}/sing-box" "${bin_dir}/sing-box"
  rm -rf "${tmp}"

  local vless_url="${EDGE_UPSTREAM_VLESS:-}"
  if [[ -z "${vless_url}" && -f "${edge_etc}/upstream.vless" ]]; then
    vless_url="$(tr -d rn < "${edge_etc}/upstream.vless")"
  fi
  if [[ -z "${vless_url}" ]]; then
    warn "EDGE_UPSTREAM_VLESS not set — write ${edge_etc}/upstream.vless or export EDGE_UPSTREAM_VLESS"
    warn "Generating placeholder config; service will not route until URL is set"
  else
    printf "%s" "${vless_url}" > "${edge_etc}/upstream.vless"
    chmod 600 "${edge_etc}/upstream.vless"
  fi

  local lan_if="${EDGE_LAN_IF:-}"
  if [[ -z "${lan_if}" ]]; then
    lan_if="$(ip -4 route show default 2>/dev/null | awk "{print \$5; exit}")"
    lan_if="${lan_if:-eth0}"
  fi
  local mode="${EDGE_MODE:-gateway}" # gateway | tun-only

  # Parse minimal fields from vless:// for sing-box outbound (best-effort)
  # Format: vless://uuid@host:port?params#name
  local uuid host port sni pbk sid fp flow
  if [[ -n "${vless_url}" && "${vless_url}" == vless://* ]]; then
    local rest="${vless_url#vless://}"
    uuid="${rest%%@*}"
    rest="${rest#*@}"
    local hostport="${rest%%\?*}"
    host="${hostport%%:*}"
    port="${hostport##*:}"
    local qs="${rest#*\?}"
    qs="${qs%%#*}"
    sni="$(echo "${qs}" | tr "&" "\n" | awk -F= "/^sni=/{print \$2; exit}")"
    pbk="$(echo "${qs}" | tr "&" "\n" | awk -F= "/^pbk=/{print \$2; exit}")"
    sid="$(echo "${qs}" | tr "&" "\n" | awk -F= "/^sid=/{print \$2; exit}")"
    fp="$(echo "${qs}" | tr "&" "\n" | awk -F= "/^fp=/{print \$2; exit}")"
    flow="$(echo "${qs}" | tr "&" "\n" | awk -F= "/^flow=/{print \$2; exit}")"
    sni="${sni:-www.microsoft.com}"
    fp="${fp:-chrome}"
    flow="${flow:-xtls-rprx-vision}"
    port="${port:-443}"
  else
    host="127.0.0.1"
    port="443"
    uuid="00000000-0000-0000-0000-000000000000"
    sni="www.microsoft.com"
    pbk=""
    sid=""
    fp="chrome"
    flow="xtls-rprx-vision"
  fi

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
    "rules": [
      { "action": "sniff" },
      { "protocol": "dns", "action": "hijack-dns" }
    ],
    "final": "proxy",
    "auto_detect_interface": true
  }
}
EOF

  if [[ "${mode}" == "gateway" ]]; then
    # Enable IPv4 forward for LAN clients using this host as default GW
    cat > /etc/sysctl.d/99-freshvps-edge.conf <<EOF
net.ipv4.ip_forward=1
net.ipv6.conf.all.forwarding=0
EOF
    sysctl -p /etc/sysctl.d/99-freshvps-edge.conf >/dev/null 2>&1 || true

    # NAT LAN -> outbound (tun will take over routed traffic when auto_route works;
    # MASQUERADE on default iface helps mixed setups)
    if command -v iptables >/dev/null 2>&1; then
      iptables -t nat -C POSTROUTING -o "${lan_if}" -j MASQUERADE 2>/dev/null || \
        iptables -t nat -A POSTROUTING -o "${lan_if}" -j MASQUERADE || true
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
    info "sing-box edge client started (mode=${mode}, lan_if=${lan_if})"
  else
    systemctl enable sing-box
    warn "sing-box enabled but not started — set upstream VLESS first"
  fi

  cat > "${edge_etc}/README.txt" <<EOF
FreshVPS edge-client
Upstream: ${edge_etc}/upstream.vless
Config:   ${conf_dir}/config.json
Mode:     ${mode}
LAN if:   ${lan_if}

MikroTik: point VPN subnet gateway to this host IP, or set DHCP gateway.
EOF
  info "Edge client ready. Host IP: $(hostname -I 2>/dev/null | awk "{print \$1}")"
}

module_edge_client_uninstall() {
  systemctl disable --now sing-box 2>/dev/null || true
  rm -f /etc/systemd/system/sing-box.service
  systemctl daemon-reload
  info "edge-client sing-box stopped"
}
