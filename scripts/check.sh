#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"
fail=0
files=(
  install.sh uninstall.sh install-edge.sh install-openwrt.sh bootstrap.sh
  lib/common.sh lib/vpn-core.sh lib/vless-parse.sh lib/tui.sh lib/idempotent.sh
  lib/install-singbox.sh lib/install-conf.sh
  bin/freshvps-vpn bin/freshvps-doctor
  scripts/check.sh scripts/vps-tests.sh scripts/plan.sh scripts/plan-deploy-targets.sh scripts/smoke-host.sh
)
for f in modules/*.sh; do files+=("${f}"); done
echo "== bash -n =="
for f in "${files[@]}"; do
  [[ -f "${f}" ]] || continue
  if bash -n "${f}"; then echo "OK  ${f}"; else echo "FAIL ${f}"; fail=1; fi
done
for f in openwrt/install-openwrt.sh openwrt/lib/*.sh; do
  [[ -f "${f}" ]] || continue
  if sh -n "${f}"; then echo "OK  ${f}"; else echo "FAIL ${f}"; fail=1; fi
done
[[ -f runtime/api/server.py ]] && python3 -m py_compile runtime/api/server.py && echo "OK  api" || true
exit "${fail}"
