#!/usr/bin/env bash
# Module: Cloudflare Tunnel → local Admin API (8787)
# Requires: CLOUDFLARE_TUNNEL_TOKEN from Zero Trust dashboard
# shellcheck disable=SC2154

module_cloudflared_install() {
  local token="${CLOUDFLARE_TUNNEL_TOKEN:-$(cat /etc/freshvps/secrets/cloudflare_tunnel_token 2>/dev/null || true)}"
  if [[ -z "${token}" ]]; then
    warn "Set CLOUDFLARE_TUNNEL_TOKEN or write /etc/freshvps/secrets/cloudflare_tunnel_token"
    warn "Create tunnel in Cloudflare Zero Trust → Networks → Tunnels → Docker/Linux"
    warn "Public hostname → http://127.0.0.1:8787"
    return 0
  fi
  mkdir -p /etc/freshvps/secrets
  printf '%s\n' "${token}" >/etc/freshvps/secrets/cloudflare_tunnel_token
  chmod 600 /etc/freshvps/secrets/cloudflare_tunnel_token

  if ! command -v cloudflared >/dev/null 2>&1; then
    local arch
    arch="$(uname -m)"
    case "${arch}" in
      x86_64|amd64) arch=amd64 ;;
      aarch64|arm64) arch=arm64 ;;
      *) die "unsupported arch for cloudflared: ${arch}" ;;
    esac
    curl -fsSL "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-${arch}" \
      -o /usr/local/bin/cloudflared
    chmod 755 /usr/local/bin/cloudflared
  fi

  cat >/etc/systemd/system/cloudflared.service <<'UNIT'
[Unit]
Description=Cloudflare Tunnel (FreshVPS Admin)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
EnvironmentFile=-/etc/freshvps/secrets/cloudflare_tunnel_token
ExecStart=/bin/bash -c 'exec /usr/local/bin/cloudflared tunnel --no-autoupdate run --token "$(cat /etc/freshvps/secrets/cloudflare_tunnel_token)"'
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
UNIT
  # EnvironmentFile won't work for token file as KEY=; use ExecStart as above
  systemctl daemon-reload
  systemctl enable --now cloudflared
  info "cloudflared running — map hostname in CF dashboard to http://127.0.0.1:8787"
}

module_cloudflared_uninstall() {
  systemctl disable --now cloudflared 2>/dev/null || true
  rm -f /etc/systemd/system/cloudflared.service
  systemctl daemon-reload 2>/dev/null || true
  info "cloudflared removed (binary kept)"
}
