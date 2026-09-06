#!/usr/bin/env bash
# Module: Beszel hub + local agent scaffold
# Admin is created via CLI. Agent KEY/TOKEN still need one Hub UI step (or env).
# shellcheck disable=SC2154

module_beszel_install() {
  local port="${BESZEL_PORT:-8091}"
  local dir=/opt/beszel

  if ! command -v docker >/dev/null 2>&1; then
    info "Installing Docker"
    pkg_install ca-certificates curl
    curl -fsSL https://get.docker.com | sh
    systemctl enable --now docker
  fi

  mkdir -p "${dir}/hub-data" "${dir}/agent-data" "${dir}/socket"

  local key="${BESZEL_KEY:-$(read_secret beszel_key || true)}"
  local token="${BESZEL_TOKEN:-$(read_secret beszel_token || true)}"

  cat >"${dir}/docker-compose.yml" <<EOF
services:
  beszel:
    image: henrygd/beszel:latest
    container_name: beszel
    restart: unless-stopped
    environment:
      APP_URL: http://127.0.0.1:${port}
    ports:
      - "127.0.0.1:${port}:8090"
    volumes:
      - ${dir}/hub-data:/beszel_data
      - ${dir}/socket:/beszel_socket

  # Agent starts only when KEY and TOKEN are non-empty (set after Hub UI setup).
  beszel-agent:
    image: henrygd/beszel-agent:latest
    container_name: beszel-agent
    restart: unless-stopped
    network_mode: host
    profiles: ["agent"]
    volumes:
      - ${dir}/agent-data:/var/lib/beszel-agent
      - ${dir}/socket:/beszel_socket
      - /var/run/docker.sock:/var/run/docker.sock:ro
    environment:
      LISTEN: "/beszel_socket/beszel.sock"
      HUB_URL: "http://127.0.0.1:${port}"
      KEY: "${key}"
      TOKEN: "${token}"
EOF

  (cd "${dir}" && docker compose up -d beszel)
  _beszel_bootstrap_admin

  cat >"${dir}/enable-agent.sh" <<EOF
#!/usr/bin/env bash
# After Hub UI → Add System: copy KEY and TOKEN, then:
#   export BESZEL_KEY='...'
#   export BESZEL_TOKEN='...'
#   sudo bash /opt/beszel/enable-agent.sh
set -euo pipefail
KEY="\${BESZEL_KEY:?}"
TOKEN="\${BESZEL_TOKEN:?}"
umask 077
printf '%s\n' "\${KEY}" >/etc/freshvps/secrets/beszel_key
printf '%s\n' "\${TOKEN}" >/etc/freshvps/secrets/beszel_token
cd /opt/beszel
sed -i "s|KEY: \".*\"|KEY: \"\${KEY}\"|" docker-compose.yml
sed -i "s|TOKEN: \".*\"|TOKEN: \"\${TOKEN}\"|" docker-compose.yml
docker compose --profile agent up -d
echo "Agent started."
EOF
  chmod 700 "${dir}/enable-agent.sh"

  if [[ -n "${key}" && -n "${token}" ]]; then
    write_secret beszel_key "${key}"
    write_secret beszel_token "${token}"
    (cd "${dir}" && docker compose --profile agent up -d) || true
    info "Beszel hub + agent started on 127.0.0.1:${port}"
  else
    info "Beszel hub on 127.0.0.1:${port}. Admin in secrets. Then Add System + enable-agent.sh"
  fi
}

_beszel_bootstrap_admin() {
  local email pass i
  email="${BESZEL_ADMIN_EMAIL:-$(read_secret beszel_admin_email || true)}"
  pass="${BESZEL_ADMIN_PASSWORD:-$(read_secret beszel_admin_password || true)}"
  if [[ -z "${email}" ]]; then
    email="admin@freshvps.local"
  fi
  if [[ -z "${pass}" ]]; then
    pass="$(random_hex 16)"
  fi

  for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do
    if docker exec beszel /beszel superuser upsert "${email}" "${pass}" --dir /beszel_data >/dev/null 2>&1; then
      write_secret beszel_admin_email "${email}"
      write_secret beszel_admin_password "${pass}"
      info "Beszel admin ready: ${email} (password in secrets/beszel_admin_password)"
      return 0
    fi
    sleep 2
  done
  warn "Beszel admin bootstrap failed — run: docker exec beszel /beszel superuser upsert EMAIL PASS --dir /beszel_data"
}

module_beszel_uninstall() {
  if [[ -f /opt/beszel/docker-compose.yml ]]; then
    (cd /opt/beszel && docker compose --profile agent down) 2>/dev/null || true
    (cd /opt/beszel && docker compose down) 2>/dev/null || true
  fi
  info "Beszel containers stopped (data left under /opt/beszel)"
}
