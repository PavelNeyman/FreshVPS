#!/usr/bin/env bash
# Module: Uptime Kuma (Docker)
# First admin is created in the browser (Kuma has no stable CLI for setup).
# Optional: KUMA_SEED_MONITORS=1 after you create an API key — see seed script.
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

  cat >"${dir}/SEED-MONITORS.md" <<EOF
Uptime Kuma — first admin is created in the browser (no official CLI).

1. SSH tunnel:  ssh -L ${port}:127.0.0.1:${port} root@VPS
2. Open http://127.0.0.1:${port} and create the admin account once.
3. Settings → API Keys → create a key.
4. Optional default monitors (ping/HTTP for this host): add manually or via API.

Suggested monitors after setup:
- TCP ${PUBLIC_IP:-VPS}:443 (sing-box VLESS)
- UDP ${PUBLIC_IP:-VPS}:${SINGBOX_HY2_PORT:-8443} (Hysteria2) if supported
- HTTP http://127.0.0.1:4000 (Blocky API, from host network)
- Host / Docker resources via Beszel (preferred for metrics)
EOF

  info "Uptime Kuma: 127.0.0.1:${port} (SSH tunnel) — create admin in browser once"
  info "See ${dir}/SEED-MONITORS.md"
}

module_kuma_uninstall() {
  if [[ -f /opt/uptime-kuma/docker-compose.yml ]]; then
    (cd /opt/uptime-kuma && docker compose down) 2>/dev/null || true
  fi
  info "Uptime Kuma stopped (data left under /opt/uptime-kuma/data)"
}
