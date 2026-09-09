#!/usr/bin/env bash
# NETDUCTOR_API_MODE=parallel|cutover
# shellcheck disable=SC2154
module_netductor_api_install() {
  local ver arch bin mode src
  ver="${NETDUCTOR_VERSION:-0.7.0-dev}"
  mode="${NETDUCTOR_API_MODE:-parallel}"
  case "$(uname -m)" in x86_64|amd64) arch=amd64 ;; aarch64|arm64) arch=arm64 ;; *) die "arch" ;; esac
  bin=/usr/local/bin/netductor
  if [[ -f "${FRESHVPS_ROOT}/dist/netductor-linux-${arch}" ]]; then
    install -m 755 "${FRESHVPS_ROOT}/dist/netductor-linux-${arch}" "${bin}"
  else
    curl -fsSL "https://github.com/PavelNeyman/netductor/releases/download/v${ver}/netductor-linux-${arch}" -o "${bin}"
    chmod 755 "${bin}"
  fi
  if [[ "${mode}" == "cutover" ]]; then
    src="${FRESHVPS_ROOT}/deploy/netductor-api-cutover.service"
    systemctl disable --now freshvps-api 2>/dev/null || true
  else
    src="${FRESHVPS_ROOT}/deploy/netductor-api.service"
  fi
  install -m 644 "${src}" /etc/systemd/system/netductor-api.service
  systemctl daemon-reload
  systemctl enable --now netductor-api
  info "netductor-api mode=${mode}"
  "${bin}" version || true
}
module_netductor_api_uninstall() {
  systemctl disable --now netductor-api 2>/dev/null || true
  rm -f /etc/systemd/system/netductor-api.service
  systemctl daemon-reload 2>/dev/null || true
}
