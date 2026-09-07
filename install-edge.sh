#!/usr/bin/env bash
# FreshVPS edge-client — home SBC (Armbian/Debian) VPN gateway only
# Does NOT install VPS panels, server inbounds, OpenSOHO, etc.
set -euo pipefail

FRESHVPS_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export FRESHVPS_ROOT
# shellcheck source=lib/common.sh
source "${FRESHVPS_ROOT}/lib/common.sh"
# shellcheck source=modules/edge-client.sh
source "${FRESHVPS_ROOT}/modules/edge-client.sh"
# shellcheck source=modules/mikrotik-edge.sh
source "${FRESHVPS_ROOT}/modules/mikrotik-edge.sh"

VERSION="$(cat "${FRESHVPS_ROOT}/VERSION" 2>/dev/null || echo 0.0.0)"
WITH_MIKROTIK=0
CONFIG_FILE=""

usage() {
  cat <<EOF
FreshVPS ${VERSION} edge-client (home VPN gateway)

Usage:
  sudo bash install-edge.sh [--config FILE] [--with-mikrotik]

Environment / config keys:
  EDGE_UPSTREAM_VLESS   vless://... to VPS (required)
  EDGE_LAN_IF           LAN iface (default: auto)
  EDGE_MODE             gateway | tun-only (default: gateway)
  MIKROTIK_HOST         e.g. 192.168.88.1
  MIKROTIK_USER         default admin
  MIKROTIK_PASSWORD     RouterOS password (or SSH keys)
  EDGE_CLIENT_IP        IP of this host on LAN

Example:
  export EDGE_UPSTREAM_VLESS='vless://uuid@host:443?...'
  sudo bash install-edge.sh
  sudo bash install-edge.sh --with-mikrotik
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --config) CONFIG_FILE="$2"; shift 2 ;;
    --with-mikrotik) WITH_MIKROTIK=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) die "Unknown option: $1" ;;
  esac
done

if [[ -n "${CONFIG_FILE}" ]]; then
  [[ -f "${CONFIG_FILE}" ]] || die "Config not found: ${CONFIG_FILE}"
  set -a
  # shellcheck source=/dev/null
  source "${CONFIG_FILE}"
  set +a
fi

require_root
ensure_dirs
touch "${FRESHVPS_LOG}"
info "FreshVPS ${VERSION} edge-client starting"

module_edge_client_install

if [[ "${WITH_MIKROTIK}" -eq 1 ]]; then
  module_mikrotik_edge_install
fi

info "edge-client finished"
