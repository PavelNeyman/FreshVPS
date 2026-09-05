#!/usr/bin/env bash
# Module: multi-user VPN control plane
# shellcheck disable=SC2154

module_vpn_users_install() {
  pkg_install jq qrencode openssl

  mkdir -p /opt/freshvps/lib /opt/freshvps/bin
  install -m 644 "${FRESHVPS_ROOT}/lib/common.sh" /opt/freshvps/lib/common.sh
  install -m 644 "${FRESHVPS_ROOT}/lib/vpn-core.sh" /opt/freshvps/lib/vpn-core.sh
  install -m 755 "${FRESHVPS_ROOT}/bin/freshvps-vpn" /opt/freshvps/bin/freshvps-vpn
  install -m 755 "${FRESHVPS_ROOT}/bin/freshvps-doctor" /opt/freshvps/bin/freshvps-doctor
  ln -sfn /opt/freshvps/bin/freshvps-vpn /usr/local/bin/freshvps-vpn
  ln -sfn /opt/freshvps/bin/freshvps-doctor /usr/local/bin/freshvps-doctor

  if [[ -n "${PUBLIC_IP:-}" ]]; then
    printf '%s\n' "${PUBLIC_IP}" >"${FRESHVPS_ETC}/public_ip"
    chmod 644 "${FRESHVPS_ETC}/public_ip"
  fi

  # shellcheck source=/dev/null
  source "${FRESHVPS_ROOT}/lib/vpn-core.sh"
  vpn_ensure_dirs
  vpn_seed_operator

  info "Operator VPN: /etc/freshvps/clients/operator/"
  info "CLI: freshvps-vpn | doctor: freshvps-doctor"
}

module_vpn_users_uninstall() {
  rm -f /usr/local/bin/freshvps-vpn /usr/local/bin/freshvps-doctor
  info "CLI removed (registry left under /etc/freshvps)"
}
