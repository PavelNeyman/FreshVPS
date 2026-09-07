#!/usr/bin/env bash
# Static checks (no root, no VPS). Run from repo root.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

fail=0
files=(
  install.sh uninstall.sh install-edge.sh install-openwrt.sh
  lib/common.sh lib/vpn-core.sh lib/vless-parse.sh
  bin/freshvps-vpn bin/freshvps-doctor
  scripts/check.sh
  runtime/telegram/bot.sh runtime/telegram/notify.sh runtime/telegram/status.sh
)
for f in modules/*.sh; do
  files+=("${f}")
done

echo "== bash -n =="
for f in "${files[@]}"; do
  if [[ -f "${f}" ]]; then
    if bash -n "${f}"; then
      echo "OK  ${f}"
    else
      echo "FAIL ${f}"
      fail=1
    fi
  fi
done

echo "== sh -n openwrt =="
for f in openwrt/install-openwrt.sh openwrt/lib/*.sh; do
  if [[ -f "${f}" ]]; then
    if sh -n "${f}"; then
      echo "OK  ${f}"
    else
      echo "FAIL ${f}"
      fail=1
    fi
  fi
done

if [[ -f runtime/api/server.py ]]; then
  echo "== python3 -m py_compile =="
  if python3 -m py_compile runtime/api/server.py; then
    echo "OK  runtime/api/server.py"
  else
    fail=1
  fi
fi

if command -v go >/dev/null 2>&1 && [[ -f cmd/freshvps-tg/main.go ]]; then
  echo "== go build =="
  if ( cd cmd/freshvps-tg && go build -o /tmp/freshvps-tg-check . ); then
    echo "OK  cmd/freshvps-tg"
  else
    fail=1
  fi
fi

if command -v shellcheck >/dev/null 2>&1; then
  echo "== shellcheck =="
  shellcheck -x install.sh uninstall.sh lib/*.sh bin/* modules/*.sh runtime/telegram/*.sh scripts/check.sh || fail=1
else
  echo "SKIP shellcheck (not installed)"
fi

exit "${fail}"
