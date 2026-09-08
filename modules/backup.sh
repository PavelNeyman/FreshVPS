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
  /var/lib/opensoho \\
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

module_backup_uninstall() {
  systemctl disable --now freshvps-backup.timer 2>/dev/null || true
  rm -f /etc/systemd/system/freshvps-backup.timer /etc/systemd/system/freshvps-backup.service
  systemctl daemon-reload 2>/dev/null || true
  rm -rf /opt/freshvps-backup
  # restic repo under /var/backups/freshvps kept unless --purge
  info "Backup timer/scripts removed (restic repo kept unless --purge)"
}
