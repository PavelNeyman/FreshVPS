#!/usr/bin/env bash
# Netductor Go API — parallel (:8790) or cutover (:8787)
# NETDUCTOR_API_MODE=parallel|cutover  NETDUCTOR_VERSION=0.7.0-dev
# shellcheck disable=SC2154

module_netductor_api_install() {
  local ver arch bin mode src root
  root="${FRESHVPS_ROOT:-/opt/freshvps}"
  ver="${NETDUCTOR_VERSION:-0.7.0-dev}"
  mode="${NETDUCTOR_API_MODE:-parallel}"
  case "$(uname -m)" in
    x86_64|amd64) arch=amd64 ;;
    aarch64|arm64) arch=arm64 ;;
    *) die "unsupported arch for netductor" ;;
  esac
  bin=/usr/local/bin/netductor
  if [[ -x "${root}/dist/netductor-linux-${arch}" ]]; then
    install -m 755 "${root}/dist/netductor-linux-${arch}" "${bin}"
  elif [[ -x "${root}/bin/netductor" ]]; then
    install -m 755 "${root}/bin/netductor" "${bin}"
  else
    info "Downloading netductor v${ver} (${arch})"
    curl -fsSL "https://github.com/PavelNeyman/netductor/releases/download/v${ver}/netductor-linux-${arch}" -o "${bin}"
    chmod 755 "${bin}"
  fi
  mkdir -p /opt/netductor
  if [[ "${mode}" == "cutover" ]]; then
    src="${root}/deploy/netductor-api-cutover.service"
    systemctl disable --now freshvps-api 2>/dev/null || true
  else
    src="${root}/deploy/netductor-api.service"
  fi
  if [[ ! -f "${src}" ]]; then
    die "missing unit ${src}"
  fi
  install -m 644 "${src}" /etc/systemd/system/netductor-api.service
  systemctl daemon-reload
  systemctl enable --now netductor-api
  info "netductor-api mode=${mode} ($("${bin}" version 2>/dev/null || echo ok))"
}

module_netductor_api_uninstall() {
  systemctl disable --now netductor-api 2>/dev/null || true
  rm -f /etc/systemd/system/netductor-api.service
  systemctl daemon-reload 2>/dev/null || true
  info "netductor-api removed (binary kept at /usr/local/bin/netductor)"
}

module_netductor_api_purge() {
  module_netductor_api_uninstall
  rm -f /usr/local/bin/netductor
}
