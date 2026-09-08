#!/usr/bin/env bash
# Module: Uptime Kuma via compose/panels.yml profile kuma
# shellcheck disable=SC2154

module_kuma_install() {
  local port="${KUMA_PORT:-3001}"
  # shellcheck source=/dev/null
  source "${FRESHVPS_ROOT}/modules/panels.sh"
  mkdir -p /opt/uptime-kuma/data
  panels_compose_up kuma

  cat >/opt/uptime-kuma/SEED-MONITORS.md <<EOF
Uptime Kuma — create admin once in the browser.

1. ssh -L ${port}:127.0.0.1:${port} root@VPS
2. http://127.0.0.1:${port}
3. Optional monitors: TCP :443 (VLESS), UDP HY2, Blocky
EOF
  info "Uptime Kuma: 127.0.0.1:${port} (unified compose profile kuma)"
}

module_kuma_uninstall() {
  if [[ -f /opt/freshvps/compose/panels.yml ]]; then
    (cd /opt/freshvps/compose && docker compose -f panels.yml --profile kuma rm -sf 2>/dev/null) || true
  fi
  docker rm -f uptime-kuma 2>/dev/null || true
  info "Uptime Kuma container removed (data kept under /opt/uptime-kuma/data unless --purge)"
}
