#!/usr/bin/env bash
# FreshVPS uninstall
#
# Default: remove services, binaries, units, containers, CLI, UFW rules.
#          Keep configs + data (so reinstall can reuse them).
# --purge: also remove configs, secrets, data dirs, docker images, state.
#
set -euo pipefail

FRESHVPS_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export FRESHVPS_ROOT
# shellcheck source=lib/common.sh
source "${FRESHVPS_ROOT}/lib/common.sh"

PURGE=0
MODULES=()

usage() {
  cat <<EOF
Usage: sudo bash uninstall.sh [--purge] [module ...]

Default: stop & remove runtime (units, binaries, containers, CLI, UFW rules).
         Keeps configs and data under /etc/freshvps, /etc/blocky, panel data dirs.

  --purge   also delete configs, secrets, data, docker images, install state

Modules (default = all):
  host-tools hardening sing-box blocky vpn-users vpn-api
  opensoho kuma beszel lampac telegram backup vps-tests
EOF
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --purge|--purge-all|--purge-secrets)
        # --purge-secrets kept as alias for full purge of secrets+data+configs
        PURGE=1
        shift
        ;;
      -h|--help) usage; exit 0 ;;
      *) MODULES+=("$1"); shift ;;
    esac
  done
  if [[ ${#MODULES[@]} -eq 0 ]]; then
    MODULES=(
      vpn-api vpn-users sing-box blocky
      opensoho kuma beszel lampac telegram
      backup host-tools vps-tests hardening
    )
  fi
  export FRESHVPS_PURGE="${PURGE}"
}

ufw_delete_port() {
  local spec="$1"
  command -v ufw >/dev/null 2>&1 || return 0
  ufw status 2>/dev/null | grep -qi 'Status: active' || return 0
  ufw --force delete allow "${spec}" >/dev/null 2>&1 || true
}

run_uninstall() {
  local name="$1"
  local script="${FRESHVPS_ROOT}/modules/${name}.sh"
  [[ -f "${script}" ]] || { warn "No module file for ${name}"; return 0; }
  # shellcheck source=/dev/null
  source "${script}"
  local fn="module_${name//-/_}_uninstall"
  if declare -F "${fn}" >/dev/null 2>&1; then
    info "Uninstall ${name}"
    "${fn}"
    rm -f "${FRESHVPS_STATE_DIR}/modules/${name}.done"
  else
    warn "No uninstall hook for ${name}"
  fi
}

purge_shared() {
  [[ "${PURGE}" -eq 1 ]] || return 0
  warn "--purge: removing configs, data, images, state"

  rm -rf \
    /etc/freshvps \
    /etc/blocky \
    /usr/local/etc/sing-box \
    /etc/sing-box \
    /var/lib/sing-box \
    /var/lib/opensoho \
    /opt/uptime-kuma \
    /opt/beszel \
    /opt/lampac \
    /opt/opensoho \
    /opt/freshvps \
    /opt/freshvps-api \
    /opt/freshvps-backup \
    /var/backups/freshvps \
    "${FRESHVPS_STATE_DIR}" \
    /etc/systemd/resolved.conf.d/freshvps-no-stub.conf \
    2>/dev/null || true

  if command -v docker >/dev/null 2>&1; then
    docker rm -f opensoho uptime-kuma beszel beszel-agent lampac 2>/dev/null || true
    docker rmi -f \
      ghcr.io/opensoho/opensoho:latest \
      louislam/uptime-kuma:1 \
      henrygd/beszel:latest \
      henrygd/beszel-agent:latest \
      ghcr.io/lampac-nextgen/lampac:latest \
      2>/dev/null || true
  fi

  # Restore upstream DNS if still pointing only at localhost
  if [[ -f /etc/resolv.conf ]] && grep -qE '^nameserver[[:space:]]+127\.' /etc/resolv.conf; then
    printf 'nameserver 9.9.9.9\nnameserver 1.1.1.1\n' >/etc/resolv.conf
  fi

  info "Purge complete"
}

main() {
  parse_args "$@"
  require_root
  # ensure_dirs only if not purging everything — avoid recreating empty tree mid-purge
  mkdir -p "${FRESHVPS_STATE_DIR:-/var/lib/freshvps}/modules" 2>/dev/null || true

  for m in "${MODULES[@]}"; do
    run_uninstall "${m}"
  done

  # Shared runtime cleanup (binaries / compose leftovers not owned by one module)
  rm -f /usr/local/bin/sing-box /usr/local/bin/blocky
  rm -f /usr/local/bin/freshvps-doctor /usr/local/bin/freshvps-smoke \
    /usr/local/bin/freshvps-vpn /usr/local/bin/freshvps-tests

  ufw_delete_port "443/tcp"
  ufw_delete_port "8443/udp"
  ufw_delete_port "53/tcp"
  ufw_delete_port "53/udp"
  ufw_delete_port "8787/tcp"

  if [[ -f /opt/freshvps/compose/panels.yml ]] && command -v docker >/dev/null 2>&1; then
    (
      cd /opt/freshvps/compose
      docker compose -f panels.yml --profile opensoho --profile kuma --profile beszel \
        --profile beszel-agent --profile lampac down --remove-orphans 2>/dev/null || true
    )
  fi

  purge_shared

  if [[ "${PURGE}" -eq 0 ]]; then
    info "Uninstall finished (configs/data kept). Use --purge to wipe them too."
  else
    info "Uninstall + purge finished"
  fi
}

main "$@"
