#!/usr/bin/env bash
# Optional: start selected panels from compose/panels.yml profiles
# Used by install when ENABLE_* panels are set; modules kuma/opensoho/… still work standalone.
# shellcheck disable=SC2154

panels_ensure_docker() {
  if ! command -v docker >/dev/null 2>&1; then
    info "Installing Docker"
    pkg_install ca-certificates curl
    curl -fsSL https://get.docker.com | sh
    systemctl enable --now docker
  fi
}

panels_install_compose_file() {
  mkdir -p /opt/freshvps/compose /var/lib/opensoho /opt/uptime-kuma/data \
    /opt/beszel/hub-data /opt/beszel/agent-data /opt/beszel/socket \
    /opt/lampac/config /opt/lampac/cache /opt/lampac/database
  if [[ -f "${FRESHVPS_ROOT}/compose/panels.yml" ]]; then
    install -m 644 "${FRESHVPS_ROOT}/compose/panels.yml" /opt/freshvps/compose/panels.yml
  fi
}

# profiles: space-separated e.g. "opensoho kuma beszel"
panels_compose_up() {
  local profiles=()
  local p
  for p in "$@"; do
    [[ -n "${p}" ]] && profiles+=(--profile "${p}")
  done
  [[ ${#profiles[@]} -eq 0 ]] && return 0
  panels_ensure_docker
  panels_install_compose_file
  (
    cd /opt/freshvps/compose
    export OPENSOHO_HTTP_PORT="${OPENSOHO_HTTP_PORT:-8090}"
    export KUMA_PORT="${KUMA_PORT:-3001}"
    export BESZEL_PORT="${BESZEL_PORT:-8091}"
    export LAMPAC_PORT="${LAMPAC_PORT:-9118}"
    export OPENSOHO_SHARED_SECRET="$(read_secret opensoho_shared_secret 2>/dev/null || true)"
    export BESZEL_KEY="$(read_secret beszel_key 2>/dev/null || true)"
    export BESZEL_TOKEN="$(read_secret beszel_token 2>/dev/null || true)"
    docker compose -f panels.yml "${profiles[@]}" pull || true
    docker compose -f panels.yml "${profiles[@]}" up -d
  )
}
