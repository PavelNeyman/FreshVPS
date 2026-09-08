#!/usr/bin/env bash
# Module: Telegram operator panel (Go bot)
# shellcheck disable=SC2154
#
# Binary resolution order:
#   1) FRESHVPS_TG_BIN (path to already-built binary, e.g. scp from Mac)
#   2) existing /opt/freshvps/bin/freshvps-tg
#   3) ${FRESHVPS_ROOT}/dist/freshvps-tg-linux-$(arch_go)
#   4) GitHub release asset for this VERSION (optional)
#   5) go build on VPS
#   6) bash bot.sh fallback

_telegram_install_binary() {
  local dest=/opt/freshvps/bin/freshvps-tg
  local arch src ver
  arch="$(arch_go)"
  mkdir -p /opt/freshvps/bin

  if [[ -n "${FRESHVPS_TG_BIN:-}" && -f "${FRESHVPS_TG_BIN}" ]]; then
    install -m 755 "${FRESHVPS_TG_BIN}" "${dest}"
    info "Telegram bot: using FRESHVPS_TG_BIN=${FRESHVPS_TG_BIN}"
    printf '%s\n' "${dest}"
    return 0
  fi

  if [[ -x "${dest}" ]]; then
    info "Telegram bot: reusing existing ${dest}"
    printf '%s\n' "${dest}"
    return 0
  fi

  src="${FRESHVPS_ROOT}/dist/freshvps-tg-linux-${arch}"
  if [[ -f "${src}" ]]; then
    install -m 755 "${src}" "${dest}"
    info "Telegram bot: installed prebuilt from dist/ (linux-${arch})"
    printf '%s\n' "${dest}"
    return 0
  fi

  ver="$(cat "${FRESHVPS_ROOT}/VERSION" 2>/dev/null || echo "")"
  if [[ -n "${ver}" ]]; then
    local url="https://github.com/PavelNeyman/FreshVPS/releases/download/v${ver}/freshvps-tg-linux-${arch}"
    if curl -fsSL "${url}" -o "${dest}.tmp" 2>/dev/null; then
      chmod 755 "${dest}.tmp"
      mv "${dest}.tmp" "${dest}"
      info "Telegram bot: downloaded release v${ver} linux-${arch}"
      printf '%s\n' "${dest}"
      return 0
    fi
    rm -f "${dest}.tmp"
  fi

  if command -v go >/dev/null 2>&1 || pkg_install golang-go 2>/dev/null; then
    info "Telegram bot: building with Go on VPS (slow)"
    ( cd "${FRESHVPS_ROOT}/cmd/freshvps-tg" && go build -o "${dest}" -trimpath -ldflags='-s -w' . )
    chmod 755 "${dest}"
    printf '%s\n' "${dest}"
    return 0
  fi

  if [[ -f "${FRESHVPS_ROOT}/runtime/telegram/bot.sh" ]]; then
    warn "Telegram bot: no prebuilt/Go — bash fallback"
    printf '%s\n' /opt/freshvps/runtime/telegram/bot.sh
    return 0
  fi

  die "Cannot install Telegram bot binary"
}

module_telegram_install() {
  pkg_install jq curl

  if [[ -n "${TELEGRAM_BOT_TOKEN:-}" ]]; then
    write_secret telegram_bot_token "${TELEGRAM_BOT_TOKEN}"
  fi
  if [[ -n "${TELEGRAM_ADMIN_ID:-}" ]]; then
    write_secret telegram_admin_id "${TELEGRAM_ADMIN_ID}"
  fi

  mkdir -p /opt/freshvps/runtime/telegram /opt/freshvps/bin /opt/freshvps-telegram
  install -m 700 "${FRESHVPS_ROOT}/runtime/telegram/notify.sh" /opt/freshvps/runtime/telegram/notify.sh
  install -m 700 "${FRESHVPS_ROOT}/runtime/telegram/status.sh" /opt/freshvps/runtime/telegram/status.sh
  if [[ -f "${FRESHVPS_ROOT}/runtime/telegram/bot.sh" ]]; then
    install -m 700 "${FRESHVPS_ROOT}/runtime/telegram/bot.sh" /opt/freshvps/runtime/telegram/bot.sh
  fi
  ln -sfn /opt/freshvps/runtime/telegram/notify.sh /opt/freshvps-telegram/notify.sh
  ln -sfn /opt/freshvps/runtime/telegram/status.sh /opt/freshvps-telegram/status.sh

  local bin
  bin="$(_telegram_install_binary)"

  cat >/etc/systemd/system/freshvps-telegram-bot.service <<EOF
[Unit]
Description=FreshVPS Telegram operator bot
After=network-online.target

[Service]
Type=simple
ExecStart=${bin}
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

  if [[ -f "${FRESHVPS_ETC}/secrets/telegram_bot_token" && -f "${FRESHVPS_ETC}/secrets/telegram_admin_id" ]]; then
    systemd_enable_start freshvps-telegram-bot
    /opt/freshvps/runtime/telegram/notify.sh "FreshVPS: operator bot online on $(hostname)" || true
  else
    info "Telegram secrets incomplete — unit installed, not started"
  fi
}

module_telegram_uninstall() {
  systemctl disable --now freshvps-telegram-bot 2>/dev/null || true
  rm -f /etc/systemd/system/freshvps-telegram-bot.service
  systemctl daemon-reload 2>/dev/null || true
}
