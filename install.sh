#!/usr/bin/env bash
# FreshVPS installer — ready-to-use; safe re-run / upgrade / VPS tests
set -euo pipefail

FRESHVPS_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export FRESHVPS_ROOT
# shellcheck source=lib/common.sh
source "${FRESHVPS_ROOT}/lib/common.sh"
# shellcheck source=lib/tui.sh
source "${FRESHVPS_ROOT}/lib/tui.sh"
# shellcheck source=lib/idempotent.sh
source "${FRESHVPS_ROOT}/lib/idempotent.sh"

VERSION="$(cat "${FRESHVPS_ROOT}/VERSION" 2>/dev/null || echo 0.0.0)"

ENABLE_HARDENING=1
ENABLE_SINGBOX=1
ENABLE_BLOCKY=1
ENABLE_VPN_USERS=1
ENABLE_VPN_API=1
ENABLE_OPENSOHO=1
ENABLE_KUMA=1
ENABLE_BESZEL=1
ENABLE_TELEGRAM=1
ENABLE_BACKUP=1
ENABLE_LAMPAC=0

PUBLIC_IP=""
SSH_PORT=22
SINGBOX_VLESS_PORT=443
SINGBOX_HY2_PORT=8443
BLOCKY_DNS_PORT=53
OPENSOHO_HTTP_PORT=8090
KUMA_PORT=3001
BESZEL_PORT=8091
VPN_API_PORT=8787
VPN_API_BIND=127.0.0.1
TELEGRAM_BOT_TOKEN=""
TELEGRAM_ADMIN_ID=""
BACKUP_REPO=""
CONFIG_FILE=""
NONINTERACTIVE=0
export NONINTERACTIVE
ROLE=vps
WITH_MIKROTIK=0
DO_UPGRADE=0
ACTION=""

usage() {
  cat <<EOF
FreshVPS ${VERSION}

  --config FILE | --non-interactive | --upgrade | --force | --force-module NAME
  --role vps|edge-client|openwrt | --with-mikrotik | --tests

CLI tests (after install):  freshvps-tests
EOF
}

load_config() {
  local f="$1"
  [[ -f "${f}" ]] || die "Config not found: ${f}"
  set -a; source "${f}"; set +a
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --config) CONFIG_FILE="$2"; shift 2 ;;
      --non-interactive) NONINTERACTIVE=1; export NONINTERACTIVE; shift ;;
      --upgrade) DO_UPGRADE=1; FRESHVPS_UPGRADE=1; export FRESHVPS_UPGRADE; ACTION=upgrade; shift ;;
      --force) FRESHVPS_FORCE=1; export FRESHVPS_FORCE; shift ;;
      --force-module)
        FRESHVPS_FORCE_MODULES="${FRESHVPS_FORCE_MODULES:+${FRESHVPS_FORCE_MODULES},}$2"
        export FRESHVPS_FORCE_MODULES; shift 2 ;;
      --role) ROLE="$2"; shift 2 ;;
      --with-mikrotik) WITH_MIKROTIK=1; shift ;;
      --tests) ACTION=tests; shift ;;
      -h|--help) usage; exit 0 ;;
      *) die "Unknown option: $1" ;;
    esac
  done
}

prompt_yes_no() { tui_yesno "$@"; }
prompt_value() { tui_input "$@"; }

pick_action_menu() {
  [[ -n "${ACTION}" ]] && return 0
  [[ "${NONINTERACTIVE}" -eq 1 ]] && { ACTION=install; return 0; }
  if command -v whiptail >/dev/null 2>&1; then
    ACTION="$(whiptail --title "FreshVPS ${VERSION}" --menu "Choose action" 16 60 5 \
      install "Install / configure modules" \
      upgrade "Upgrade / re-run (idempotent)" \
      tests "VPS tests (network / quality)" \
      quit "Exit" 3>&1 1>&2 2>&3)" || ACTION=quit
  else
    echo "1) install  2) upgrade  3) VPS tests  4) quit"
    read -r -p "> " n
    case "${n}" in
      2) ACTION=upgrade ;;
      3) ACTION=tests ;;
      4) ACTION=quit ;;
      *) ACTION=install ;;
    esac
  fi
}

collect_settings() {
  if [[ -z "${CONFIG_FILE}" && "${DO_UPGRADE}" -eq 1 && -f /root/freshvps.conf ]]; then
    CONFIG_FILE=/root/freshvps.conf
  fi
  [[ -n "${CONFIG_FILE}" ]] && load_config "${CONFIG_FILE}"

  if [[ -z "${PUBLIC_IP}" && -f "${FRESHVPS_ETC}/public_ip" ]]; then
    PUBLIC_IP="$(cat "${FRESHVPS_ETC}/public_ip")"
  fi
  PUBLIC_IP="${PUBLIC_IP:-$(detect_public_ip)}"
  if [[ ! ( "${DO_UPGRADE}" -eq 1 && -n "${PUBLIC_IP}" && "${NONINTERACTIVE}" -eq 1 ) ]]; then
    PUBLIC_IP="$(prompt_value "Public IPv4 of this VPS" "${PUBLIC_IP}")"
  fi
  [[ -n "${PUBLIC_IP}" ]] || die "PUBLIC_IP required"

  if [[ "${NONINTERACTIVE}" -eq 0 && "${DO_UPGRADE}" -eq 0 ]]; then
    prompt_yes_no "Hardening?" y && ENABLE_HARDENING=1 || ENABLE_HARDENING=0
    prompt_yes_no "sing-box VPN?" y && ENABLE_SINGBOX=1 || ENABLE_SINGBOX=0
    prompt_yes_no "Blocky DNS?" y && ENABLE_BLOCKY=1 || ENABLE_BLOCKY=0
    prompt_yes_no "Multi-user VPN CLI?" y && ENABLE_VPN_USERS=1 || ENABLE_VPN_USERS=0
    prompt_yes_no "Admin API?" y && ENABLE_VPN_API=1 || ENABLE_VPN_API=0
    prompt_yes_no "OpenSOHO?" y && ENABLE_OPENSOHO=1 || ENABLE_OPENSOHO=0
    prompt_yes_no "Uptime Kuma?" y && ENABLE_KUMA=1 || ENABLE_KUMA=0
    prompt_yes_no "Beszel?" y && ENABLE_BESZEL=1 || ENABLE_BESZEL=0
    prompt_yes_no "Telegram bot?" y && ENABLE_TELEGRAM=1 || ENABLE_TELEGRAM=0
    prompt_yes_no "restic backups?" y && ENABLE_BACKUP=1 || ENABLE_BACKUP=0
    prompt_yes_no "Lampac?" n && ENABLE_LAMPAC=1 || ENABLE_LAMPAC=0
  fi

  if [[ "${ENABLE_TELEGRAM}" -eq 1 && "${NONINTERACTIVE}" -eq 0 && "${DO_UPGRADE}" -eq 0 ]]; then
    TELEGRAM_BOT_TOKEN="$(prompt_value "Telegram bot token" "${TELEGRAM_BOT_TOKEN}")"
    TELEGRAM_ADMIN_ID="$(prompt_value "Telegram admin chat id" "${TELEGRAM_ADMIN_ID}")"
  fi
}

install_panels_compose() {
  mkdir -p /opt/freshvps/compose
  [[ -f "${FRESHVPS_ROOT}/compose/panels.yml" ]] && \
    install -m 644 "${FRESHVPS_ROOT}/compose/panels.yml" /opt/freshvps/compose/panels.yml
}

write_ready_summary() {
  local f="${FRESHVPS_ETC}/READY.txt"
  {
    echo "FreshVPS ${VERSION} READY — $(date -Iseconds)"
    echo "Host: $(hostname)  IP: ${PUBLIC_IP}"
    echo
    if [[ -f /etc/freshvps/clients/operator/subscription.txt ]]; then
      cat /etc/freshvps/clients/operator/subscription.txt
    fi
    echo
    echo "CLI: freshvps-vpn | freshvps-doctor | freshvps-tests"
    echo "Tests menu: sudo freshvps-tests   or   sudo bash install.sh --tests"
  } >"${f}"
  chmod 600 "${f}"
  cat "${f}"
}

run_tests_action() {
  # ensure runner present
  # shellcheck source=/dev/null
  source "${FRESHVPS_ROOT}/modules/vps-tests.sh"
  module_vps_tests_install
  bash /opt/freshvps/scripts/vps-tests.sh
}

main() {
  parse_args "$@"
  if [[ "${ROLE}" == "openwrt" ]]; then exec bash "${FRESHVPS_ROOT}/install-openwrt.sh"; fi
  if [[ "${ROLE}" == "edge-client" ]]; then
    extra=(); [[ -n "${CONFIG_FILE}" ]] && extra+=(--config "${CONFIG_FILE}")
    [[ "${WITH_MIKROTIK}" -eq 1 ]] && extra+=(--with-mikrotik)
    exec bash "${FRESHVPS_ROOT}/install-edge.sh" "${extra[@]}"
  fi

  pick_action_menu
  if [[ "${ACTION}" == "quit" ]]; then exit 0; fi
  if [[ "${ACTION}" == "tests" ]]; then
    require_root
    run_tests_action
    exit 0
  fi
  if [[ "${ACTION}" == "upgrade" ]]; then
    DO_UPGRADE=1; FRESHVPS_UPGRADE=1; export FRESHVPS_UPGRADE
  fi

  require_root
  require_debian
  ensure_dirs
  touch "${FRESHVPS_LOG}"

  if [[ "${DO_UPGRADE}" -eq 1 ]]; then
    NONINTERACTIVE="${NONINTERACTIVE:-1}"; export NONINTERACTIVE
  fi

  info "FreshVPS ${VERSION} action=${ACTION:-install}"
  collect_settings
  export PUBLIC_IP SSH_PORT SINGBOX_VLESS_PORT SINGBOX_HY2_PORT SINGBOX_REALITY_SNI
  export BLOCKY_DNS_PORT OPENSOHO_HTTP_PORT KUMA_PORT BESZEL_PORT
  export VPN_API_PORT VPN_API_BIND TELEGRAM_BOT_TOKEN TELEGRAM_ADMIN_ID BACKUP_REPO ENABLE_LAMPAC
  printf '%s\n' "${PUBLIC_IP}" >"${FRESHVPS_ETC}/public_ip"

  pkg_install ca-certificates curl wget openssl jq tar gzip coreutils
  install_panels_compose

  [[ "${ENABLE_HARDENING}" -eq 1 ]] && run_module_idempotent hardening
  [[ "${ENABLE_SINGBOX}" -eq 1 ]] && run_module_idempotent sing-box
  [[ "${ENABLE_BLOCKY}" -eq 1 ]] && run_module_idempotent blocky
  [[ "${ENABLE_VPN_USERS}" -eq 1 && "${ENABLE_SINGBOX}" -eq 1 ]] && run_module_idempotent vpn-users
  [[ "${ENABLE_VPN_API}" -eq 1 ]] && run_module_idempotent vpn-api
  [[ "${ENABLE_OPENSOHO}" -eq 1 ]] && run_module_idempotent opensoho
  [[ "${ENABLE_KUMA}" -eq 1 ]] && run_module_idempotent kuma
  [[ "${ENABLE_BESZEL}" -eq 1 ]] && run_module_idempotent beszel
  [[ "${ENABLE_LAMPAC}" -eq 1 ]] && run_module_idempotent lampac
  [[ "${ENABLE_TELEGRAM}" -eq 1 ]] && run_module_idempotent telegram
  [[ "${ENABLE_BACKUP}" -eq 1 ]] && run_module_idempotent backup
  # always install test runner
  run_module_idempotent vps-tests

  write_installed_version "${VERSION}"
  write_ready_summary
  info "Finished — installed_version=${VERSION}"

  if [[ "${NONINTERACTIVE}" -eq 0 ]]; then
    if prompt_yes_no "Run VPS tests now?" n; then
      run_tests_action
    fi
  fi
}

main "$@"
