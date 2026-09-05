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
