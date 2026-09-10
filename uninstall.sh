#!/usr/bin/env bash
# Netductor uninstall helper
set -euo pipefail
[[ ${EUID:-} -eq 0 ]] || { echo "root required"; exit 1; }
systemctl disable --now netductor-api netductor-metrics.timer netductor-telegram-bot sing-box blocky 2>/dev/null || true
systemctl disable --now freshvps-api freshvps-metrics.timer freshvps-telegram-bot 2>/dev/null || true
if [[ "${1:-}" == "--purge" ]]; then
  rm -rf /etc/netductor /var/lib/netductor /opt/netductor
  rm -rf /etc/freshvps /var/lib/freshvps /opt/freshvps
  rm -f /usr/local/bin/netductor /usr/local/bin/netductor-tg
  echo "purged"
else
  echo "units stopped (configs kept). Use --purge to delete data."
fi
