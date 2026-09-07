#!/usr/bin/env bash
# Host-side helper: reminds that OpenWrt script runs ON the router
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cat <<EOF
FreshVPS OpenWrt role

This does not configure the VPS. Copy the openwrt/ tree to the router and run there:

  scp -r ${ROOT}/openwrt root@ROUTER_IP:/root/freshvps-openwrt
  ssh root@ROUTER_IP
  cd /root/freshvps-openwrt
  cp site.conf.example site.conf   # edit LAN/Wi-Fi/VPN/OpenSOHO
  sh install-openwrt.sh

Or from repo root:
  bash install.sh --role openwrt   # prints this help
EOF
