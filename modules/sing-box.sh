#!/usr/bin/env bash
# Module: sing-box binary + Reality/HY2 secrets (users by vpn-core)
# shellcheck disable=SC2154

module_sing_box_install() {
  local bin_dir=/usr/local/bin conf_dir=/usr/local/etc/sing-box
  # shellcheck source=/dev/null
  source "${FRESHVPS_ROOT}/lib/install-singbox.sh"

  pkg_install curl tar jq openssl
  mkdir -p "${conf_dir}" /var/lib/sing-box /etc/sing-box/certs "${FRESHVPS_STATE_DIR}"

  install_singbox_binary "${bin_dir}"

  local short_id private_key public_key sni
  short_id="$(read_secret singbox_short_id || true)"
  if [[ -z "${short_id}" ]]; then
    short_id="$(random_hex 4)"
    write_secret singbox_short_id "${short_id}"
  fi

  if [[ ! -f "${FRESHVPS_ETC}/secrets/singbox_reality_private" ]]; then
    local kp
    kp="$("${bin_dir}/sing-box" generate reality-keypair)"
    private_key="$(echo "${kp}" | awk '/PrivateKey/{print $2}')"
    public_key="$(echo "${kp}" | awk '/PublicKey/{print $2}')"
    write_secret singbox_reality_private "${private_key}"
    write_secret singbox_reality_public "${public_key}"
  else
    private_key="$(read_secret singbox_reality_private)"
    public_key="$(read_secret singbox_reality_public)"
  fi

  sni="${SINGBOX_REALITY_SNI:-www.cloudflare.com}"
  local vless_port="${SINGBOX_VLESS_PORT:-443}"
  local hy2_port="${SINGBOX_HY2_PORT:-8443}"

  if [[ ! -f /etc/sing-box/certs/hy2.crt ]]; then
    openssl req -x509 -nodes -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 \
      -keyout /etc/sing-box/certs/hy2.key -out /etc/sing-box/certs/hy2.crt \
      -days 3650 -subj "/CN=${sni}" 2>/dev/null
    chmod 600 /etc/sing-box/certs/hy2.key
  fi

  # Empty users until vpn-users seeds operator (no throwaway UUID)
  cat >"${conf_dir}/config.json" <<EOF
{
  "log": { "level": "info", "timestamp": true },
  "inbounds": [
    {
      "type": "vless",
      "tag": "vless-reality",
      "listen": "::",
      "listen_port": ${vless_port},
      "users": [],
      "tls": {
        "enabled": true,
        "server_name": "${sni}",
        "reality": {
          "enabled": true,
          "handshake": { "server": "${sni}", "server_port": 443 },
          "private_key": "${private_key}",
          "short_id": [ "${short_id}" ]
        }
      }
    },
    {
      "type": "hysteria2",
      "tag": "hy2",
      "listen": "::",
      "listen_port": ${hy2_port},
      "users": [],
      "tls": {
        "enabled": true,
        "alpn": [ "h3" ],
        "certificate_path": "/etc/sing-box/certs/hy2.crt",
        "key_path": "/etc/sing-box/certs/hy2.key"
      },
      "masquerade": "https://${sni}"
    }
  ],
  "outbounds": [ { "type": "direct", "tag": "direct" } ]
}
EOF
  chmod 600 "${conf_dir}/config.json"

  cat >/etc/systemd/system/sing-box.service <<'EOF'
[Unit]
Description=sing-box proxy
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/sing-box run -c /usr/local/etc/sing-box/config.json
Restart=on-failure
RestartSec=3
LimitNOFILE=1048576
AmbientCapabilities=CAP_NET_BIND_SERVICE
CapabilityBoundingSet=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
EOF

  systemd_enable_start sing-box
  firewall_allow_tcp "${vless_port}" "sing-box-vless"
  firewall_allow_udp "${hy2_port}" "sing-box-hy2"
  [[ -n "${PUBLIC_IP:-}" ]] && printf '%s\n' "${PUBLIC_IP}" >"${FRESHVPS_ETC}/public_ip"

  info "sing-box ready; Reality pubkey=${public_key} short_id=${short_id} SNI=${sni}"
}

module_sing_box_uninstall() {
  systemctl disable --now sing-box 2>/dev/null || true
  rm -f /etc/systemd/system/sing-box.service
  systemctl daemon-reload 2>/dev/null || true
  rm -f /usr/local/bin/sing-box
  # configs/certs kept unless FRESHVPS_PURGE=1 (handled in uninstall.sh --purge)
  info "sing-box stopped; binary removed (config kept unless --purge)"
}
