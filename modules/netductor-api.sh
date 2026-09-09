#!/usr/bin/env bash
# Install netductor binary + systemd unit (parallel to freshvps-api)
# shellcheck disable=SC2154

module_netductor_api_install() {
  local ver arch url bin
  ver="${NETDUCTOR_VERSION:-0.7.0-dev}"
  case "$(uname -m)" in
    x86_64|amd64) arch=amd64 ;;
    aarch64|arm64) arch=arm64 ;;
    *) die "unsupported arch" ;;
  esac
  bin=/usr/local/bin/netductor
  url="https://github.com/PavelNeyman/netductor/releases/download/v${ver}/netductor-linux-${arch}"
  if [[ -f "${FRESHVPS_ROOT}/dist/netductor-linux-${arch}" ]]; then
    install -m 755 "${FRESHVPS_ROOT}/dist/netductor-linux-${arch}" "${bin}"
  else
    curl -fsSL "${url}" -o "${bin}"
    chmod 755 "${bin}"
  fi
  mkdir -p /opt/netductor
  if [[ -f "${FRESHVPS_ROOT}/deploy/netductor-api.service" ]]; then
    install -m 644 "${FRESHVPS_ROOT}/deploy/netductor-api.service" /etc/systemd/system/netductor-api.service
  fi
  systemctl daemon-reload
  systemctl enable --now netductor-api
  info "netductor-api on :8790 (proxies to freshvps-api :8787)"
  "${bin}" version || true
}

module_netductor_api_uninstall() {
  systemctl disable --now netductor-api 2>/dev/null || true
  rm -f /etc/systemd/system/netductor-api.service
  systemctl daemon-reload 2>/dev/null || true
  info "netductor-api stopped (binary kept)"
}
