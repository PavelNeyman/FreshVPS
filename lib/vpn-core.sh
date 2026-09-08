#!/usr/bin/env bash
# FreshVPS VPN user core — multi-user VLESS + HY2, subscription bundle
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
  if [[ -n "${PUBLIC_IP:-}" ]]; then echo "${PUBLIC_IP}"; return; fi
  if [[ -f "${FRESHVPS_ETC}/public_ip" ]]; then cat "${FRESHVPS_ETC}/public_ip"; return; fi
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
  printf 'vless://%s@%s:%s?encryption=none&flow=xtls-rprx-vision&security=reality&sni=%s&fp=chrome&pbk=%s&sid=%s&type=tcp#%s\n' \
    "${uuid}" "${ip}" "${port}" "${sni}" "${pub}" "${sid}" "${name}"
}

# HY2 share link (self-signed cert → insecure=1 for client convenience)
vpn_hy2_link() {
  local name="$1" hy2pass="$2"
  local ip sni port
  ip="$(vpn_public_ip)"
  sni="$(vpn_sni)"
  port="$(vpn_hy2_port)"
  printf 'hysteria2://%s@%s:%s?sni=%s&insecure=1#%s-hy2\n' \
    "${hy2pass}" "${ip}" "${port}" "${sni}" "${name}"
}

vpn_write_artifacts() {
  local name="$1" uuid="$2" hy2pass="$3"
  local dir="${VPN_CLIENTS_DIR}/${name}"
  mkdir -p "${dir}"
  chmod 700 "${dir}"
  vpn_vless_link "${name}" "${uuid}" >"${dir}/link-vless.txt"
  vpn_hy2_link "${name}" "${hy2pass}" >"${dir}/link-hy2.txt"
  # Unified handoff: both protocols (import as subscription / multi-line)
  {
    cat "${dir}/link-vless.txt"
    cat "${dir}/link-hy2.txt"
  } >"${dir}/subscription.txt"
  # Primary link.txt = VLESS (widest client support); sub = both
  cp "${dir}/link-vless.txt" "${dir}/link.txt"
  base64 -w0 "${dir}/subscription.txt" >"${dir}/subscription.b64" 2>/dev/null \
    || base64 "${dir}/subscription.txt" | tr -d '\n' >"${dir}/subscription.b64"
  chmod 600 "${dir}/"link*.txt "${dir}/subscription.txt" "${dir}/subscription.b64"
  if command -v qrencode >/dev/null 2>&1; then
    qrencode -o "${dir}/qr.png" -t PNG <"${dir}/link-vless.txt" 2>/dev/null || true
    qrencode -o "${dir}/qr-subscription.png" -t PNG <"${dir}/subscription.txt" 2>/dev/null || true
    [[ -f "${dir}/qr.png" ]] && chmod 600 "${dir}/qr.png"
  fi
  cat >"${dir}/README.txt" <<EOF
FreshVPS client: ${name}

Delivery (pick one):
1) link-vless.txt / qr.png     — VLESS+Reality (most apps)
2) link-hy2.txt               — Hysteria2 (same user, own password)
3) subscription.txt           — BOTH lines; import as subscription/clipboard
4) subscription.b64           — base64 of (3) for apps that expect base64 sub

There is no single official URI that embeds VLESS+HY2; subscription is the portable bundle.
EOF
  chmod 600 "${dir}/README.txt"
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
  local uuid hy2pass
  if [[ -x "${SINGBOX_BIN}" ]]; then
    uuid="$("${SINGBOX_BIN}" generate uuid 2>/dev/null || true)"
  fi
  [[ -n "${uuid}" ]] || uuid="$(cat /proc/sys/kernel/random/uuid)"
  hy2pass="$(openssl rand -hex 16)"
  local tmp
  tmp="$(mktemp)"
  jq --arg n "${name}" --arg u "${uuid}" --arg h "${hy2pass}" --arg note "${note}" --arg ts "$(date -Iseconds)" \
    '.users += [{name:$n, uuid:$u, hy2_password:$h, enabled:true, note:$note, created:$ts}]' \
    "${VPN_USERS_FILE}" >"${tmp}"
  mv "${tmp}" "${VPN_USERS_FILE}"
  chmod 600 "${VPN_USERS_FILE}"
  vpn_write_artifacts "${name}" "${uuid}" "${hy2pass}"
  vpn_apply_config
  echo "${name} ${uuid}"
}

vpn_set_note() {
  local name="$1" note="${2:-}"
  vpn_ensure_dirs
  vpn_user_exists "${name}" || { echo "user not found: ${name}" >&2; return 1; }
  local tmp
  tmp="$(mktemp)"
  jq --arg n "${name}" --arg note "${note}" \
    '(.users[] | select(.name==$n) | .note) = $note' \
    "${VPN_USERS_FILE}" >"${tmp}"
  mv "${tmp}" "${VPN_USERS_FILE}"
  chmod 600 "${VPN_USERS_FILE}"
}

vpn_set_enabled() {
  local name="$1" en="$2"
  local tmp; tmp="$(mktemp)"
  jq --arg n "${name}" --argjson e "${en}" \
    '(.users[] | select(.name==$n) | .enabled) = $e' \
    "${VPN_USERS_FILE}" >"${tmp}"
  mv "${tmp}" "${VPN_USERS_FILE}"
  chmod 600 "${VPN_USERS_FILE}"
  vpn_apply_config
}

vpn_revoke_user() {
  local name="$1"
  local tmp; tmp="$(mktemp)"
  jq --arg n "${name}" '.users |= map(select(.name!=$n))' "${VPN_USERS_FILE}" >"${tmp}"
  mv "${tmp}" "${VPN_USERS_FILE}"
  chmod 600 "${VPN_USERS_FILE}"
  rm -rf "${VPN_CLIENTS_DIR}/${name}"
  vpn_apply_config
}

vpn_list_users() {
  jq -r '.users[] | "\(.name)\t\(if .enabled then "on" else "off" end)\t\(.uuid)\t\(.note // "")\t\(.created // "")"' "${VPN_USERS_FILE}"
}

vpn_build_vless_users() {
  jq -c '[.users[] | select(.enabled==true) | {uuid:.uuid, flow:"xtls-rprx-vision"}]' "${VPN_USERS_FILE}"
}

vpn_build_hy2_users() {
  # name + password per user (sing-box hysteria2)
  jq -c '[.users[] | select(.enabled==true) | {name:.name, password:(.hy2_password // .uuid)}]' "${VPN_USERS_FILE}"
}

vpn_try_check() {
  local conf="$1"
  [[ -x "${SINGBOX_BIN}" ]] || return 0
  "${SINGBOX_BIN}" check -c "${conf}" >/dev/null 2>&1
}

vpn_apply_config() {
  vpn_ensure_dirs
  local private_key short_id sni vless_port hy2_port vless_users hy2_users
  private_key="$(vpn_secret singbox_reality_private)"
  short_id="$(vpn_secret singbox_short_id)"
  sni="$(vpn_sni)"
  vless_port="$(vpn_vless_port)"
  hy2_port="$(vpn_hy2_port)"
  vless_users="$(vpn_build_vless_users)"
  hy2_users="$(vpn_build_hy2_users)"

  if [[ -z "${private_key}" || -z "${short_id}" ]]; then
    echo "Reality secrets missing" >&2
    return 1
  fi

  # No enabled users → still valid empty listen (no fake UUID)
  if [[ "${vless_users}" == "[]" ]]; then
    echo "no enabled VPN users — writing empty inbounds users arrays" >&2
  fi

  mkdir -p /usr/local/etc/sing-box /etc/sing-box/certs
  local work; work="$(mktemp -d)"

  jq -n \
    --argjson vusers "${vless_users}" \
    --argjson husers "${hy2_users}" \
    --arg priv "${private_key}" \
    --arg sid "${short_id}" \
    --arg sni "${sni}" \
    --argjson vport "${vless_port}" \
    --argjson hport "${hy2_port}" \
    '{
      log: {level:"info", timestamp:true},
      dns: {
        servers: [
          {type:"udp", tag:"blocky", server:"127.0.0.1"},
          {type:"local", tag:"local"}
        ],
        final: "blocky",
        strategy: "ipv4_only"
      },
      inbounds: [
        {
          type: "vless",
          tag: "vless-reality",
          listen: "::",
          listen_port: $vport,
          users: $vusers,
          tls: {
            enabled: true,
            server_name: $sni,
            reality: {
              enabled: true,
              handshake: {server: $sni, server_port: 443},
              private_key: $priv,
              short_id: [$sid]
            }
          }
        },
        {
          type: "hysteria2",
          tag: "hy2",
          listen: "::",
          listen_port: $hport,
          users: $husers,
          tls: {
            enabled: true,
            alpn: ["h3"],
            certificate_path: "/etc/sing-box/certs/hy2.crt",
            key_path: "/etc/sing-box/certs/hy2.key"
          },
          masquerade: ("https://" + $sni)
        }
      ],
      outbounds: [ {type:"direct", tag:"direct"} ],
      route: {
        rules: [
          {action:"sniff"},
          {protocol:"dns", action:"hijack-dns"}
        ],
        final: "direct",
        auto_detect_interface: true,
        default_domain_resolver: "blocky"
      }
    }' >"${work}/a.json"

  jq 'del(.dns) | .route.default_domain_resolver="local" | .dns={servers:[{type:"local",tag:"local"}],final:"local"}' \
    "${work}/a.json" >"${work}/b.json" 2>/dev/null || cp "${work}/a.json" "${work}/b.json"

  jq 'del(.dns) | .route={final:"direct",auto_detect_interface:true}' \
    "${work}/a.json" >"${work}/c.json" 2>/dev/null || true

  local chosen="" variant=""
  for v in a b c; do
    if vpn_try_check "${work}/${v}.json"; then
      chosen="${work}/${v}.json"; variant="${v}"; break
    fi
  done

  if [[ -z "${chosen}" ]]; then
    echo "sing-box config validation failed for all variants" >&2
    rm -rf "${work}"
    return 1
  fi

  install -m 600 "${chosen}" "${SINGBOX_CONF}"
  printf '%s\n' "${variant}" >"${FRESHVPS_ETC}/singbox-config-variant"
  rm -rf "${work}"

  if systemctl is-enabled sing-box >/dev/null 2>&1 || systemctl is-active sing-box >/dev/null 2>&1; then
    systemctl restart sing-box
  fi
  echo "sing-box config variant=${variant}" >&2
}

vpn_seed_operator() {
  vpn_ensure_dirs
  if ! vpn_user_exists "operator"; then
    vpn_add_user "operator" "auto-created at install"
  else
    # migrate: ensure hy2_password exists
    if jq -e '.users[] | select(.name=="operator" and (.hy2_password|not))' "${VPN_USERS_FILE}" >/dev/null 2>&1; then
      local tmp h; h="$(openssl rand -hex 16)"; tmp="$(mktemp)"
      jq --arg h "${h}" '(.users[] | select(.name=="operator") | .hy2_password) = $h' "${VPN_USERS_FILE}" >"${tmp}"
      mv "${tmp}" "${VPN_USERS_FILE}"
      chmod 600 "${VPN_USERS_FILE}"
      local u; u="$(jq -r '.users[] | select(.name=="operator") | .uuid' "${VPN_USERS_FILE}")"
      vpn_write_artifacts "operator" "${u}" "${h}"
    fi
    vpn_apply_config
  fi
}
