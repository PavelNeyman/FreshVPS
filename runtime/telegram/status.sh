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
if [[ -f /etc/freshvps/singbox-config-variant ]]; then
  out+="sing-box config variant: $(cat /etc/freshvps/singbox-config-variant)\n"
fi
echo -e "${out}"
