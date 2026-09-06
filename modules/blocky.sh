#!/usr/bin/env bash
# Module: blocky
# shellcheck disable=SC2154

module_blocky_install() {
  local arch conf_dir=/etc/blocky
  arch="$(arch_go)"
  local asset_arch
  case "${arch}" in
    amd64) asset_arch=x86_64 ;;
    arm64) asset_arch=arm64 ;;
    armv7) asset_arch=armv7 ;;
    *) die "Unsupported arch for blocky: ${arch}" ;;
  esac

  pkg_install curl tar

  local tag url tmp
  tag="$(curl -fsSL https://api.github.com/repos/0xERR0R/blocky/releases/latest | jq -r .tag_name)"
  [[ -n "${tag}" && "${tag}" != null ]] || die "Cannot resolve blocky release"
  url="https://github.com/0xERR0R/blocky/releases/download/${tag}/blocky_${tag}_Linux_${asset_arch}.tar.gz"
  tmp="$(mktemp -d)"
  info "Downloading blocky ${tag}"
  if ! curl -fsSL "${url}" -o "${tmp}/b.tgz"; then
    url="https://github.com/0xERR0R/blocky/releases/download/${tag}/blocky_${tag}_Linux_${arch}.tar.gz"
    curl -fsSL "${url}" -o "${tmp}/b.tgz" || die "Failed to download blocky"
  fi
  tar -xzf "${tmp}/b.tgz" -C "${tmp}"
  install -m 755 "${tmp}/blocky" /usr/local/bin/blocky
  rm -rf "${tmp}"

  mkdir -p "${conf_dir}"
  if [[ ! -f "${conf_dir}/config.yml" ]]; then
    cat >"${conf_dir}/config.yml" <<'EOF'
upstreams:
  groups:
    default:
      - 9.9.9.9
      - 1.1.1.1
      - https://dns.quad9.net/dns-query
  strategy: parallel_best

blocking:
  denylists:
    ads:
      - https://cdn.jsdelivr.net/gh/hagezi/dns-blocklists@latest/wildcard/multi.txt
    security:
      - https://cdn.jsdelivr.net/gh/hagezi/dns-blocklists@latest/wildcard/tif.txt
      - https://cdn.jsdelivr.net/gh/hagezi/dns-blocklists@latest/wildcard/fake.txt
  clientGroupsBlock:
    default:
      - ads
      - security
  blockType: nxDomain

ports:
  dns: 53
  http: 4000

caching:
  minTime: 5m
  maxTime: 30m
  prefetching: true

prometheus:
  enable: true
  path: /metrics

log:
  level: info
EOF
  fi

  if systemctl is-active systemd-resolved >/dev/null 2>&1; then
    mkdir -p /etc/systemd/resolved.conf.d
    cat >/etc/systemd/resolved.conf.d/freshvps-no-stub.conf <<'EOF'
[Resolve]
DNSStubListener=no
EOF
    systemctl restart systemd-resolved || true
    if [[ -L /etc/resolv.conf ]]; then
      rm -f /etc/resolv.conf
      printf 'nameserver 9.9.9.9\nnameserver 1.1.1.1\n' >/etc/resolv.conf
    fi
  fi

  cat >/etc/systemd/system/blocky.service <<'EOF'
[Unit]
Description=Blocky DNS proxy
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/blocky --config /etc/blocky/config.yml
Restart=on-failure
RestartSec=5
AmbientCapabilities=CAP_NET_BIND_SERVICE
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/etc/blocky

[Install]
WantedBy=multi-user.target
EOF

  systemd_enable_start blocky
  firewall_allow_tcp 53 "blocky-dns"
  firewall_allow_udp 53 "blocky-dns"
  # API localhost/VPN only
  # firewall_allow_tcp 4000 "blocky-api"

  printf 'nameserver 127.0.0.1\noptions edns0\n' >/etc/resolv.conf
  info "Blocky listening on :53 (API :4000). Lists: HaGeZi multi + tif + fake"
}

module_blocky_uninstall() {
  systemctl disable --now blocky 2>/dev/null || true
  rm -f /etc/systemd/system/blocky.service
  systemctl daemon-reload 2>/dev/null || true
  info "blocky stopped (config left under /etc/blocky)"
}
