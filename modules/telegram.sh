#!/usr/bin/env bash
# Module: Telegram notify + simple admin bot (status / vpn export)
# shellcheck disable=SC2154

module_telegram_install() {
  mkdir -p /opt/freshvps-telegram

  if [[ -n "${TELEGRAM_BOT_TOKEN:-}" ]]; then
    write_secret telegram_bot_token "${TELEGRAM_BOT_TOKEN}"
  fi
  if [[ -n "${TELEGRAM_ADMIN_ID:-}" ]]; then
    write_secret telegram_admin_id "${TELEGRAM_ADMIN_ID}"
  fi

  cat >/opt/freshvps-telegram/notify.sh <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
TOKEN_FILE=/etc/freshvps/secrets/telegram_bot_token
CHAT_FILE=/etc/freshvps/secrets/telegram_admin_id
[[ -f "${TOKEN_FILE}" && -f "${CHAT_FILE}" ]] || exit 0
TOKEN="$(cat "${TOKEN_FILE}")"
CHAT="$(cat "${CHAT_FILE}")"
MSG="${1:-FreshVPS notification}"
curl -fsS -X POST "https://api.telegram.org/bot${TOKEN}/sendMessage" \
  -d chat_id="${CHAT}" \
  --data-urlencode text="${MSG}" >/dev/null || true
EOF
  chmod 700 /opt/freshvps-telegram/notify.sh

  cat >/opt/freshvps-telegram/vpn-export.sh <<'EOF'
#!/usr/bin/env bash
# Print client connection hints (no full share links with secrets in logs by default)
set -euo pipefail
S=/etc/freshvps/secrets
IP="${PUBLIC_IP:-$(curl -4 -fsS --max-time 3 https://ifconfig.me 2>/dev/null || echo YOUR_IP)}"
UUID="$(cat ${S}/singbox_uuid 2>/dev/null || echo missing)"
PUB="$(cat ${S}/singbox_reality_public 2>/dev/null || echo missing)"
SID="$(cat ${S}/singbox_short_id 2>/dev/null || echo missing)"
HY2="$(cat ${S}/singbox_hy2_password 2>/dev/null || echo missing)"
SNI="${SINGBOX_REALITY_SNI:-www.cloudflare.com}"
VLESS_PORT="${SINGBOX_VLESS_PORT:-443}"
HY2_PORT="${SINGBOX_HY2_PORT:-8443}"
cat <<OUT
FreshVPS VPN parameters
Server: ${IP}

VLESS + Reality
  port: ${VLESS_PORT}
  uuid: ${UUID}
  flow: xtls-rprx-vision
  SNI / server_name: ${SNI}
  public_key: ${PUB}
  short_id: ${SID}

Hysteria2
  port: ${HY2_PORT}
  password: ${HY2}
  (self-signed TLS — allow insecure or pin /etc/sing-box/certs/hy2.crt)
OUT
EOF
  chmod 700 /opt/freshvps-telegram/vpn-export.sh

  cat >/opt/freshvps-telegram/status.sh <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
out="FreshVPS status on $(hostname)\n"
for u in sing-box blocky opensoho; do
  if systemctl is-active --quiet "${u}" 2>/dev/null; then
    out+="${u}: active\n"
  else
    out+="${u}: inactive\n"
  fi
done
if command -v docker >/dev/null 2>&1; then
  out+="docker: $(docker ps --format '{{.Names}}' 2>/dev/null | tr '\n' ' ')\n"
fi
echo -e "${out}"
EOF
  chmod 700 /opt/freshvps-telegram/status.sh

  cat >/opt/freshvps-telegram/bot.sh <<'EOF'
#!/usr/bin/env bash
# Minimal long-polling bot for a single admin chat.
# Commands: /status /vpn /help
set -euo pipefail
TOKEN_FILE=/etc/freshvps/secrets/telegram_bot_token
CHAT_FILE=/etc/freshvps/secrets/telegram_admin_id
[[ -f "${TOKEN_FILE}" && -f "${CHAT_FILE}" ]] || exit 1
TOKEN="$(cat "${TOKEN_FILE}")"
ADMIN="$(cat "${CHAT_FILE}")"
OFFSET=0
API="https://api.telegram.org/bot${TOKEN}"

send() {
  curl -fsS -X POST "${API}/sendMessage" -d chat_id="$1" --data-urlencode text="$2" >/dev/null || true
}

while true; do
  resp="$(curl -fsS "${API}/getUpdates?timeout=30&offset=${OFFSET}" || true)"
  [[ -n "${resp}" ]] || { sleep 2; continue; }
  # shellcheck disable=SC2207
  mapfile -t lines < <(echo "${resp}" | jq -c '.result[]?' 2>/dev/null || true)
  for line in "${lines[@]:-}"; do
    [[ -z "${line}" ]] && continue
    OFFSET="$(echo "${line}" | jq -r '.update_id + 1')"
    chat="$(echo "${line}" | jq -r '.message.chat.id // empty')"
    text="$(echo "${line}" | jq -r '.message.text // empty')"
    [[ "${chat}" == "${ADMIN}" ]] || continue
    case "${text}" in
      /start|/help)
        send "${chat}" "Commands: /status /vpn /help"
        ;;
      /status)
        send "${chat}" "$(/opt/freshvps-telegram/status.sh)"
        ;;
      /vpn)
        send "${chat}" "$(PUBLIC_IP=${PUBLIC_IP:-} /opt/freshvps-telegram/vpn-export.sh)"
        ;;
    esac
  done
done
EOF
  chmod 700 /opt/freshvps-telegram/bot.sh

  cat >/etc/systemd/system/freshvps-telegram-bot.service <<'EOF'
[Unit]
Description=FreshVPS Telegram admin bot
After=network-online.target

[Service]
Type=simple
EnvironmentFile=-/etc/freshvps/telegram-bot.env
ExecStart=/opt/freshvps-telegram/bot.sh
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

  if [[ -f "${FRESHVPS_ETC}/secrets/telegram_bot_token" && -f "${FRESHVPS_ETC}/secrets/telegram_admin_id" ]]; then
    pkg_install jq
    systemd_enable_start freshvps-telegram-bot
    /opt/freshvps-telegram/notify.sh "FreshVPS: Telegram bot online on $(hostname)"
  else
    info "Telegram secrets incomplete — bot unit installed but not started"
  fi
  info "Helpers: notify.sh status.sh vpn-export.sh bot.sh under /opt/freshvps-telegram"
}

module_telegram_uninstall() {
  systemctl disable --now freshvps-telegram-bot 2>/dev/null || true
  rm -f /etc/systemd/system/freshvps-telegram-bot.service
  systemctl daemon-reload 2>/dev/null || true
  info "Telegram bot stopped (scripts left in /opt/freshvps-telegram)"
}
