#!/usr/bin/env bash
# FreshVPS installer entrypoint
set -euo pipefail

FRESHVPS_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export FRESHVPS_ROOT
# shellcheck source=lib/common.sh
source "${FRESHVPS_ROOT}/lib/common.sh"

VERSION="$(cat "${FRESHVPS_ROOT}/VERSION" 2>/dev/null || echo 0.0.0)"

# Defaults (overridden by config / TUI)
ENABLE_HARDENING=1
ENABLE_SINGBOX=1
ENABLE_BLOCKY=1
ENABLE_OPENSOHO=1
ENABLE_KUMA=1
ENABLE_BESZEL=1
ENABLE_TELEGRAM=1
ENABLE_BACKUP=1

PUBLIC_IP=""
SSH_PORT=22
SINGBOX_VLESS_PORT=443
SINGBOX_HY2_PORT=8443
BLOCKY_DNS_PORT=53
OPENSOHO_HTTP_PORT=8090
KUMA_PORT=3001
BESZEL_PORT=8091
TELEGRAM_BOT_TOKEN=""
TELEGRAM_ADMIN_ID=""
BACKUP_REPO=""
CONFIG_FILE=""
NONINTERACTIVE=0

usage() {
  cat <<EOF
FreshVPS ${VERSION} — modular Debian VPS bootstrap

Usage: sudo bash install.sh [options]

  --config FILE     Load settings from FILE (key=value)
  --non-interactive Do not prompt (requires --config or defaults)
  -h, --help        Show this help
EOF
}

load_config() {
  local f="$1"
  [[ -f "${f}" ]] || die "Config not found: ${f}"
  # shellcheck disable=SC1090
  set -a
  # shellcheck source=/dev/null
  source "${f}"
  set +a
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --config) CONFIG_FILE="$2"; shift 2 ;;
      --non-interactive) NONINTERACTIVE=1; shift ;;
      -h|--help) usage; exit 0 ;;
      *) die "Unknown option: $1" ;;
    esac
  done
}

prompt_yes_no() {
  local q="$1" default="${2:-y}"
  if [[ "${NONINTERACTIVE}" -eq 1 ]]; then
    [[ "${default}" == "y" ]] && return 0 || return 1
  fi
  if command -v whiptail >/dev/null 2>&1; then
    if [[ "${default}" == "y" ]]; then
      whiptail --yesno "${q}" 10 60 && return 0 || return 1
    else
      whiptail --defaultno --yesno "${q}" 10 60 && return 0 || return 1
    fi
  fi
  local a
  read -r -p "${q} [Y/n] " a || true
  a="${a:-${default}}"
  [[ "${a}" =~ ^[Yy] ]]
}

prompt_value() {
  local q="$1" default="${2:-}"
  if [[ "${NONINTERACTIVE}" -eq 1 ]]; then
    echo "${default}"
    return
  fi
  if command -v whiptail >/dev/null 2>&1; then
    whiptail --inputbox "${q}" 10 60 "${default}" 3>&1 1>&2 2>&3 || echo "${default}"
    return
  fi
  local a
  read -r -p "${q} [${default}]: " a || true
  echo "${a:-${default}}"
}

collect_settings() {
  if [[ -n "${CONFIG_FILE}" ]]; then
    load_config "${CONFIG_FILE}"
  fi

  PUBLIC_IP="${PUBLIC_IP:-$(detect_public_ip)}"
  PUBLIC_IP="$(prompt_value "Public IPv4 of this VPS" "${PUBLIC_IP}")"

  if [[ "${NONINTERACTIVE}" -eq 0 ]]; then
    prompt_yes_no "Install hardening (SSH, fail2ban, firewall, BBR)?" y && ENABLE_HARDENING=1 || ENABLE_HARDENING=0
    prompt_yes_no "Install sing-box (VLESS+Reality + Hysteria2)?" y && ENABLE_SINGBOX=1 || ENABLE_SINGBOX=0
    prompt_yes_no "Install Blocky (DNS filtering)?" y && ENABLE_BLOCKY=1 || ENABLE_BLOCKY=0
    prompt_yes_no "Install OpenSOHO (OpenWrt management)?" y && ENABLE_OPENSOHO=1 || ENABLE_OPENSOHO=0
    prompt_yes_no "Install Uptime Kuma?" y && ENABLE_KUMA=1 || ENABLE_KUMA=0
    prompt_yes_no "Install Beszel?" y && ENABLE_BESZEL=1 || ENABLE_BESZEL=0
    prompt_yes_no "Install Telegram bot helpers?" y && ENABLE_TELEGRAM=1 || ENABLE_TELEGRAM=0
    prompt_yes_no "Install restic backup scaffolding?" y && ENABLE_BACKUP=1 || ENABLE_BACKUP=0
  fi

  if [[ "${ENABLE_TELEGRAM}" -eq 1 ]]; then
    TELEGRAM_BOT_TOKEN="$(prompt_value "Telegram bot token (empty to skip runtime config)" "${TELEGRAM_BOT_TOKEN}")"
    TELEGRAM_ADMIN_ID="$(prompt_value "Telegram admin chat id" "${TELEGRAM_ADMIN_ID}")"
  fi
}

run_module() {
  local name="$1"
  local script="${FRESHVPS_ROOT}/modules/${name}.sh"
  [[ -f "${script}" ]] || die "Module missing: ${script}"
  info "=== Module: ${name} ==="
  # shellcheck source=/dev/null
  source "${script}"
  "module_${name//-/_}_install"
  mark_done "${name}"
  info "=== Done: ${name} ==="
}

main() {
  parse_args "$@"
  require_root
  require_debian
  ensure_dirs
  touch "${FRESHVPS_LOG}"

  info "FreshVPS ${VERSION} starting"
  collect_settings

  # Export for modules
  export PUBLIC_IP SSH_PORT
  export SINGBOX_VLESS_PORT SINGBOX_HY2_PORT
  export BLOCKY_DNS_PORT OPENSOHO_HTTP_PORT KUMA_PORT BESZEL_PORT
  export TELEGRAM_BOT_TOKEN TELEGRAM_ADMIN_ID BACKUP_REPO
  export ENABLE_HARDENING ENABLE_SINGBOX ENABLE_BLOCKY ENABLE_OPENSOHO
  export ENABLE_KUMA ENABLE_BESZEL ENABLE_TELEGRAM ENABLE_BACKUP

  pkg_install ca-certificates curl wget openssl jq tar gzip coreutils

  [[ "${ENABLE_HARDENING}" -eq 1 ]] && run_module hardening
  [[ "${ENABLE_SINGBOX}" -eq 1 ]] && run_module sing-box
  [[ "${ENABLE_BLOCKY}" -eq 1 ]] && run_module blocky
  [[ "${ENABLE_OPENSOHO}" -eq 1 ]] && run_module opensoho
  [[ "${ENABLE_KUMA}" -eq 1 ]] && run_module kuma
  [[ "${ENABLE_BESZEL}" -eq 1 ]] && run_module beszel
  [[ "${ENABLE_TELEGRAM}" -eq 1 ]] && run_module telegram
  [[ "${ENABLE_BACKUP}" -eq 1 ]] && run_module backup

  info "FreshVPS install finished. State: ${FRESHVPS_STATE_DIR}"
  info "Secrets (if any): ${FRESHVPS_ETC}/secrets"
  info "Log: ${FRESHVPS_LOG}"
}

main "$@"
