#!/usr/bin/env bash
# Module: OpenSOHO (OpenWrt central management)
# https://github.com/rubenbe/opensoho — binary server + openwisp-config on routers
# shellcheck disable=SC2154

module_opensoho_install() {
  local arch port="${OPENSOHO_HTTP_PORT:-8090}"
  arch="$(arch_go)"
  pkg_install curl jq

  local tag asset url tmp bin
  tag="$(curl -fsSL https://api.github.com/repos/rubenbe/opensoho/releases/latest | jq -r .tag_name)"
  if [[ -z "${tag}" || "${tag}" == null ]]; then
    warn "OpenSOHO releases API empty; trying docker fallback note"
    tag=""
  fi

  mkdir -p /opt/opensoho /var/lib/opensoho

  local secret
  secret="$(read_secret opensoho_shared_secret || true)"
  if [[ -z "${secret}" ]]; then
    secret="$(random_hex 24)"
    write_secret opensoho_shared_secret "${secret}"
  fi

  if [[ -n "${tag}" ]]; then
    # Best-effort asset name discovery
    asset="$(curl -fsSL https://api.github.com/repos/rubenbe/opensoho/releases/latest | jq -r --arg a "${arch}" '.assets[] | select(.name | test("linux";"i")) | select(.name | test($a)|not | not) | .browser_download_url' 2>/dev/null | head -1 || true)"
    if [[ -z "${asset}" ]]; then
      asset="$(curl -fsSL https://api.github.com/repos/rubenbe/opensoho/releases/latest | jq -r '.assets[0].browser_download_url' 2>/dev/null || true)"
    fi
    if [[ -n "${asset}" && "${asset}" != null ]]; then
      tmp="$(mktemp -d)"
      info "Downloading OpenSOHO ${tag}"
      curl -fsSL "${asset}" -o "${tmp}/os.bin"
      if file "${tmp}/os.bin" | grep -qi 'gzip\|tar'; then
        tar -xzf "${tmp}/os.bin" -C "${tmp}" 2>/dev/null || tar -xf "${tmp}/os.bin" -C "${tmp}"
        bin="$(find "${tmp}" -type f -name 'opensoho' | head -1)"
        [[ -n "${bin}" ]] || bin="$(find "${tmp}" -type f -executable | head -1)"
        install -m 755 "${bin}" /opt/opensoho/opensoho
      else
        install -m 755 "${tmp}/os.bin" /opt/opensoho/opensoho
      fi
      rm -rf "${tmp}"
    fi
  fi

  if [[ ! -x /opt/opensoho/opensoho ]]; then
    warn "OpenSOHO binary not installed automatically. Install from https://github.com/rubenbe/opensoho/releases or Docker."
    cat >/opt/opensoho/README.freshvps <<EOF
OpenSOHO shared secret (for openwisp-config on routers):
  ${secret}

Start example:
  OPENSOHO_SHARED_SECRET=${secret} /opt/opensoho/opensoho serve --http 0.0.0.0:${port}

Docs: https://opensoho.github.io/docs
EOF
    info "Wrote /opt/opensoho/README.freshvps with shared secret pointer"
    firewall_allow_tcp "${port}" "opensoho"
    return 0
  fi

  cat >/etc/systemd/system/opensoho.service <<EOF
[Unit]
Description=OpenSOHO OpenWrt management
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=/var/lib/opensoho
Environment=OPENSOHO_SHARED_SECRET=${secret}
ExecStart=/opt/opensoho/opensoho serve --http 0.0.0.0:${port}
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

  systemd_enable_start opensoho
  firewall_allow_tcp "${port}" "opensoho"
  info "OpenSOHO on :${port}. Shared secret in ${FRESHVPS_ETC}/secrets/opensoho_shared_secret"
  info "On OpenWrt: install openwisp-config and point URL to http://${PUBLIC_IP:-SERVER}:${port}"
}
