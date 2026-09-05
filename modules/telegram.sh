#!/usr/bin/env bash
# Module: Telegram bot scaffold (config + helper script)
# Full bot logic can grow; v1 stores credentials and a status helper.
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
# Usage: notify.sh "message text"
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
out="FreshVPS status\n"
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
/opt/freshvps-telegram/notify.sh "$(echo -e "${out}")"
EOF
  chmod 700 /opt/freshvps-telegram/status.sh

  if [[ -n "${TELEGRAM_BOT_TOKEN:-}" ]]; then
    /opt/freshvps-telegram/notify.sh "FreshVPS: Telegram notify configured on $(hostname)"
  else
    info "Telegram token empty — put token in ${FRESHVPS_ETC}/secrets/telegram_bot_token later"
  fi
  info "Helpers: /opt/freshvps-telegram/notify.sh and status.sh"
}
