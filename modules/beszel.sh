#!/usr/bin/env bash
# Module: Beszel hub + agent on same VPS
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

  mkdir -p "${dir}/hub-data"
  cat >"${dir}/docker-compose.yml" <<EOF
services:
  beszel:
    image: henrygd/beszel:latest
    container_name: beszel
    restart: unless-stopped
    ports:
      - "${port}:8090"
    volumes:
      - ${dir}/hub-data:/beszel_data
EOF

  (cd "${dir}" && docker compose up -d)
  firewall_allow_tcp "${port}" "beszel-hub"

  cat >"${dir}/README.agent" <<EOF
Beszel hub: http://${PUBLIC_IP:-SERVER}:${port}

After creating an account in the UI, add this system and deploy the agent
(binary or Docker) using the KEY / TOKEN from the hub UI.

Official agent install: https://beszel.dev/guide/agent-installation
EOF
  info "Beszel hub on :${port}. Add agent via UI — see ${dir}/README.agent"
}
