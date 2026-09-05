#!/usr/bin/env bash
# FreshVPS shared helpers. Sourced by install.sh and modules.
# shellcheck disable=SC2034

set -euo pipefail

FRESHVPS_ROOT="${FRESHVPS_ROOT:-}"
FRESHVPS_STATE_DIR="${FRESHVPS_STATE_DIR:-/var/lib/freshvps}"
FRESHVPS_ETC="${FRESHVPS_ETC:-/etc/freshvps}"
FRESHVPS_LOG="${FRESHVPS_LOG:-/var/log/freshvps.log}"

log()  { printf '[%s] %s\n' "$(date -Iseconds)" "$*" | tee -a "${FRESHVPS_LOG}" 2>/dev/null || printf '[%s] %s\n' "$(date -Iseconds)" "$*"; }
info() { log "INFO  $*"; }
warn() { log "WARN  $*"; }
err()  { log "ERROR $*"; }
die()  { err "$*"; exit 1; }

require_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    die "Run as root (sudo bash install.sh)"
  fi
}

require_debian() {
  if [[ ! -f /etc/debian_version ]]; then
    die "Only Debian is supported"
  fi
}

ensure_dirs() {
  mkdir -p "${FRESHVPS_STATE_DIR}" "${FRESHVPS_ETC}" "${FRESHVPS_ETC}/secrets"
  chmod 700 "${FRESHVPS_ETC}/secrets"
}

apt_update_once() {
  if [[ ! -f "${FRESHVPS_STATE_DIR}/apt-updated" ]] || [[ "$(find "${FRESHVPS_STATE_DIR}/apt-updated" -mmin +60 2>/dev/null || true)" ]]; then
    export DEBIAN_FRONTEND=noninteractive
    apt-get update -y
    touch "${FRESHVPS_STATE_DIR}/apt-updated"
  fi
}

pkg_install() {
  apt_update_once
  export DEBIAN_FRONTEND=noninteractive
  apt-get install -y --no-install-recommends "$@"
}

arch_go() {
  local m
  m="$(uname -m)"
  case "${m}" in
    x86_64|amd64) echo amd64 ;;
    aarch64|arm64) echo arm64 ;;
    armv7l|armhf) echo armv7 ;;
    *) die "Unsupported architecture: ${m}" ;;
  esac
}

detect_public_ip() {
  local ip=""
  ip="$(curl -4 -fsS --max-time 5 https://ifconfig.me 2>/dev/null || true)"
  if [[ -z "${ip}" ]]; then
    ip="$(curl -4 -fsS --max-time 5 https://api.ipify.org 2>/dev/null || true)"
  fi
  echo "${ip}"
}

random_hex() {
  local n="${1:-16}"
  openssl rand -hex "${n}"
}

random_uuid() {
  if command -v uuidgen >/dev/null 2>&1; then
    uuidgen | tr '[:upper:]' '[:lower:]'
  else
    cat /proc/sys/kernel/random/uuid
  fi
}

mark_done() {
  local name="$1"
  mkdir -p "${FRESHVPS_STATE_DIR}/modules"
  date -Iseconds > "${FRESHVPS_STATE_DIR}/modules/${name}.done"
}

is_done() {
  local name="$1"
  [[ -f "${FRESHVPS_STATE_DIR}/modules/${name}.done" ]]
}

write_secret() {
  local file="$1"
  local value="$2"
  umask 077
  printf '%s\n' "${value}" > "${FRESHVPS_ETC}/secrets/${file}"
  chmod 600 "${FRESHVPS_ETC}/secrets/${file}"
}

read_secret() {
  local file="$1"
  if [[ -f "${FRESHVPS_ETC}/secrets/${file}" ]]; then
    cat "${FRESHVPS_ETC}/secrets/${file}"
  fi
}

systemd_enable_start() {
  local unit="$1"
  systemctl daemon-reload
  systemctl enable --now "${unit}"
}

firewall_allow_tcp() {
  local port="$1"
  local comment="${2:-freshvps}"
  if command -v ufw >/dev/null 2>&1 && ufw status 2>/dev/null | grep -qi active; then
    ufw allow "${port}/tcp" comment "${comment}" || true
  elif command -v nft >/dev/null 2>&1; then
    # Best-effort; full policy is in hardening module
    true
  fi
}

firewall_allow_udp() {
  local port="$1"
  local comment="${2:-freshvps}"
  if command -v ufw >/dev/null 2>&1 && ufw status 2>/dev/null | grep -qi active; then
    ufw allow "${port}/udp" comment "${comment}" || true
  fi
}
