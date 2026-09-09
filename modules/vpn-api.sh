#!/usr/bin/env bash
# LEGACY: Python Admin API (freshvps-api). Prefer modules/netductor-api.sh (Go).
# Parallel mode: both can run; cutover uses netductor only.
# Module: admin API (default 127.0.0.1) — session auth only
# shellcheck disable=SC2154

module_vpn_api_install() {
  pkg_install python3

  local port="${VPN_API_PORT:-8787}"
  local bind="${VPN_API_BIND:-127.0.0.1}"

  mkdir -p /opt/freshvps/runtime/api /etc/freshvps/sessions /opt/freshvps-api
  chmod 700 /etc/freshvps/sessions

  printf '%s\n' "${bind}" >"${FRESHVPS_ETC}/api_bind"
  printf '%s\n' "${port}" >"${FRESHVPS_ETC}/api_port"

  install -m 755 "${FRESHVPS_ROOT}/runtime/api/server.py" /opt/freshvps/runtime/api/server.py
  if [[ -f "${FRESHVPS_ROOT}/runtime/api/edge_store.py" ]]; then
    install -m 644 "${FRESHVPS_ROOT}/runtime/api/edge_store.py" /opt/freshvps/runtime/api/edge_store.py
  fi
  mkdir -p /opt/freshvps/runtime/api/admin
  if [[ -d "${FRESHVPS_ROOT}/runtime/api/admin" ]]; then
    cp -a "${FRESHVPS_ROOT}/runtime/api/admin/." /opt/freshvps/runtime/api/admin/
  fi
  ln -sfn /opt/freshvps/runtime/api/server.py /opt/freshvps-api/server.py
  info "Admin UI: http://${bind}:${port}/admin/ (session token)"

  cat >/opt/freshvps-api/rebind.sh <<'EOF'
#!/usr/bin/env bash
# Usage: rebind.sh <ip|localhost|detect> [--ufw]
set -euo pipefail
PORT="$(cat /etc/freshvps/api_port 2>/dev/null || echo 8787)"
arg="${1:-localhost}"
open_ufw=0
[[ "${2:-}" == "--ufw" ]] && open_ufw=1
case "${arg}" in
  localhost|127.0.0.1) BIND=127.0.0.1 ;;
  detect)
    # all interfaces (localhost + public) — auth still required
    BIND=0.0.0.0
    ;;
  *) BIND="${arg}" ;;
esac
echo "${BIND}" >/etc/freshvps/api_bind
mkdir -p /etc/systemd/system/freshvps-api.service.d
cat >/etc/systemd/system/freshvps-api.service.d/bind.conf <<EON
[Service]
Environment=VPN_API_BIND=${BIND}
Environment=VPN_API_PORT=${PORT}
EON
systemctl daemon-reload
systemctl restart freshvps-api
echo "API listening on ${BIND}:${PORT}"
if [[ "${open_ufw}" -eq 1 && "${BIND}" != "127.0.0.1" ]] && command -v ufw >/dev/null 2>&1; then
  ufw allow "${PORT}/tcp" comment 'freshvps-api' || true
  echo "UFW opened ${PORT}/tcp (explicit --ufw)"
elif [[ "${BIND}" != "127.0.0.1" ]]; then
  echo "Note: firewall not changed; use: rebind.sh ${BIND} --ufw  if needed"
fi
EOF
  chmod 700 /opt/freshvps-api/rebind.sh

  cat >/etc/systemd/system/freshvps-api.service <<EOF
[Unit]
Description=FreshVPS admin API
After=network-online.target

[Service]
Type=simple
WorkingDirectory=/opt/freshvps/runtime/api
Environment=VPN_API_BIND=${bind}
Environment=VPN_API_PORT=${port}
Environment=PYTHONPATH=/opt/freshvps/runtime/api
ExecStart=/usr/bin/python3 /opt/freshvps/runtime/api/server.py
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

  systemd_enable_start freshvps-api
  info "API on ${bind}:${port} — auth: freshvps-vpn session 72"
  info "Rebind: freshvps-vpn api-bind localhost|detect|<ip> [--ufw]"
}

module_vpn_api_uninstall() {
  systemctl disable --now freshvps-api 2>/dev/null || true
  rm -f /etc/systemd/system/freshvps-api.service
  rm -rf /etc/systemd/system/freshvps-api.service.d
  systemctl daemon-reload 2>/dev/null || true
  rm -rf /opt/freshvps-api
  # sessions under /etc/freshvps kept unless --purge
  info "API unit and /opt/freshvps-api removed"
}
