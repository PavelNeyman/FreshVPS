#!/usr/bin/env bash
set -euo pipefail
printf 'FreshVPS status on %s\n' "$(hostname)"
for u in sing-box blocky freshvps-api freshvps-telegram-bot; do
  if systemctl is-active --quiet "${u}" 2>/dev/null; then
    printf '%s: active\n' "${u}"
  else
    printf '%s: inactive\n' "${u}"
  fi
done
if systemctl is-active --quiet freshvps-metrics.timer 2>/dev/null; then
  printf 'metrics-timer: active\n'
else
  printf 'metrics-timer: inactive\n'
fi
if command -v docker >/dev/null 2>&1; then
  printf 'docker: %s\n' "$(docker ps --format '{{.Names}}' 2>/dev/null | tr '\n' ' ')"
fi
if command -v freshvps-vpn >/dev/null 2>&1; then
  printf '\nVPN users:\n'
  freshvps-vpn list 2>/dev/null || true
fi
if [[ -f /var/lib/freshvps/metrics/latest.json ]]; then
  python3 - <<'PY' 2>/dev/null || true
import json
d=json.load(open("/var/lib/freshvps/metrics/latest.json"))
print("cpu: {}%  mem: {}%".format(d.get("cpu_pct"), d.get("mem_pct")))
for p in d.get("probes") or []:
    print("probe {}: {}".format(p.get("name"), "ok" if p.get("ok") else "FAIL"))
PY
fi
