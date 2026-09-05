#!/usr/bin/env bash
# FreshVPS uninstall / disable hooks
set -euo pipefail

FRESHVPS_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export FRESHVPS_ROOT
# shellcheck source=lib/common.sh
source "${FRESHVPS_ROOT}/lib/common.sh"

PURGE_SECRETS=0
MODULES=()

usage() {
  cat <<EOF
FreshVPS uninstall

Usage: sudo bash uninstall.sh [options] [module ...]

  module            One of: hardening sing-box blocky opensoho kuma beszel telegram backup
                    If omitted, uninstalls all known modules (except hardening packages).
  --purge-secrets   Also remove /etc/freshvps/secrets (destructive)
  -h, --help        Help

Does not purge system packages from hardening (fail2ban/ufw) unless you remove them manually.
EOF
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --purge-secrets) PURGE_SECRETS=1; shift ;;
      -h|--help) usage; exit 0 ;;
      *) MODULES+=("$1"); shift ;;
    esac
  done
  if [[ ${#MODULES[@]} -eq 0 ]]; then
    MODULES=(sing-box blocky opensoho kuma beszel telegram backup)
  fi
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

main() {
  parse_args "$@"
  require_root
  ensure_dirs
  for m in "${MODULES[@]}"; do
    run_uninstall "${m}"
  done
  if [[ "${PURGE_SECRETS}" -eq 1 ]]; then
    warn "Removing ${FRESHVPS_ETC}/secrets"
    rm -rf "${FRESHVPS_ETC}/secrets"
    mkdir -p "${FRESHVPS_ETC}/secrets"
    chmod 700 "${FRESHVPS_ETC}/secrets"
  fi
  info "Uninstall finished"
}

main "$@"
