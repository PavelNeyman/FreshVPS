#!/bin/sh
# FreshVPS OpenWrt site installer — run ON the router as root
# Usage:
#   sh install-openwrt.sh
#   sh install-openwrt.sh /path/to/site.conf
set -e

ROOT="$(CDPATH= cd -- "$(dirname "$0")" && pwd)"
CONF="${1:-$ROOT/site.conf}"

if [ ! -f "$CONF" ]; then
    echo "[ERROR] Config not found: $CONF"
    echo "Copy site.conf.example -> site.conf and edit."
    exit 1
fi

# shellcheck disable=SC1090
. "$CONF"
. "$ROOT/lib/uci-helpers.sh"
. "$ROOT/lib/site-network.sh"
. "$ROOT/lib/vpn-client.sh"
. "$ROOT/lib/opensoho-agent.sh"

if [ "$WIFI_PASSWORD" = "CHANGE_THIS_PASSWORD" ] || [ -z "$WIFI_PASSWORD" ]; then
    echo "[ERROR] Set WIFI_PASSWORD in $CONF"
    exit 1
fi

echo "============================================================"
echo "FreshVPS OpenWrt — $(cat /tmp/sysinfo/model 2>/dev/null || uname -a)"
echo "Config: $CONF"
echo "LAN: ${LAN_NETWORK}.${LAN_ROUTER_HOST}"
echo "VPN: ${ENABLE_VPN:-0}  mode=${ROUTE_MODE:-ru-direct}"
echo "OpenSOHO: ${ENABLE_OPENSOHO:-0}"
echo "============================================================"

NETWORK_CHANGED=0
DHCP_CHANGED=0
WIRELESS_CHANGED=0

apply_site_network
commit_and_reload_site

apply_vpn_client
apply_opensoho_agent

echo
echo "Done."
echo "Verify: ip route; logread -e sing-box; iwinfo"
