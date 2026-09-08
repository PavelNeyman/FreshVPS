#!/usr/bin/env bash
set -euo pipefail
LANG_CODE="${FRESHVPS_LANG:-}"
if [[ -z "${LANG_CODE}" && -f /etc/freshvps/telegram_lang ]]; then
  LANG_CODE="$(tr -d '[:space:]' </etc/freshvps/telegram_lang)"
fi
[[ "${LANG_CODE}" == "en" ]] || LANG_CODE=ru

if [[ "${LANG_CODE}" == "en" ]]; then
  L_HDR="FreshVPS status on %s\n"
  L_ACT="active"
  L_INACT="inactive"
  L_VPN="\nVPN users:\n"
else
  L_HDR="Статус FreshVPS на %s\n"
  L_ACT="активен"
  L_INACT="неактивен"
  L_VPN="\nПользователи VPN:\n"
fi

printf "${L_HDR}" "$(hostname)"
for u in sing-box blocky freshvps-api freshvps-telegram-bot; do
  if systemctl is-active --quiet "${u}" 2>/dev/null; then
    printf '%s: %s\n' "${u}" "${L_ACT}"
  else
    printf '%s: %s\n' "${u}" "${L_INACT}"
  fi
done
if systemctl is-active --quiet freshvps-metrics.timer 2>/dev/null; then
  printf 'metrics-timer: %s\n' "${L_ACT}"
else
  printf 'metrics-timer: %s\n' "${L_INACT}"
fi
if command -v docker >/dev/null 2>&1; then
  printf 'docker: %s\n' "$(docker ps --format '{{.Names}}' 2>/dev/null | tr '\n' ' ')"
fi
if command -v freshvps-vpn >/dev/null 2>&1; then
  printf "${L_VPN}"
  freshvps-vpn list 2>/dev/null || true
fi
if [[ -f /var/lib/freshvps/metrics/latest.json ]]; then
  FRESHVPS_LANG="${LANG_CODE}" python3 - <<'PY' 2>/dev/null || true
import json, os
lang = os.environ.get("FRESHVPS_LANG", "ru")
d = json.load(open("/var/lib/freshvps/metrics/latest.json"))
if lang == "en":
    print("cpu: {}%  mem: {}%".format(d.get("cpu_pct"), d.get("mem_pct")))
    ok, fail, pref = "ok", "FAIL", "probe"
else:
    print("ЦП: {}%  ОЗУ: {}%".format(d.get("cpu_pct"), d.get("mem_pct")))
    ok, fail, pref = "ок", "СБОЙ", "проба"
for p in d.get("probes") or []:
    print("{} {}: {}".format(pref, p.get("name"), ok if p.get("ok") else fail))
PY
fi
