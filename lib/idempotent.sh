#!/usr/bin/env bash
# Idempotent install helpers
# shellcheck disable=SC2034

: "${FRESHVPS_FORCE:=0}"
: "${FRESHVPS_UPGRADE:=0}"
: "${FRESHVPS_FORCE_MODULES:=}"

module_force_requested() {
  local name="$1"
  [[ "${FRESHVPS_FORCE}" -eq 1 ]] && return 0
  [[ "${FRESHVPS_FORCE_MODULES}" == "all" ]] && return 0
  [[ ",${FRESHVPS_FORCE_MODULES}," == *",${name},"* ]] && return 0
  return 1
}

write_installed_version() {
  local v="${1:-}"
  mkdir -p "${FRESHVPS_STATE_DIR}"
  printf '%s\n' "${v}" >"${FRESHVPS_STATE_DIR}/installed_version"
  printf '%s\n' "$(date -Iseconds)" >"${FRESHVPS_STATE_DIR}/last_install"
}

read_installed_version() {
  [[ -f "${FRESHVPS_STATE_DIR}/installed_version" ]] && cat "${FRESHVPS_STATE_DIR}/installed_version"
}

unit_active() { systemctl is-active --quiet "$1" 2>/dev/null; }
container_running() {
  command -v docker >/dev/null 2>&1 || return 1
  docker inspect -f '{{.State.Running}}' "$1" 2>/dev/null | grep -qi true
}

module_default_healthy() {
  local name="$1"
  is_done "${name}" || return 1
  case "${name}" in
    host-tools) [[ -x /usr/local/bin/freshvps-doctor ]] && [[ -x /opt/freshvps/scripts/smoke-host.sh ]] ;;
    hardening) unit_active fail2ban && command -v ufw >/dev/null ;;
    sing-box)
      unit_active sing-box && [[ -x /usr/local/bin/sing-box ]] \
        && /usr/local/bin/sing-box check -c /usr/local/etc/sing-box/config.json >/dev/null 2>&1
      ;;
    blocky) unit_active blocky && [[ -x /usr/local/bin/blocky ]] ;;
    vpn-users) [[ -x /usr/local/bin/freshvps-vpn ]] && [[ -f /etc/freshvps/vpn-users.json ]] ;;
    vpn-api) unit_active freshvps-api 2>/dev/null || return 1 ;;
    opensoho) container_running opensoho || unit_active opensoho ;;
    kuma) container_running uptime-kuma ;;
    beszel) container_running beszel ;;
    lampac) container_running lampac ;;
    telegram) unit_active freshvps-telegram-bot 2>/dev/null || return 1 ;;
    backup) [[ -x /opt/freshvps-backup/backup.sh ]] || [[ -x /opt/freshvps/backup/backup.sh ]] ;;
    vps-tests) [[ -x /usr/local/bin/freshvps-tests ]] && [[ -f /opt/freshvps/scripts/vps-tests.sh ]] ;;
    *) is_done "${name}" ;;
  esac
}

run_module_idempotent() {
  local name="$1"
  local script="${FRESHVPS_ROOT}/modules/${name}.sh"
  [[ -f "${script}" ]] || die "Module missing: ${script}"
  if ! module_force_requested "${name}" && module_default_healthy "${name}"; then
    info "=== Skip ${name} (already healthy) ==="
    return 0
  fi
  info "=== Module: ${name} ==="
  # shellcheck source=/dev/null
  source "${script}"
  "module_${name//-/_}_install"
  mark_done "${name}"
  info "=== Done: ${name} ==="
}
