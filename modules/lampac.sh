#!/usr/bin/env bash
# Module: Lampac NextGen (Lampa backend) — light profile for small VPS
# Image: ghcr.io/lampac-nextgen/lampac
# shellcheck disable=SC2154

module_lampac_install() {
  local port="${LAMPAC_PORT:-9118}"
  local dir=/opt/lampac
  local image="${LAMPAC_IMAGE:-ghcr.io/lampac-nextgen/lampac:latest}"

  if ! command -v docker >/dev/null 2>&1; then
    info "Installing Docker"
    pkg_install ca-certificates curl
    curl -fsSL https://get.docker.com | sh
    systemctl enable --now docker
  fi

  mkdir -p "${dir}/config" "${dir}/cache" "${dir}/database" "${dir}/plugins"

  local pass
  pass="${LAMPAC_ROOT_PASSWORD:-$(read_secret lampac_root_password || true)}"
  if [[ -z "${pass}" ]]; then
    pass="$(random_hex 16)"
  fi
  write_secret lampac_root_password "${pass}"
  printf "%s" "${pass}" >"${dir}/config/passwd"
  chmod 600 "${dir}/config/passwd"

  # Light profile: lowMemoryMode, no Chromium/Playwright, skip heavy modules.
  # Online stays on. JacRed + TorrServer enabled for torrent path.
  if [[ ! -f "${dir}/config/init.conf" ]]; then
    cat >"${dir}/config/init.conf" <<EOF
{
  "lowMemoryMode": true,
  "GC": {
    "Concurrent": true,
    "ConserveMemory": 2,
    "HighMemoryPercent": 80,
    "RetainVM": false
  },
  "listen": {
    "version": true,
    "ip": "0.0.0.0",
    "port": 9118,
    "scheme": "http",
    "localhost": "127.0.0.1"
  },
  "BaseModule": {
    "SkipModules": [
      "Catalog",
      "DLNA",
      "Sync",
      "TimeCode",
      "Tracks",
      "Transcoding",
      "WebLog"
    ]
  },
  "chromium": {
    "enable": false
  },
  "firefox": {
    "enable": false
  },
  "openstat": {
    "enable": false
  },
  "online": {
    "name": "FreshVPS Lampac",
    "spiderName": "FreshVPS",
    "version": true,
    "btn_priority_forced": true
  },
  "sisi": {
    "lgbt": false,
    "spider": false,
    "history": { "enable": false }
  },
  "LampaWeb": {
    "initPlugins": {
      "torrserver": true
    }
  }
}
EOF
  fi

  # Optional override from env: LAMPAC_SKIP_TORRENT=1 skips JacRed+TorrServer too
  if [[ "${LAMPAC_SKIP_TORRENT:-0}" == "1" ]]; then
    warn "LAMPAC_SKIP_TORRENT=1 — JacRed/TorrServer not forced; edit init.conf if needed"
  fi

  cat >"${dir}/docker-compose.yml" <<EOF
services:
  lampac:
    image: ${image}
    container_name: lampac
    restart: unless-stopped
    ports:
      - "127.0.0.1:${port}:9118"
    # No Playwright → small shm is enough
    shm_size: 128mb
    mem_limit: ${LAMPAC_MEM_LIMIT:-1536m}
    volumes:
      - ${dir}/config/passwd:/lampac/passwd:ro
      - ${dir}/config/init.conf:/lampac/init.conf
      - ${dir}/cache:/lampac/cache
      - ${dir}/database:/lampac/database
EOF

  (cd "${dir}" && docker compose pull && docker compose up -d)

  info "Lampac NextGen (light) on 127.0.0.1:${port}"
  info "Root password: ${FRESHVPS_ETC}/secrets/lampac_root_password"
  info "Access: ssh -L ${port}:127.0.0.1:${port} root@VPS → http://127.0.0.1:${port}"
  info "Lampa plugins: http://127.0.0.1:${port}/online.js  and  /ts.js (TorrServer)"
  info "Edit ${dir}/config/init.conf then: cd ${dir} && docker compose restart"
}

module_lampac_uninstall() {
  if [[ -f /opt/lampac/docker-compose.yml ]]; then
    (cd /opt/lampac && docker compose down) 2>/dev/null || true
  fi
  info "Lampac stopped (data left under /opt/lampac)"
}
