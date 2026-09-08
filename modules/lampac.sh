#!/usr/bin/env bash
# Module: Lampac optional — compose profile lampac
# shellcheck disable=SC2154

module_lampac_install() {
  local port="${LAMPAC_PORT:-9118}"
  local dir=/opt/lampac
  # shellcheck source=/dev/null
  source "${FRESHVPS_ROOT}/modules/panels.sh"

  mkdir -p "${dir}/config" "${dir}/cache" "${dir}/database"

  local pass
  pass="${LAMPAC_ROOT_PASSWORD:-$(read_secret lampac_root_password || true)}"
  [[ -n "${pass}" ]] || pass="$(random_hex 16)"
  write_secret lampac_root_password "${pass}"
  printf '%s' "${pass}" >"${dir}/config/passwd"

  # init.conf must be readable by container UID 1000 (lampac). Root-only 600 → defaults applied (Chromium ON).
  cat >"${dir}/config/init.conf" <<'EOF'
{
  "lowMemoryMode": true,
  "listen": { "ip": "0.0.0.0", "port": 9118, "scheme": "http", "localhost": "127.0.0.1" },
  "chromium": { "enable": false },
  "firefox": { "enable": false },
  "online": { "name": "FreshVPS Lampac", "version": true }
}
EOF

  chown -R 1000:1000 "${dir}/config" "${dir}/cache" "${dir}/database" 2>/dev/null || true
  chmod 644 "${dir}/config/passwd" "${dir}/config/init.conf" 2>/dev/null || true

  panels_compose_up lampac
  info "Lampac on 127.0.0.1:${port} (compose profile lampac; chromium disabled in init.conf)"
  info "Password: ${FRESHVPS_ETC}/secrets/lampac_root_password"
}

module_lampac_uninstall() {
  if [[ -f /opt/freshvps/compose/panels.yml ]]; then
    (cd /opt/freshvps/compose && docker compose -f panels.yml --profile lampac rm -sf 2>/dev/null) || true
  fi
  docker rm -f lampac 2>/dev/null || true
  info "Lampac container removed (data kept under /opt/lampac unless --purge)"
}
