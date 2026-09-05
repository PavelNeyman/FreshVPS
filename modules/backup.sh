#!/usr/bin/env bash
# Module: restic backup scaffolding
# shellcheck disable=SC2154

module_backup_install() {
  pkg_install restic

  mkdir -p /opt/freshvps-backup /var/backups/freshvps

  if [[ ! -f "${FRESHVPS_ETC}/secrets/restic_password" ]]; then
    write_secret restic_password "$(random_hex 24)"
  fi

  local repo="${BACKUP_REPO:-/var/backups/freshvps/repo}"
  export RESTIC_PASSWORD_FILE="${FRESHVPS_ETC}/secrets/restic_password"

  if [[ ! -f "${repo}/config" ]]; then
    restic init --repo "${repo}" || true
  fi

  cat >/opt/freshvps-backup/backup.sh <<EOF
#!/usr/bin/env bash
set -euo pipefail
export RESTIC_PASSWORD_FILE=${FRESHVPS_ETC}/secrets/restic_password
REPO=${repo}
restic -r "\${REPO}" backup \\
  /etc/freshvps \\
  /usr/local/etc/sing-box \\
  /etc/blocky \\
  /opt/opensoho \\
  /opt/uptime-kuma/data \\
  /opt/beszel/hub-data \\
  --exclude-caches || true
restic -r "\${REPO}" forget --keep-daily 7 --keep-weekly 4 --prune || true
EOF
  chmod 700 /opt/freshvps-backup/backup.sh

  cat >/etc/systemd/system/freshvps-backup.service <<'EOF'
[Unit]
Description=FreshVPS restic backup

[Service]
Type=oneshot
ExecStart=/opt/freshvps-backup/backup.sh
EOF

  cat >/etc/systemd/system/freshvps-backup.timer <<'EOF'
[Unit]
Description=Daily FreshVPS backup

[Timer]
OnCalendar=daily
Persistent=true

[Install]
WantedBy=timers.target
EOF

  systemctl daemon-reload
  systemctl enable --now freshvps-backup.timer
  info "restic repo ${repo}; password in ${FRESHVPS_ETC}/secrets/restic_password; daily timer enabled"
}
