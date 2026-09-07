#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"
fail=0
files=(
  install.sh uninstall.sh install-edge.sh install-openwrt.sh bootstrap.sh
  lib/common.sh lib/vpn-core.sh lib/vless-parse.sh lib/tui.sh lib/idempotent.sh
  bin/freshvps-vpn bin/freshvps-doctor
  scripts/check.sh scripts/vps-tests.sh scripts/plan.sh
  runtime/telegram/bot.sh runtime/telegram/notify.sh runtime/telegram/status.sh
)
for f in modules/*.sh; do files+=("${f}"); done
echo "== bash -n =="
for f in "${files[@]}"; do
  [[ -f "${f}" ]] || continue
  if bash -n "${f}"; then echo "OK  ${f}"; else echo "FAIL ${f}"; fail=1; fi
done
echo "== sh -n openwrt =="
for f in openwrt/install-openwrt.sh openwrt/lib/*.sh; do
  [[ -f "${f}" ]] || continue
  if sh -n "${f}"; then echo "OK  ${f}"; else echo "FAIL ${f}"; fail=1; fi
done
if [[ -f runtime/api/server.py ]]; then
  python3 -m py_compile runtime/api/server.py && echo "OK  runtime/api/server.py" || fail=1
fi
if command -v go >/dev/null 2>&1 && [[ -f cmd/freshvps-tg/main.go ]]; then
  ( cd cmd/freshvps-tg && go build -o /tmp/freshvps-tg-check . ) && echo "OK  go bot" || fail=1
fi
exit "${fail}"
