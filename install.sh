#!/usr/bin/env bash
# FreshVPS installer — produce a ready-to-use system, not bare defaults
set -euo pipefail

FRESHVPS_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export FRESHVPS_ROOT
# shellcheck source=lib/common.sh
source "${FRESHVPS_ROOT}/lib/common.sh"

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

usage() {
  cat <<EOF
FreshVPS ${VERSION} — ready-to-use Debian VPS bootstrap

Usage: sudo bash install.sh [options]
  --config FILE
  --non-interactive
  -h, --help
EOF
}

load_config() {
  local f="$1"
  [[ -f "${f}" ]] || die "Config not found: ${f}"
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
  [[ -n "${PUBLIC_IP}" ]] || die "PUBLIC_IP required for ready-to-use client links"

  if [[ "${NONINTERACTIVE}" -eq 0 ]]; then
    prompt_yes_no "Hardening?" y && ENABLE_HARDENING=1 || ENABLE_HARDENING=0
    prompt_yes_no "sing-box VPN?" y && ENABLE_SINGBOX=1 || ENABLE_SINGBOX=0
    prompt_yes_no "Blocky DNS?" y && ENABLE_BLOCKY=1 || ENABLE_BLOCKY=0
    prompt_yes_no "Multi-user VPN CLI + operator profile?" y && ENABLE_VPN_USERS=1 || ENABLE_VPN_USERS=0
    prompt_yes_no "Admin API for Shortcuts (localhost)?" y && ENABLE_VPN_API=1 || ENABLE_VPN_API=0
    prompt_yes_no "OpenSOHO?" y && ENABLE_OPENSOHO=1 || ENABLE_OPENSOHO=0
    prompt_yes_no "Uptime Kuma?" y && ENABLE_KUMA=1 || ENABLE_KUMA=0
    prompt_yes_no "Beszel?" y && ENABLE_BESZEL=1 || ENABLE_BESZEL=0
    prompt_yes_no "Telegram operator bot?" y && ENABLE_TELEGRAM=1 || ENABLE_TELEGRAM=0
    prompt_yes_no "restic backups?" y && ENABLE_BACKUP=1 || ENABLE_BACKUP=0
  fi

  if [[ "${ENABLE_TELEGRAM}" -eq 1 ]]; then
    TELEGRAM_BOT_TOKEN="$(prompt_value "Telegram bot token" "${TELEGRAM_BOT_TOKEN}")"
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

write_ready_summary() {
  local f="${FRESHVPS_ETC}/READY.txt"
  {
    echo "FreshVPS ${VERSION} READY — $(date -Iseconds)"
    echo "Host: $(hostname)  IP: ${PUBLIC_IP}"
    echo
    echo "=== VPN (your operator profile) ==="
    if [[ -f /etc/freshvps/clients/operator/link.txt ]]; then
      cat /etc/freshvps/clients/operator/link.txt
      echo "QR: /etc/freshvps/clients/operator/qr.png"
    else
      echo "(run: freshvps-vpn add operator)"
    fi
    echo
    echo "HY2 password: $(read_secret singbox_hy2_password 2>/dev/null || echo n/a) port ${SINGBOX_HY2_PORT}"
    echo "Reality public key: $(read_secret singbox_reality_public 2>/dev/null || echo n/a)"
    echo
    echo "=== Manage users ==="
    echo "CLI: freshvps-vpn add|list|link|disable|enable|revoke|session"
    echo "TG:  /vpn_add /vpn_list /vpn_link /session /status /ready"
    echo "API: 127.0.0.1:${VPN_API_PORT} (session via freshvps-vpn session)"
    echo "Shortcuts: build on device — docs/SHORTCUT-IOS.md"
    echo
    echo "=== Services (panels: localhost only) ==="
    echo "Blocky DNS: 127.0.0.1:53 (VPN clients use server DNS)"
    echo "Access panels: ssh -L 8090:127.0.0.1:8090 -L 3001:127.0.0.1:3001 -L 8091:127.0.0.1:8091 root@${PUBLIC_IP}"
    echo "OpenSOHO: http://127.0.0.1:${OPENSOHO_HTTP_PORT}  admin: $(read_secret opensoho_admin_email 2>/dev/null || echo n/a) / secrets/opensoho_admin_password"
    echo "Kuma:     http://127.0.0.1:${KUMA_PORT}  (create admin once in browser)"
    echo "Beszel:   http://127.0.0.1:${BESZEL_PORT}  admin: $(read_secret beszel_admin_email 2>/dev/null || echo n/a) / secrets/beszel_admin_password"
    [[ "${ENABLE_LAMPAC:-0}" -eq 1 ]] && echo "Lampac:   http://127.0.0.1:${LAMPAC_PORT:-9118}  rootpwd: secrets/lampac_root_password"
    echo
    echo "Secrets dir: ${FRESHVPS_ETC}/secrets"
    echo "Log: ${FRESHVPS_LOG}"
  } >"${f}"
  chmod 600 "${f}"
  info "Wrote ${f}"
  if [[ -x /opt/freshvps-telegram/notify.sh ]]; then
    /opt/freshvps-telegram/notify.sh "FreshVPS READY on ${PUBLIC_IP}. Operator link in /etc/freshvps/READY.txt"
  fi
  cat "${f}"
}

main() {
  parse_args "$@"
  require_root
  require_debian
  ensure_dirs
  touch "${FRESHVPS_LOG}"

  info "FreshVPS ${VERSION} starting (ready-to-use mode)"
  collect_settings

  export PUBLIC_IP SSH_PORT
  export SINGBOX_VLESS_PORT SINGBOX_HY2_PORT SINGBOX_REALITY_SNI
  export BLOCKY_DNS_PORT OPENSOHO_HTTP_PORT KUMA_PORT BESZEL_PORT
  export VPN_API_PORT VPN_API_BIND
  export TELEGRAM_BOT_TOKEN TELEGRAM_ADMIN_ID BACKUP_REPO

  printf '%s\n' "${PUBLIC_IP}" >"${FRESHVPS_ETC}/public_ip"

  pkg_install ca-certificates curl wget openssl jq tar gzip coreutils

  [[ "${ENABLE_HARDENING}" -eq 1 ]] && run_module hardening
  [[ "${ENABLE_SINGBOX}" -eq 1 ]] && run_module sing-box
  [[ "${ENABLE_BLOCKY}" -eq 1 ]] && run_module blocky
  [[ "${ENABLE_VPN_USERS}" -eq 1 && "${ENABLE_SINGBOX}" -eq 1 ]] && run_module vpn-users
  [[ "${ENABLE_VPN_API}" -eq 1 ]] && run_module vpn-api
  [[ "${ENABLE_OPENSOHO}" -eq 1 ]] && run_module opensoho
  [[ "${ENABLE_KUMA}" -eq 1 ]] && run_module kuma
  [[ "${ENABLE_BESZEL}" -eq 1 ]] && run_module beszel
  [[ "${ENABLE_LAMPAC}" -eq 1 ]] && run_module lampac
  [[ "${ENABLE_TELEGRAM}" -eq 1 ]] && run_module telegram
  [[ "${ENABLE_BACKUP}" -eq 1 ]] && run_module backup

  write_ready_summary
  info "FreshVPS install finished — system is ready for use"
}

main "$@"
