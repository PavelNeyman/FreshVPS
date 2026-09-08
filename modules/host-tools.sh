#!/usr/bin/env bash
# Host operator tools that must survive bootstrap tree deletion
# shellcheck disable=SC2154

module_host_tools_install() {
  mkdir -p /opt/freshvps/bin /opt/freshvps/lib /opt/freshvps/scripts
  install -m 644 "${FRESHVPS_ROOT}/lib/"*.sh /opt/freshvps/lib/ 2>/dev/null || true
  install -m 755 "${FRESHVPS_ROOT}/bin/freshvps-doctor" /opt/freshvps/bin/freshvps-doctor
  ln -sfn /opt/freshvps/bin/freshvps-doctor /usr/local/bin/freshvps-doctor
  if [[ -f "${FRESHVPS_ROOT}/scripts/smoke-host.sh" ]]; then
    install -m 755 "${FRESHVPS_ROOT}/scripts/smoke-host.sh" /opt/freshvps/scripts/smoke-host.sh
    ln -sfn /opt/freshvps/scripts/smoke-host.sh /usr/local/bin/freshvps-smoke
  fi
  if [[ -f "${FRESHVPS_ROOT}/bin/freshvps-vpn" ]]; then
    install -m 755 "${FRESHVPS_ROOT}/bin/freshvps-vpn" /opt/freshvps/bin/freshvps-vpn
    ln -sfn /opt/freshvps/bin/freshvps-vpn /usr/local/bin/freshvps-vpn
  fi
  info "Host tools: freshvps-doctor | freshvps-smoke | freshvps-vpn"
}

module_host_tools_uninstall() {
  rm -f /usr/local/bin/freshvps-doctor /usr/local/bin/freshvps-smoke \
    /usr/local/bin/freshvps-vpn /usr/local/bin/freshvps-tests
  info "Host CLI symlinks removed (/opt/freshvps tree kept unless --purge)"
}
