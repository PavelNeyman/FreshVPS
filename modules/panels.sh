#!/usr/bin/env bash
# Unified Docker Compose for optional panels (profiles)
# shellcheck disable=SC2154

panels_ensure_docker() {
  if ! command -v docker >/dev/null 2>&1; then
    info "Installing Docker"
    pkg_install ca-certificates curl
    curl -fsSL https://get.docker.com | sh
    systemctl enable --now docker
  fi
  if ! docker compose version >/dev/null 2>&1; then
    pkg_install docker-compose-plugin 2>/dev/null || true
  fi
}

panels_install_compose_file() {
  mkdir -p /opt/freshvps/compose /var/lib/opensoho /opt/uptime-kuma/data \
    /opt/beszel/hub-data /opt/beszel/agent-data /opt/beszel/socket \
    /opt/lampac/config /opt/lampac/cache /opt/lampac/database
  if [[ -f "${FRESHVPS_ROOT}/compose/panels.yml" ]]; then
    install -m 644 "${FRESHVPS_ROOT}/compose/panels.yml" /opt/freshvps/compose/panels.yml
  elif [[ ! -f /opt/freshvps/compose/panels.yml ]]; then
    die "panels.yml missing"
  fi
}

panels_write_env() {
  local envf=/opt/freshvps/compose/.env
  umask 077
  {
    echo "OPENSOHO_HTTP_PORT=${OPENSOHO_HTTP_PORT:-8090}"
    echo "KUMA_PORT=${KUMA_PORT:-3001}"
    echo "BESZEL_PORT=${BESZEL_PORT:-8091}"
    echo "LAMPAC_PORT=${LAMPAC_PORT:-9118}"
    echo "LAMPAC_MEM_LIMIT=${LAMPAC_MEM_LIMIT:-1536m}"
    echo "OPENSOHO_IMAGE=${OPENSOHO_IMAGE:-ghcr.io/opensoho/opensoho:latest}"
    echo "LAMPAC_IMAGE=${LAMPAC_IMAGE:-ghcr.io/lampac-nextgen/lampac:latest}"
    echo "OPENSOHO_SHARED_SECRET=$(read_secret opensoho_shared_secret 2>/dev/null || true)"
    echo "BESZEL_KEY=$(read_secret beszel_key 2>/dev/null || true)"
    echo "BESZEL_TOKEN=$(read_secret beszel_token 2>/dev/null || true)"
  } >"${envf}"
  chmod 600 "${envf}"
}

# panels_compose_up profile [profile ...]
panels_compose_up() {
  local profiles=() p
  for p in "$@"; do
    [[ -n "${p}" ]] && profiles+=(--profile "${p}")
  done
  [[ ${#profiles[@]} -eq 0 ]] && return 0
  panels_ensure_docker
  panels_install_compose_file
  panels_write_env
  (
    cd /opt/freshvps/compose
    docker compose --env-file .env -f panels.yml "${profiles[@]}" pull || true
    docker compose --env-file .env -f panels.yml "${profiles[@]}" up -d
  )
}

panels_compose_down_profile() {
  local p="$1"
  [[ -f /opt/freshvps/compose/panels.yml ]] || return 0
  (
    cd /opt/freshvps/compose
    docker compose --env-file .env -f panels.yml --profile "${p}" stop 2>/dev/null || true
  )
}
