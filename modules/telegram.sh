#!/usr/bin/env bash
# Module: Telegram operator panel (Go bot — netductor-tg)
# shellcheck disable=SC2154
#
# Binary resolution:
#   1) FRESHVPS_TG_BIN / NETDUCTOR_TG_BIN
#   2) /opt/freshvps/bin/netductor-tg or freshvps-tg
#   3) dist/netductor-tg-linux-* or freshvps-tg-linux-*
#   4) GitHub release (PavelNeyman/netductor)
#   5) go build cmd/netductor-tg or cmd/freshvps-tg
#   6) bash fallback

_telegram_install_binary() {
  local dest=/opt/freshvps/bin/netductor-tg
  local legacy=/opt/freshvps/bin/freshvps-tg
  local arch src ver
  arch="$(arch_go)"
  mkdir -p /opt/freshvps/bin

  if [[ -n "${NETDUCTOR_TG_BIN:-${FRESHVPS_TG_BIN:-}}" && -f "${NETDUCTOR_TG_BIN:-${FRESHVPS_TG_BIN}}" ]]; then
    local pre="${NETDUCTOR_TG_BIN:-${FRESHVPS_TG_BIN}}"
    install -m 755 "${pre}" "${dest}"
    ln -sfn "${dest}" "${legacy}"
    info "Telegram bot: using prebuilt ${pre}"
    printf '%s\n' "${dest}"
    return 0
  fi

  if [[ -x "${dest}" ]]; then
    ln -sfn "${dest}" "${legacy}" 2>/dev/null || true
    info "Telegram bot: reusing ${dest}"
    printf '%s\n' "${dest}"
    return 0
  fi
  if [[ -x "${legacy}" ]]; then
    info "Telegram bot: reusing legacy ${legacy}"
    printf '%s\n' "${legacy}"
    return 0
  fi

  for src in \
    "${FRESHVPS_ROOT}/dist/netductor-tg-linux-${arch}" \
    "${FRESHVPS_ROOT}/dist/freshvps-tg-linux-${arch}"; do
    if [[ -f "${src}" ]]; then
      install -m 755 "${src}" "${dest}"
      ln -sfn "${dest}" "${legacy}"
      info "Telegram bot: installed from dist/ $(basename "${src}")"
      printf '%s\n' "${dest}"
      return 0
    fi
  done

  ver="$(tr -d '[:space:]' <"${FRESHVPS_ROOT}/VERSION" 2>/dev/null || echo "0.7.0-dev")"
  for name in "netductor-tg-linux-${arch}" "freshvps-tg-linux-${arch}"; do
    local url="https://github.com/PavelNeyman/netductor/releases/download/v${ver}/${name}"
    if curl -fsSL "${url}" -o "${dest}.tmp" 2>/dev/null; then
      chmod 755 "${dest}.tmp"
      mv "${dest}.tmp" "${dest}"
      ln -sfn "${dest}" "${legacy}"
      info "Telegram bot: downloaded ${name}"
      printf '%s\n' "${dest}"
      return 0
    fi
    rm -f "${dest}.tmp"
  done

  if command -v go >/dev/null 2>&1 || pkg_install golang-go 2>/dev/null; then
    local srcdir
    if [[ -d "${FRESHVPS_ROOT}/cmd/netductor-tg" ]]; then
      srcdir="${FRESHVPS_ROOT}/cmd/netductor-tg"
    else
      srcdir="${FRESHVPS_ROOT}/cmd/freshvps-tg"
    fi
    info "Telegram bot: building ${srcdir}"
    ( cd "${srcdir}" && go build -o "${dest}" -trimpath -ldflags='-s -w' . )
    chmod 755 "${dest}"
    ln -sfn "${dest}" "${legacy}"
    printf '%s\n' "${dest}"
    return 0
  fi

  if [[ -f "${FRESHVPS_ROOT}/runtime/telegram/bot.sh" ]]; then
    warn "Telegram bot: bash fallback"
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

  # Primary unit (new name) + legacy alias
  for unit in netductor-telegram-bot freshvps-telegram-bot; do
    cat >"/etc/systemd/system/${unit}.service" <<EOF
[Unit]
Description=Netductor Telegram operator bot
After=network-online.target

[Service]
Type=simple
ExecStart=${bin}
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
