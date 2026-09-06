#!/usr/bin/env bash
# Module: OpenSOHO (OpenWrt central management)
# Prefer Docker (ghcr.io/opensoho/opensoho); binary from GitHub releases as fallback.
# shellcheck disable=SC2154

module_opensoho_install() {
  local port="${OPENSOHO_HTTP_PORT:-8090}"
  local dir=/opt/opensoho
  pkg_install curl jq

  mkdir -p "${dir}" /var/lib/opensoho

  local secret
  secret="$(read_secret opensoho_shared_secret || true)"
  if [[ -z "${secret}" ]]; then
    secret="$(random_hex 24)"
    write_secret opensoho_shared_secret "${secret}"
  fi

  if ! command -v docker >/dev/null 2>&1; then
    info "Installing Docker for OpenSOHO"
    curl -fsSL https://get.docker.com | sh
    systemctl enable --now docker
  fi

  local image="${OPENSOHO_IMAGE:-ghcr.io/opensoho/opensoho:latest}"

  # Persist PocketBase data under /data (host: /var/lib/opensoho)
  cat >"${dir}/docker-compose.yml" <<EOF
services:
  opensoho:
    image: ${image}
    container_name: opensoho
    restart: unless-stopped
    command: ["serve", "--http", "0.0.0.0:8090", "--dir", "/data"]
    environment:
      OPENSOHO_SHARED_SECRET: "${secret}"
    ports:
      - "127.0.0.1:${port}:8090"
    volumes:
      - /var/lib/opensoho:/data
EOF

  if (cd "${dir}" && docker compose pull && docker compose up -d); then
    _opensoho_bootstrap_admin
    info "OpenSOHO (Docker) on 127.0.0.1:${port} (SSH tunnel)"
    info "Shared secret: ${FRESHVPS_ETC}/secrets/opensoho_shared_secret"
    info "Admin: ${FRESHVPS_ETC}/secrets/opensoho_admin_email + opensoho_admin_password"
    return 0
  fi

  warn "Docker pull failed; trying GitHub release binary"
  local tag arch asset_url tmp
  arch="$(arch_go)"
  tag="$(curl -fsSL https://api.github.com/repos/rubenbe/opensoho/releases/latest | jq -r .tag_name)"
  asset_url="$(curl -fsSL https://api.github.com/repos/rubenbe/opensoho/releases/latest | jq -r --arg arch "${arch}" '
    .assets[]
    | select(.name | test("linux"; "i"))
    | select(.name | test($arch) or test("x86_64") and $arch=="amd64" or test("aarch64") and $arch=="arm64")
    | .browser_download_url
  ' | head -1)"
  if [[ -z "${asset_url}" || "${asset_url}" == null ]]; then
    asset_url="$(curl -fsSL https://api.github.com/repos/rubenbe/opensoho/releases/latest | jq -r '.assets[] | select(.name|test("linux";"i")) | .browser_download_url' | head -1)"
  fi

  if [[ -n "${asset_url}" && "${asset_url}" != null ]]; then
    tmp="$(mktemp -d)"
    curl -fsSL "${asset_url}" -o "${tmp}/os.asset"
    if file "${tmp}/os.asset" | grep -qiE 'gzip|tar|Zip'; then
      tar -xzf "${tmp}/os.asset" -C "${tmp}" 2>/dev/null || tar -xf "${tmp}/os.asset" -C "${tmp}" || true
      local bin
      bin="$(find "${tmp}" -type f -name 'opensoho' | head -1)"
      [[ -n "${bin}" ]] || bin="$(find "${tmp}" -type f -executable ! -name '*.sh' | head -1)"
      install -m 755 "${bin}" "${dir}/opensoho"
    else
      install -m 755 "${tmp}/os.asset" "${dir}/opensoho"
    fi
    rm -rf "${tmp}"
  fi

  if [[ -x "${dir}/opensoho" ]]; then
    cat >/etc/systemd/system/opensoho.service <<EOF
[Unit]
Description=OpenSOHO OpenWrt management
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=/var/lib/opensoho
Environment=OPENSOHO_SHARED_SECRET=${secret}
ExecStart=${dir}/opensoho serve --http 127.0.0.1:${port} --dir /var/lib/opensoho
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
    systemd_enable_start opensoho
    info "OpenSOHO binary service on 127.0.0.1:${port}"
    return 0
  fi

  cat >"${dir}/README.freshvps" <<EOF
OpenSOHO install incomplete.
Shared secret: see ${FRESHVPS_ETC}/secrets/opensoho_shared_secret
Docs: https://opensoho.github.io/docs
Image: ${image}
EOF
  warn "OpenSOHO not started — see ${dir}/README.freshvps"
}

# Create or update superuser; store credentials under secrets/
_opensoho_bootstrap_admin() {
  local email pass i
  email="${OPENSOHO_ADMIN_EMAIL:-$(read_secret opensoho_admin_email || true)}"
  pass="${OPENSOHO_ADMIN_PASSWORD:-$(read_secret opensoho_admin_password || true)}"
  if [[ -z "${email}" ]]; then
    email="admin@freshvps.local"
  fi
  if [[ -z "${pass}" ]]; then
    pass="$(random_hex 16)"
  fi

  # Wait for container + PocketBase
  for i in 1 2 3 4 5 6 7 8 9 10; do
    if docker exec opensoho /ko-app/opensoho superuser upsert "${email}" "${pass}" --dir /data >/dev/null 2>&1; then
      write_secret opensoho_admin_email "${email}"
      write_secret opensoho_admin_password "${pass}"
      info "OpenSOHO admin ready: ${email} (password in secrets/opensoho_admin_password)"
      return 0
    fi
    sleep 2
  done
  warn "OpenSOHO admin bootstrap failed — run: docker exec opensoho /ko-app/opensoho superuser upsert EMAIL PASS --dir /data"
}

module_opensoho_uninstall() {
  if [[ -f /opt/opensoho/docker-compose.yml ]]; then
    (cd /opt/opensoho && docker compose down) 2>/dev/null || true
  fi
  systemctl disable --now opensoho 2>/dev/null || true
  rm -f /etc/systemd/system/opensoho.service
  systemctl daemon-reload 2>/dev/null || true
  info "OpenSOHO stopped (data left in /var/lib/opensoho)"
}
