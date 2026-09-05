#!/usr/bin/env bash
# Module: Telegram operator panel
# shellcheck disable=SC2154

module_telegram_install() {
  pkg_install jq curl

  if [[ -n "${TELEGRAM_BOT_TOKEN:-}" ]]; then
    write_secret telegram_bot_token "${TELEGRAM_BOT_TOKEN}"
  fi
  if [[ -n "${TELEGRAM_ADMIN_ID:-}" ]]; then
    write_secret telegram_admin_id "${TELEGRAM_ADMIN_ID}"
  fi

  mkdir -p /opt/freshvps/runtime/telegram
  install -m 700 "${FRESHVPS_ROOT}/runtime/telegram/notify.sh" /opt/freshvps/runtime/telegram/notify.sh
  install -m 700 "${FRESHVPS_ROOT}/runtime/telegram/status.sh" /opt/freshvps/runtime/telegram/status.sh
  install -m 700 "${FRESHVPS_ROOT}/runtime/telegram/bot.sh" /opt/freshvps/runtime/telegram/bot.sh
  # compat symlinks for old paths
  mkdir -p /opt/freshvps-telegram
  ln -sfn /opt/freshvps/runtime/telegram/notify.sh /opt/freshvps-telegram/notify.sh
  ln -sfn /opt/freshvps/runtime/telegram/status.sh /opt/freshvps-telegram/status.sh
  ln -sfn /opt/freshvps/runtime/telegram/bot.sh /opt/freshvps-telegram/bot.sh

  cat >/etc/systemd/system/freshvps-telegram-bot.service <<'EOF'
[Unit]
Description=FreshVPS Telegram operator bot
After=network-online.target

[Service]
Type=simple
ExecStart=/opt/freshvps/runtime/telegram/bot.sh
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

  if [[ -f "${FRESHVPS_ETC}/secrets/telegram_bot_token" && -f "${FRESHVPS_ETC}/secrets/telegram_admin_id" ]]; then
    systemd_enable_start freshvps-telegram-bot
    /opt/freshvps/runtime/telegram/notify.sh "FreshVPS: operator bot online on $(hostname)"
  else
    info "Telegram secrets incomplete — unit installed, not started"
  fi
}

module_telegram_uninstall() {
  systemctl disable --now freshvps-telegram-bot 2>/dev/null || true
  rm -f /etc/systemd/system/freshvps-telegram-bot.service
  systemctl daemon-reload 2>/dev/null || true
}
