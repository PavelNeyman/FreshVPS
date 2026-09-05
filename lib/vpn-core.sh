#!/usr/bin/env bash
# FreshVPS VPN user core — sourced by CLI, modules, API, Telegram
# shellcheck disable=SC2034

: "${FRESHVPS_ETC:=/etc/freshvps}"
: "${FRESHVPS_STATE_DIR:=/var/lib/freshvps}"

VPN_USERS_FILE="${FRESHVPS_ETC}/vpn-users.json"
VPN_CLIENTS_DIR="${FRESHVPS_ETC}/clients"
SINGBOX_CONF="/usr/local/etc/sing-box/config.json"
SINGBOX_BIN="/usr/local/bin/sing-box"

vpn_ensure_dirs() {
  mkdir -p "${FRESHVPS_ETC}/secrets" "${VPN_CLIENTS_DIR}"
  chmod 700 "${FRESHVPS_ETC}/secrets" "${VPN_CLIENTS_DIR}"
  if [[ ! -f "${VPN_USERS_FILE}" ]]; then
    echo '{"users":[]}' >"${VPN_USERS_FILE}"
    chmod 600 "${VPN_USERS_FILE}"
  fi
}

vpn_secret() {
  local f="${FRESHVPS_ETC}/secrets/$1"
  [[ -f "${f}" ]] && cat "${f}"
}

vpn_public_ip() {
  if [[ -n "${PUBLIC_IP:-}" ]]; then
    echo "${PUBLIC_IP}"
    return
  fi
  if [[ -f "${FRESHVPS_ETC}/public_ip" ]]; then
    cat "${FRESHVPS_ETC}/public_ip"
    return
  fi
  curl -4 -fsS --max-time 5 https://ifconfig.me 2>/dev/null || echo "YOUR_IP"
}

vpn_sni() { echo "${SINGBOX_REALITY_SNI:-www.cloudflare.com}"; }
vpn_vless_port() { echo "${SINGBOX_VLESS_PORT:-443}"; }
vpn_hy2_port() { echo "${SINGBOX_HY2_PORT:-8443}"; }

vpn_vless_link() {
  local name="$1" uuid="$2"
  local ip pub sid sni port
  ip="$(vpn_public_ip)"
  pub="$(vpn_secret singbox_reality_public)"
  sid="$(vpn_secret singbox_short_id)"
  sni="$(vpn_sni)"
  port="$(vpn_vless_port)"
  # vless://uuid@host:port?encryption=none&flow=xtls-rprx-vision&security=reality&sni=&fp=chrome&pbk=&sid=&type=tcp#name
  printf 'vless://%s@%s:%s?encryption=none&flow=xtls-rprx-vision&security=reality&sni=%s&fp=chrome&pbk=%s&sid=%s&type=tcp#%s\n' \
    "${uuid}" "${ip}" "${port}" "${sni}" "${pub}" "${sid}" "${name}"
}

vpn_write_artifacts() {
  local name="$1" uuid="$2"
  local dir="${VPN_CLIENTS_DIR}/${name}"
  mkdir -p "${dir}"
  chmod 700 "${dir}"
  vpn_vless_link "${name}" "${uuid}" >"${dir}/link.txt"
  chmod 600 "${dir}/link.txt"
  if command -v qrencode >/dev/null 2>&1; then
    qrencode -o "${dir}/qr.png" -t PNG <"${dir}/link.txt" 2>/dev/null || true
    [[ -f "${dir}/qr.png" ]] && chmod 600 "${dir}/qr.png"
  fi
  cat >"${dir}/README.txt" <<EOF
FreshVPS client: ${name}
Import link.txt or qr.png into v2rayN / Streisand / Hiddify / Shadowrocket.
DNS is applied on the server (Blocky) while connected — no phone DNS setup required.
EOF
  chmod 600 "${dir}/README.txt"
}

vpn_users_json() {
  cat "${VPN_USERS_FILE}"
}

vpn_user_exists() {
  local name="$1"
  jq -e --arg n "${name}" '.users[] | select(.name==$n)' "${VPN_USERS_FILE}" >/dev/null 2>&1
}

vpn_add_user() {
  local name="$1" note="${2:-}"
  vpn_ensure_dirs
  [[ -n "${name}" ]] || { echo "name required" >&2; return 1; }
  if vpn_user_exists "${name}"; then
    echo "user already exists: ${name}" >&2
    return 1
  fi
  local uuid
  if [[ -x "${SINGBOX_BIN}" ]]; then
    uuid="$("${SINGBOX_BIN}" generate uuid 2>/dev/null || true)"
  fi
  [[ -n "${uuid}" ]] || uuid="$(cat /proc/sys/kernel/random/uuid)"
  local tmp
  tmp="$(mktemp)"
  jq --arg n "${name}" --arg u "${uuid}" --arg note "${note}" --arg ts "$(date -Iseconds)" \
    '.users += [{name:$n, uuid:$u, enabled:true, note:$note, created:$ts}]' \
    "${VPN_USERS_FILE}" >"${tmp}"
  mv "${tmp}" "${VPN_USERS_FILE}"
  chmod 600 "${VPN_USERS_FILE}"
  vpn_write_artifacts "${name}" "${uuid}"
  vpn_apply_config
  echo "${name} ${uuid}"
}

vpn_set_enabled() {
  local name="$1" en="$2"
  local tmp
  tmp="$(mktemp)"
  jq --arg n "${name}" --argjson e "${en}" \
    '(.users[] | select(.name==$n) | .enabled) = $e' \
    "${VPN_USERS_FILE}" >"${tmp}"
  mv "${tmp}" "${VPN_USERS_FILE}"
  chmod 600 "${VPN_USERS_FILE}"
  vpn_apply_config
}

vpn_revoke_user() {
  local name="$1"
  local tmp
  tmp="$(mktemp)"
  jq --arg n "${name}" '.users |= map(select(.name!=$n))' "${VPN_USERS_FILE}" >"${tmp}"
  mv "${tmp}" "${VPN_USERS_FILE}"
  chmod 600 "${VPN_USERS_FILE}"
  rm -rf "${VPN_CLIENTS_DIR}/${name}"
  vpn_apply_config
}

vpn_list_users() {
  jq -r '.users[] | "\(.name)\t\(if .enabled then "on" else "off" end)\t\(.uuid)\t\(.note // "")\t\(.created // "")"' "${VPN_USERS_FILE}"
}

vpn_build_users_array() {
  # JSON array of {uuid, flow, name} for enabled users
  jq -c '[.users[] | select(.enabled==true) | {uuid:.uuid, flow:"xtls-rprx-vision", name:.name}]' "${VPN_USERS_FILE}"
}

vpn_apply_config() {
  vpn_ensure_dirs
  local private_key public_key short_id hy2_pass sni vless_port hy2_port users_json
  private_key="$(vpn_secret singbox_reality_private)"
  public_key="$(vpn_secret singbox_reality_public)"
  short_id="$(vpn_secret singbox_short_id)"
  hy2_pass="$(vpn_secret singbox_hy2_password)"
  sni="$(vpn_sni)"
  vless_port="$(vpn_vless_port)"
  hy2_port="$(vpn_hy2_port)"
  users_json="$(vpn_build_users_array)"

  if [[ -z "${private_key}" || -z "${short_id}" ]]; then
    echo "Reality secrets missing — run sing-box module first" >&2
    return 1
  fi

  # If no users yet, keep a placeholder disabled-friendly empty array — sing-box needs at least structure
  if [[ "${users_json}" == "[]" ]]; then
    # inject temporary operator-less empty: some versions require >=1 user; create ephemeral no-op disabled by empty listen? 
    # Use a random uuid not distributed
    users_json="[{\"uuid\":\"$(cat /proc/sys/kernel/random/uuid)\",\"flow\":\"xtls-rprx-vision\",\"name\":\"_reserved\"}]"
  fi

  mkdir -p /usr/local/etc/sing-box /etc/sing-box/certs

  jq -n \
    --argjson users "${users_json}" \
    --arg priv "${private_key}" \
    --arg sid "${short_id}" \
    --arg sni "${sni}" \
    --arg hy2 "${hy2_pass}" \
    --argjson vport "${vless_port}" \
    --argjson hport "${hy2_port}" \
    '{
      log: {level:"info", timestamp:true},
      dns: {
        servers: [
          {tag:"blocky", address:"127.0.0.1", detour:"direct"},
          {tag:"local", address:"local", detour:"direct"}
        ],
        strategy: "ipv4_only",
        final: "blocky"
      },
      inbounds: [
        {
          type: "vless",
          tag: "vless-reality",
          listen: "::",
          listen_port: $vport,
          users: $users,
          tls: {
            enabled: true,
            server_name: $sni,
            reality: {
              enabled: true,
              handshake: {server: $sni, server_port: 443},
              private_key: $priv,
              short_id: [$sid]
            }
          },
          sniff: true,
          sniffer: true
        },
        {
          type: "hysteria2",
          tag: "hy2",
          listen: "::",
          listen_port: $hport,
          users: (if $hy2 != "" then [{password:$hy2}] else [] end),
          tls: {
            enabled: true,
            alpn: ["h3"],
            certificate_path: "/etc/sing-box/certs/hy2.crt",
            key_path: "/etc/sing-box/certs/hy2.key"
          },
          masquerade: ("https://" + $sni)
        }
      ],
      outbounds: [
        {type:"direct", tag:"direct"},
        {type:"dns", tag:"dns-out"}
      ],
      route: {
        rules: [
          {protocol:"dns", outbound:"dns-out"}
        ],
        final: "direct",
        auto_detect_interface: true
      }
    }' >"${SINGBOX_CONF}.new"

  # Fix possible wrong keys for sing-box version — sniff field names
  # Remove sniffer if invalid; keep sniff
  jq 'walk(if type=="object" then del(.sniffer) else . end)' "${SINGBOX_CONF}.new" >"${SINGBOX_CONF}.new2" 2>/dev/null || cp "${SINGBOX_CONF}.new" "${SINGBOX_CONF}.new2"

  if [[ -x "${SINGBOX_BIN}" ]]; then
    if ! "${SINGBOX_BIN}" check -c "${SINGBOX_CONF}.new2" 2>/dev/null; then
      # fallback simpler config without dns route if check fails
      jq 'del(.dns) | del(.route) | del(.outbounds[1])' "${SINGBOX_CONF}.new2" >"${SINGBOX_CONF}.new3" || cp "${SINGBOX_CONF}.new2" "${SINGBOX_CONF}.new3"
      if ! "${SINGBOX_BIN}" check -c "${SINGBOX_CONF}.new3" 2>/dev/null; then
        echo "sing-box config validation failed" >&2
        return 1
      fi
      mv "${SINGBOX_CONF}.new3" "${SINGBOX_CONF}"
    else
      mv "${SINGBOX_CONF}.new2" "${SINGBOX_CONF}"
    fi
  else
    mv "${SINGBOX_CONF}.new2" "${SINGBOX_CONF}"
  fi
  chmod 600 "${SINGBOX_CONF}"
  rm -f "${SINGBOX_CONF}.new" "${SINGBOX_CONF}.new2" "${SINGBOX_CONF}.new3" 2>/dev/null || true

  if systemctl is-enabled sing-box >/dev/null 2>&1 || systemctl is-active sing-box >/dev/null 2>&1; then
    systemctl restart sing-box
  fi
}

vpn_seed_operator() {
  vpn_ensure_dirs
  if ! vpn_user_exists "operator"; then
    vpn_add_user "operator" "auto-created at install"
  else
    vpn_apply_config
  fi
}
