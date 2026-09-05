#!/usr/bin/env bash
# Module: Telegram operator panel (no end-user chat)
# shellcheck disable=SC2154

module_telegram_install() {
  mkdir -p /opt/freshvps-telegram
  pkg_install jq

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

  cat >/opt/freshvps-telegram/status.sh <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
out="FreshVPS status on $(hostname)\n"
for u in sing-box blocky opensoho freshvps-api freshvps-telegram-bot; do
  if systemctl is-active --quiet "${u}" 2>/dev/null; then
    out+="${u}: active\n"
  else
    out+="${u}: inactive\n"
  fi
done
if command -v docker >/dev/null 2>&1; then
  out+="docker: $(docker ps --format '{{.Names}}' 2>/dev/null | tr '\n' ' ')\n"
fi
if command -v freshvps-vpn >/dev/null 2>&1; then
  out+="\nVPN users:\n$(freshvps-vpn list 2>/dev/null || true)\n"
fi
echo -e "${out}"
EOF
  chmod 700 /opt/freshvps-telegram/status.sh

  cat >/opt/freshvps-telegram/bot.sh <<'EOF'
#!/usr/bin/env bash
# Operator-only long-poll bot
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

send_photo() {
  local chat="$1" file="$2" cap="${3:-}"
  [[ -f "${file}" ]] || return 0
  curl -fsS -X POST "${API}/sendPhoto" -F chat_id="${chat}" -F photo=@"${file}" -F caption="${cap}" >/dev/null || true
}

HELP='FreshVPS operator bot
/status
/vpn_list
/vpn_add <name> [note]
/vpn_link <name>
/vpn_disable <name>
/vpn_enable <name>
/vpn_revoke <name>
/session [hours] — API session token for Shortcuts
/shortcut — how to import Admin Shortcut
/ready — path to install summary'

while true; do
  resp="$(curl -fsS "${API}/getUpdates?timeout=30&offset=${OFFSET}" || true)"
  [[ -n "${resp}" ]] || { sleep 2; continue; }
  mapfile -t lines < <(echo "${resp}" | jq -c '.result[]?' 2>/dev/null || true)
  for line in "${lines[@]:-}"; do
    [[ -z "${line}" ]] && continue
    OFFSET="$(echo "${line}" | jq -r '.update_id + 1')"
    chat="$(echo "${line}" | jq -r '.message.chat.id // empty')"
    text="$(echo "${line}" | jq -r '.message.text // empty')"
    [[ "${chat}" == "${ADMIN}" ]] || continue
    set -- ${text}
    cmd="${1:-}"
    case "${cmd}" in
      /start|/help) send "${chat}" "${HELP}" ;;
      /status) send "${chat}" "$(/opt/freshvps-telegram/status.sh)" ;;
      /vpn_list) send "${chat}" "$(freshvps-vpn list 2>&1 || echo none)" ;;
      /vpn_add)
        name="${2:-}"; note="${3:-}"
        if [[ -z "${name}" ]]; then send "${chat}" "usage: /vpn_add name [note]"; continue; fi
        out="$(freshvps-vpn add "${name}" "${note}" 2>&1)" || true
        send "${chat}" "${out}"
        [[ -f "/etc/freshvps/clients/${name}/qr.png" ]] && send_photo "${chat}" "/etc/freshvps/clients/${name}/qr.png" "${name}"
        [[ -f "/etc/freshvps/clients/${name}/link.txt" ]] && send "${chat}" "$(cat "/etc/freshvps/clients/${name}/link.txt")"
        ;;
      /vpn_link)
        name="${2:-}"
        [[ -n "${name}" ]] || { send "${chat}" "usage: /vpn_link name"; continue; }
        if [[ -f "/etc/freshvps/clients/${name}/link.txt" ]]; then
          send "${chat}" "$(cat "/etc/freshvps/clients/${name}/link.txt")"
          [[ -f "/etc/freshvps/clients/${name}/qr.png" ]] && send_photo "${chat}" "/etc/freshvps/clients/${name}/qr.png" "${name}"
        else
          send "${chat}" "not found"
        fi
        ;;
      /vpn_disable) freshvps-vpn disable "${2:-}" 2>&1 | while read -r l; do send "${chat}" "$l"; done ;;
      /vpn_enable) freshvps-vpn enable "${2:-}" 2>&1 | while read -r l; do send "${chat}" "$l"; done ;;
      /vpn_revoke) freshvps-vpn revoke "${2:-}" 2>&1 | while read -r l; do send "${chat}" "$l"; done ;;
      /session)
        hours="${2:-72}"
        tok="$(freshvps-vpn session "${hours}" 2>/dev/null | head -1)"
        send "${chat}" "Session (${hours}h) — store on iPhone local file only:\n${tok}"
        ;;
      /shortcut)
        send "${chat}" "See docs/SHORTCUT-IOS.md on the server repo. After VPN is up: open import URL from /opt/freshvps/shortcuts/IMPORT.txt if present, or build once on iOS following that doc."
        [[ -f /opt/freshvps/shortcuts/IMPORT.txt ]] && send "${chat}" "$(cat /opt/freshvps/shortcuts/IMPORT.txt)"
        ;;
      /ready)
        [[ -f /etc/freshvps/READY.txt ]] && send "${chat}" "$(head -c 3500 /etc/freshvps/READY.txt)" || send "${chat}" "no READY.txt yet"
        ;;
      *) send "${chat}" "Unknown. /help" ;;
    esac
  done
done
EOF
  chmod 700 /opt/freshvps-telegram/bot.sh

  cat >/etc/systemd/system/freshvps-telegram-bot.service <<'EOF'
[Unit]
Description=FreshVPS Telegram operator bot
After=network-online.target

[Service]
Type=simple
ExecStart=/opt/freshvps-telegram/bot.sh
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

  if [[ -f "${FRESHVPS_ETC}/secrets/telegram_bot_token" && -f "${FRESHVPS_ETC}/secrets/telegram_admin_id" ]]; then
    systemd_enable_start freshvps-telegram-bot
    /opt/freshvps-telegram/notify.sh "FreshVPS: operator bot online on $(hostname)"
  else
    info "Telegram secrets incomplete — unit installed, not started"
  fi
}

module_telegram_uninstall() {
  systemctl disable --now freshvps-telegram-bot 2>/dev/null || true
  rm -f /etc/systemd/system/freshvps-telegram-bot.service
  systemctl daemon-reload 2>/dev/null || true
}
