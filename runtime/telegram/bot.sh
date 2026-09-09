#!/usr/bin/env bash
# DEPRECATED: use netductor-tg (Go). Kept as emergency fallback only.
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
/session [hours]
/ready
Shortcuts: docs/SHORTCUT-IOS.md'

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
    [[ -n "${text}" ]] || continue

    cmd="$(echo "${text}" | awk '{print $1}')"
    arg1="$(echo "${text}" | awk '{print $2}')"
    # note = everything after second field
    note="$(echo "${text}" | sed -E 's/^[^ ]+ +[^ ]+ ?//')"
    [[ "${note}" == "${text}" ]] && note=""

    case "${cmd}" in
      /start|/help) send "${chat}" "${HELP}" ;;
      /status) send "${chat}" "$(/opt/freshvps/runtime/telegram/status.sh)" ;;
      /vpn_list) send "${chat}" "$(freshvps-vpn list 2>&1 || echo none)" ;;
      /vpn_add)
        if [[ -z "${arg1}" ]]; then send "${chat}" "usage: /vpn_add name [note]"; continue; fi
        out="$(freshvps-vpn add "${arg1}" "${note}" 2>&1)" || true
        send "${chat}" "${out}"
        [[ -f "/etc/freshvps/clients/${arg1}/qr.png" ]] && send_photo "${chat}" "/etc/freshvps/clients/${arg1}/qr.png" "${arg1}"
        [[ -f "/etc/freshvps/clients/${arg1}/link.txt" ]] && send "${chat}" "$(cat "/etc/freshvps/clients/${arg1}/link.txt")"
        ;;
      /vpn_link)
        if [[ -z "${arg1}" ]]; then send "${chat}" "usage: /vpn_link name"; continue; fi
        if [[ -f "/etc/freshvps/clients/${arg1}/link.txt" ]]; then
          send "${chat}" "$(cat "/etc/freshvps/clients/${arg1}/link.txt")"
          [[ -f "/etc/freshvps/clients/${arg1}/qr.png" ]] && send_photo "${chat}" "/etc/freshvps/clients/${arg1}/qr.png" "${arg1}"
        else
          send "${chat}" "not found"
        fi
        ;;
      /vpn_disable)
        [[ -n "${arg1}" ]] || { send "${chat}" "usage: /vpn_disable name"; continue; }
        send "${chat}" "$(freshvps-vpn disable "${arg1}" 2>&1)"
        ;;
      /vpn_enable)
        [[ -n "${arg1}" ]] || { send "${chat}" "usage: /vpn_enable name"; continue; }
        send "${chat}" "$(freshvps-vpn enable "${arg1}" 2>&1)"
        ;;
      /vpn_revoke)
        [[ -n "${arg1}" ]] || { send "${chat}" "usage: /vpn_revoke name"; continue; }
        send "${chat}" "$(freshvps-vpn revoke "${arg1}" 2>&1)"
        ;;
      /session)
        hours="${arg1:-72}"
        tok="$(freshvps-vpn session "${hours}" 2>/dev/null | head -1)"
        send "${chat}" "Session (${hours}h) — local file only:\n${tok}"
        ;;
      /ready)
        if [[ -f /etc/freshvps/READY.txt ]]; then
          send "${chat}" "$(head -c 3500 /etc/freshvps/READY.txt)"
        else
          send "${chat}" "no READY.txt yet"
        fi
        ;;
      *) send "${chat}" "Unknown. /help" ;;
    esac
  done
done
