#!/usr/bin/env bash
# Module: Uptime Kuma (Docker)
# shellcheck disable=SC2154

module_kuma_install() {
  local port="${KUMA_PORT:-3001}"
  local dir=/opt/uptime-kuma

  if ! command -v docker >/dev/null 2>&1; then
    info "Installing Docker"
    pkg_install ca-certificates curl
    curl -fsSL https://get.docker.com | sh
    systemctl enable --now docker
  fi

  mkdir -p "${dir}/data"
  cat >"${dir}/docker-compose.yml" <<EOF
services:
  uptime-kuma:
    image: louislam/uptime-kuma:1
    container_name: uptime-kuma
    restart: unless-stopped
    volumes:
      - ${dir}/data:/app/data
    ports:
      - "127.0.0.1:${port}:3001"
    security_opt:
      - no-new-privileges:true
EOF

  (cd "${dir}" && docker compose up -d)
  # panels: localhost only
  # firewall_allow_tcp "${port}" "uptime-kuma"
  info "Uptime Kuma: http://${PUBLIC_IP:-SERVER}:${port} — complete setup in browser"
}

module_kuma_uninstall() {
  if [[ -f /opt/uptime-kuma/docker-compose.yml ]]; then
    (cd /opt/uptime-kuma && docker compose down) 2>/dev/null || true
  fi
  info "Uptime Kuma stopped (data left under /opt/uptime-kuma/data)"
}
