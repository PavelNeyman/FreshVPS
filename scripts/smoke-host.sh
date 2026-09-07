#!/usr/bin/env bash
# Run on the VPS after install — automated slice of docs/SMOKE.md
set -euo pipefail
fail=0
run() {
  local n="$1"; shift
  if "$@" >/dev/null 2>&1; then echo "OK  ${n}"; else echo "FAIL ${n}"; fail=$((fail+1)); fi
}
echo "FreshVPS smoke-host $(date -Iseconds)"
run installed_version test -f /var/lib/freshvps/installed_version
run READY test -f /etc/freshvps/READY.txt
run vpn_json test -f /etc/freshvps/vpn-users.json
run operator_sub test -f /etc/freshvps/clients/operator/subscription.txt
run singbox_bin test -x /usr/local/bin/sing-box
run singbox_unit systemctl is-active sing-box
run singbox_check /usr/local/bin/sing-box check -c /usr/local/etc/sing-box/config.json
run blocky_unit systemctl is-active blocky
run dig dig +time=2 +tries=1 @127.0.0.1 example.com +short
run cli command -v freshvps-vpn
run doctor command -v freshvps-doctor
if command -v freshvps-doctor >/dev/null 2>&1; then
  set +e
  freshvps-doctor
  d=$?
  set -e
  [[ "${d}" -eq 0 ]] && echo "OK  doctor" || { echo "FAIL doctor"; fail=$((fail+1)); }
fi
echo "smoke-host summary fail=${fail}"
exit "${fail}"
