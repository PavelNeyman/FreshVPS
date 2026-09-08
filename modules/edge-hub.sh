#!/usr/bin/env bash
# Module: edge hub (token + dirs for OpenWrt agents)
# shellcheck disable=SC2154

module_edge_hub_install() {
  mkdir -p /var/lib/freshvps/edge /etc/freshvps/secrets
  chmod 700 /var/lib/freshvps/edge /etc/freshvps/secrets
  if [[ ! -f /etc/freshvps/secrets/edge_token ]]; then
    local t
    t="$(openssl rand -hex 24)"
    printf '%s\n' "${t}" >/etc/freshvps/secrets/edge_token
    chmod 600 /etc/freshvps/secrets/edge_token
    info "Edge token created: /etc/freshvps/secrets/edge_token"
  else
    info "Edge token already present"
  fi
  install -m 644 "${FRESHVPS_ROOT}/runtime/api/edge_store.py" /opt/freshvps/runtime/api/edge_store.py
  # agent files for download from VPS
  mkdir -p /opt/freshvps/edge/openwrt
  if [[ -d "${FRESHVPS_ROOT}/edge/openwrt" ]]; then
    cp -a "${FRESHVPS_ROOT}/edge/openwrt/." /opt/freshvps/edge/openwrt/
  fi
  systemctl restart freshvps-api 2>/dev/null || true
  info "Edge hub ready. Token: $(cat /etc/freshvps/secrets/edge_token 2>/dev/null | head -c 8)…"
  info "Agent files: /opt/freshvps/edge/openwrt/"
}

module_edge_hub_uninstall() {
  info "edge-hub uninstall keeps token/data unless purge"
}
