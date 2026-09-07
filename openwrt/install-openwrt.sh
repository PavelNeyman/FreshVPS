#!/bin/sh
# FreshVPS OpenWrt site installer — run ON the router as root
# Idempotent LAN/Wi-Fi (Cudy-tested logic) + optional VPN/OpenSOHO
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
. "$ROOT/lib/detect-env.sh"
. "$ROOT/lib/site-network.sh"
. "$ROOT/lib/vless-parse.sh"
. "$ROOT/lib/vpn-client.sh"
. "$ROOT/lib/opensoho-agent.sh"

if [ "$WIFI_PASSWORD" = "CHANGE_THIS_PASSWORD" ] || [ -z "$WIFI_PASSWORD" ]; then
    echo "[ERROR] Set WIFI_PASSWORD in $CONF"
    exit 1
fi
if [ -z "$LAN_NETWORK" ] || [ -z "$LAN_ROUTER_HOST" ] || [ -z "$LAN_NETMASK" ]; then
    echo "[ERROR] LAN configuration incomplete"
    exit 1
fi
if [ -z "$DHCP_POOL_START" ] || [ -z "$DHCP_POOL_SIZE" ] || [ -z "$DHCP_LEASETIME" ]; then
    echo "[ERROR] DHCP configuration incomplete"
    exit 1
fi

detect_openwrt_env

LAN_IP="${LAN_NETWORK}.${LAN_ROUTER_HOST}"
NETWORK_CHANGED=0
DHCP_CHANGED=0
WIRELESS_CHANGED=0

echo
echo "============================================================"
echo "TARGET CONFIGURATION"
echo "============================================================"
echo "LAN: ${LAN_IP}/${LAN_NETMASK}"
echo "WAN: device=${WAN_DEVICE} proto=${WAN_PROTO:-dhcp}"
echo "DHCP: ${LAN_NETWORK}.${DHCP_POOL_START} size=${DHCP_POOL_SIZE} lease=${DHCP_LEASETIME}"
echo "Wi-Fi 2.4: ${WIFI_2G_SSID} ch=${WIFI_2G_CHANNEL} ${WIFI_2G_HTMODE}"
echo "Wi-Fi 5:   ${WIFI_5G_SSID} ch=${WIFI_5G_CHANNEL} ${WIFI_5G_HTMODE}"
echo "VPN: ${ENABLE_VPN:-0} mode=${ROUTE_MODE:-ru-direct}"
echo "OpenSOHO: ${ENABLE_OPENSOHO:-0}"
echo "============================================================"

apply_site_network
commit_and_reload_site

apply_vpn_client
apply_opensoho_agent

echo
echo "============================================================"
echo "VERIFICATION"
echo "============================================================"
echo "  iw reg get"
echo "  iwinfo"
echo "  ip addr; ip route"
echo "  logread -e sing-box"
echo "Done."
