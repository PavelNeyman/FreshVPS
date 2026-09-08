#!/usr/bin/env bash
# Module: OpenSOHO via compose/panels.yml profile opensoho
# shellcheck disable=SC2154

module_opensoho_install() {
  local port="${OPENSOHO_HTTP_PORT:-8090}"
  # shellcheck source=/dev/null
  source "${FRESHVPS_ROOT}/modules/panels.sh"

  local secret
  secret="$(read_secret opensoho_shared_secret || true)"
  if [[ -z "${secret}" ]]; then
    secret="$(random_hex 24)"
    write_secret opensoho_shared_secret "${secret}"
  fi

  mkdir -p /var/lib/opensoho
  panels_compose_up opensoho
  _opensoho_bootstrap_admin
  info "OpenSOHO on 127.0.0.1:${port} (compose profile opensoho)"
  info "Secret: ${FRESHVPS_ETC}/secrets/opensoho_shared_secret"
}

_opensoho_bootstrap_admin() {
  local email pass i
  email="${OPENSOHO_ADMIN_EMAIL:-$(read_secret opensoho_admin_email || true)}"
  pass="${OPENSOHO_ADMIN_PASSWORD:-$(read_secret opensoho_admin_password || true)}"
  [[ -n "${email}" ]] || email="admin@freshvps.local"
  [[ -n "${pass}" ]] || pass="$(random_hex 16)"

  for i in $(seq 1 15); do
    if docker exec opensoho /ko-app/opensoho superuser upsert "${email}" "${pass}" --dir /data >/dev/null 2>&1 \
      || docker exec opensoho opensoho superuser upsert "${email}" "${pass}" --dir /data >/dev/null 2>&1; then
      write_secret opensoho_admin_email "${email}"
      write_secret opensoho_admin_password "${pass}"
      info "OpenSOHO admin: ${email}"
      return 0
    fi
    sleep 2
  done
  warn "OpenSOHO admin bootstrap failed — create manually via docker exec"
}

module_opensoho_uninstall() {
  if [[ -f /opt/freshvps/compose/panels.yml ]]; then
    (cd /opt/freshvps/compose && docker compose -f panels.yml --profile opensoho rm -sf 2>/dev/null) || true
  fi
  docker rm -f opensoho 2>/dev/null || true
  systemctl disable --now opensoho 2>/dev/null || true
  info "OpenSOHO container removed (data kept in /var/lib/opensoho unless --purge)"
}
