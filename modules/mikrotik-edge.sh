#!/usr/bin/env bash
# Apply MikroTik routes so a subnet uses edge-client as gateway (SSH/API)
# shellcheck disable=SC2154

module_mikrotik_edge_install() {
  local host="${MIKROTIK_HOST:-}"
  local user="${MIKROTIK_USER:-admin}"
  local pass="${MIKROTIK_PASSWORD:-}"
  local edge_ip="${EDGE_CLIENT_IP:-}"
  local vpn_subnet="${VPN_SUBNET:-192.168.89.0/24}"
  local vpn_gw_name="${VPN_GW_COMMENT:-freshvps-edge}"

  [[ -n "${host}" ]] || die "MIKROTIK_HOST is required"
  [[ -n "${edge_ip}" ]] || die "EDGE_CLIENT_IP is required (IP of Pi/Banana/A95X on LAN)"

  pkg_install sshpass openssh-client || true

  local ssh_opts=(-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ConnectTimeout=10)
  local run_ssh
  if [[ -n "${pass}" ]]; then
    run_ssh() { sshpass -p "${pass}" ssh "${ssh_opts[@]}" "${user}@${host}" "$@"; }
  else
    run_ssh() { ssh "${ssh_opts[@]}" "${user}@${host}" "$@"; }
  fi

  info "Pushing MikroTik config to ${host} (edge ${edge_ip}, subnet ${vpn_subnet})"

  # RouterOS commands: dedicated VPN subnet routed via edge host
  # Safe comments; does not wipe existing config
  local cmds
  cmds=$(cat <<EOF
/ip route remove [find comment="${vpn_gw_name}"]
/ip route add dst-address=0.0.0.0/0 gateway=${edge_ip} distance=2 check-gateway=ping comment="${vpn_gw_name}-default-backup"
/ip firewall address-list remove [find list=freshvps-vpn-clients]
# Example: mark nothing by default — admin adds IPs to list freshvps-vpn-clients
/ip firewall mangle remove [find comment~"${vpn_gw_name}"]
/ip firewall mangle add chain=prerouting src-address-list=freshvps-vpn-clients action=mark-routing new-routing-mark=freshvps-vpn passthrough=yes comment="${vpn_gw_name}-mark"
/ip route remove [find routing-mark=freshvps-vpn]
/ip route add dst-address=0.0.0.0/0 gateway=${edge_ip} routing-mark=freshvps-vpn distance=1 check-gateway=ping comment="${vpn_gw_name}"
/ip dns set allow-remote-requests=yes
EOF
)

  if ! run_ssh "${cmds}"; then
    warn "SSH apply failed — print manual RouterOS commands:"
    echo "${cmds}"
    return 1
  fi

  cat > "${FRESHVPS_ETC}/edge/mikrotik-applied.txt" <<EOF
host=${host}
edge_ip=${edge_ip}
vpn_subnet=${vpn_subnet}
mode=policy-routing-list-freshvps-vpn-clients
Add client IPs: /ip firewall address-list add list=freshvps-vpn-clients address=x.x.x.x
EOF
  info "MikroTik: policy route via ${edge_ip} for address-list freshvps-vpn-clients"
  info "Add devices: /ip firewall address-list add list=freshvps-vpn-clients address=<device-ip>"
}
