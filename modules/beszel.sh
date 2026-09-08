#!/usr/bin/env bash
# Module: Beszel hub via compose/panels.yml
# shellcheck disable=SC2154

module_beszel_install() {
  local port="${BESZEL_PORT:-8091}"
  # shellcheck source=/dev/null
  source "${FRESHVPS_ROOT}/modules/panels.sh"

  mkdir -p /opt/beszel/hub-data /opt/beszel/agent-data /opt/beszel/socket

  local key="${BESZEL_KEY:-$(read_secret beszel_key || true)}"
  local token="${BESZEL_TOKEN:-$(read_secret beszel_token || true)}"
  [[ -n "${key}" ]] && write_secret beszel_key "${key}"
  [[ -n "${token}" ]] && write_secret beszel_token "${token}"

  panels_compose_up beszel
  _beszel_bootstrap_admin

  cat >/opt/beszel/enable-agent.sh <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
KEY="${BESZEL_KEY:?}"
TOKEN="${BESZEL_TOKEN:?}"
umask 077
printf '%s\n' "${KEY}" >/etc/freshvps/secrets/beszel_key
printf '%s\n' "${TOKEN}" >/etc/freshvps/secrets/beszel_token
# shellcheck source=/dev/null
source /opt/freshvps/lib/common.sh 2>/dev/null || true
if [[ -f /opt/freshvps/compose/panels.yml ]]; then
  cd /opt/freshvps/compose
  grep -q BESZEL_KEY .env 2>/dev/null || true
  sed -i "s|^BESZEL_KEY=.*|BESZEL_KEY=${KEY}|" .env 2>/dev/null || echo "BESZEL_KEY=${KEY}" >>.env
  sed -i "s|^BESZEL_TOKEN=.*|BESZEL_TOKEN=${TOKEN}|" .env 2>/dev/null || echo "BESZEL_TOKEN=${TOKEN}" >>.env
  docker compose --env-file .env -f panels.yml --profile beszel-agent up -d
fi
echo "Agent started."
EOF
  chmod 700 /opt/beszel/enable-agent.sh

  if [[ -n "${key}" && -n "${token}" ]]; then
    panels_compose_up beszel beszel-agent
    info "Beszel hub+agent on 127.0.0.1:${port}"
  else
    info "Beszel hub on 127.0.0.1:${port}. Add System in UI, then enable-agent.sh"
  fi
}

_beszel_bootstrap_admin() {
  local email pass i
  email="${BESZEL_ADMIN_EMAIL:-$(read_secret beszel_admin_email || true)}"
  pass="${BESZEL_ADMIN_PASSWORD:-$(read_secret beszel_admin_password || true)}"
  [[ -n "${email}" ]] || email="admin@freshvps.local"
  [[ -n "${pass}" ]] || pass="$(random_hex 16)"
  for i in $(seq 1 15); do
    if docker exec beszel /beszel superuser upsert "${email}" "${pass}" --dir /beszel_data >/dev/null 2>&1; then
      write_secret beszel_admin_email "${email}"
      write_secret beszel_admin_password "${pass}"
      info "Beszel admin: ${email}"
      return 0
    fi
    sleep 2
  done
  warn "Beszel admin bootstrap failed"
}

module_beszel_uninstall() {
  if [[ -f /opt/freshvps/compose/panels.yml ]]; then
    (cd /opt/freshvps/compose && docker compose -f panels.yml --profile beszel-agent --profile beszel rm -sf 2>/dev/null) || true
  fi
  docker rm -f beszel beszel-agent 2>/dev/null || true
  info "Beszel containers removed (data kept under /opt/beszel unless --purge)"
}
