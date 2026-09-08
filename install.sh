#!/usr/bin/env bash
# FreshVPS installer — ready-to-use; safe re-run / upgrade / VPS tests / prepare
set -euo pipefail

FRESHVPS_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export FRESHVPS_ROOT
source "${FRESHVPS_ROOT}/lib/common.sh"
source "${FRESHVPS_ROOT}/lib/tui.sh"
source "${FRESHVPS_ROOT}/lib/idempotent.sh"
source "${FRESHVPS_ROOT}/lib/install-conf.sh"
source "${FRESHVPS_ROOT}/lib/prepare.sh"

VERSION="$(cat "${FRESHVPS_ROOT}/VERSION" 2>/dev/null || echo 0.0.0)"

ENABLE_HARDENING=1 ENABLE_SINGBOX=1 ENABLE_BLOCKY=1 ENABLE_VPN_USERS=1
ENABLE_VPN_API=1 ENABLE_OPENSOHO=1 ENABLE_KUMA=0 ENABLE_BESZEL=0 ENABLE_METRICS=1
ENABLE_TELEGRAM=1 ENABLE_BACKUP=1 ENABLE_LAMPAC=0 ENABLE_METRICS=1

PUBLIC_IP="" SSH_PORT=22
SINGBOX_VLESS_PORT=443 SINGBOX_HY2_PORT=8443
BLOCKY_DNS_PORT=53 OPENSOHO_HTTP_PORT=8090 KUMA_PORT=3001 BESZEL_PORT=8091
VPN_API_PORT=8787 VPN_API_BIND=127.0.0.1
TELEGRAM_BOT_TOKEN="" TELEGRAM_ADMIN_ID="" BACKUP_REPO=""
CONFIG_FILE="" NONINTERACTIVE=0 ROLE=vps WITH_MIKROTIK=0 DO_UPGRADE=0 ACTION=""
INSTALL_AFTER_PREPARE=0
export NONINTERACTIVE

usage() {
  cat <<EOF
FreshVPS ${VERSION}
  --config FILE | --non-interactive | --upgrade | --force | --force-module NAME
  --role vps|edge-client|openwrt | --with-mikrotik | --tests | --prepare
EOF
}

load_config() { [[ -f "$1" ]] || die "Config not found: $1"; set -a; source "$1"; set +a; }

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --config) CONFIG_FILE="$2"; shift 2 ;;
      --non-interactive) NONINTERACTIVE=1; export NONINTERACTIVE; shift ;;
      --upgrade) DO_UPGRADE=1; FRESHVPS_UPGRADE=1; export FRESHVPS_UPGRADE; ACTION=upgrade; shift ;;
      --force) FRESHVPS_FORCE=1; export FRESHVPS_FORCE; shift ;;
      --force-module) FRESHVPS_FORCE_MODULES="${FRESHVPS_FORCE_MODULES:+${FRESHVPS_FORCE_MODULES},}$2"; export FRESHVPS_FORCE_MODULES; shift 2 ;;
      --role) ROLE="$2"; shift 2 ;;
      --with-mikrotik) WITH_MIKROTIK=1; shift ;;
      --tests) ACTION=tests; shift ;;
      --prepare) ACTION=prepare; shift ;;
      -h|--help) usage; exit 0 ;;
      *) die "Unknown option: $1" ;;
    esac
  done
}

prompt_yes_no() { tui_yesno "$@"; }
prompt_value() { tui_input "$@"; }

pick_action_menu() {
  [[ -n "${ACTION}" ]] && return 0
  [[ "${NONINTERACTIVE}" -eq 1 ]] && { ACTION=install; return 0; }
  ACTION="$(tui_menu "FreshVPS ${VERSION}" \
    install "Install / configure services" \
    prepare "Prepare VPS (apt/docker/golang/bot…)" \
    rebuild-bot "Fetch/rebuild Telegram bot only" \
    upgrade "Upgrade (idempotent)" \
    tests "VPS network tests" \
    quit "Exit")" || ACTION=quit
}

# Checklist for product modules → ENABLE_*
pick_components_checklist() {
  local result item
  result="$(tui_checklist "Components to install" \
    hardening "SSH hardening / BBR / UFW base" 1 \
    singbox "sing-box VLESS+Reality + HY2" 1 \
    blocky "Blocky DNS" 1 \
    vpn_users "Multi-user VPN CLI" 1 \
    vpn_api "Admin API (session auth)" 1 \
    opensoho "OpenSOHO (needs Docker)" 1 \
    kuma "Uptime Kuma (optional Docker)" 0 \
    beszel "Beszel (optional Docker)" 0 \
    telegram "Telegram operator bot" 1 \
    backup "restic backups" 1 \
    metrics "Built-in metrics (JSONL, no Prometheus)" 1 \
    lampac "Lampac (optional, needs Docker)" 0 \
    )" || true

  ENABLE_HARDENING=0 ENABLE_SINGBOX=0 ENABLE_BLOCKY=0 ENABLE_VPN_USERS=0
  ENABLE_VPN_API=0 ENABLE_OPENSOHO=0 ENABLE_KUMA=0 ENABLE_BESZEL=0
  ENABLE_TELEGRAM=0 ENABLE_BACKUP=0 ENABLE_LAMPAC=0 ENABLE_METRICS=0

  for item in ${result}; do
    case "${item}" in
      hardening) ENABLE_HARDENING=1 ;;
      singbox) ENABLE_SINGBOX=1 ;;
      blocky) ENABLE_BLOCKY=1 ;;
      vpn_users) ENABLE_VPN_USERS=1 ;;
      vpn_api) ENABLE_VPN_API=1 ;;
      opensoho) ENABLE_OPENSOHO=1 ;;
      kuma) ENABLE_KUMA=1 ;;
      beszel) ENABLE_BESZEL=1 ;;
      telegram) ENABLE_TELEGRAM=1 ;;
      backup) ENABLE_BACKUP=1 ;;
      metrics) ENABLE_METRICS=1 ;;
      lampac) ENABLE_LAMPAC=1 ;;
    esac
  done
}

collect_settings() {
  if [[ "${DO_UPGRADE}" -eq 1 ]]; then load_install_conf_if_present; fi
  if [[ -z "${CONFIG_FILE}" && "${DO_UPGRADE}" -eq 1 && -f /root/freshvps.conf ]]; then
    CONFIG_FILE=/root/freshvps.conf
  fi
  [[ -n "${CONFIG_FILE}" ]] && load_config "${CONFIG_FILE}"

  if [[ -z "${PUBLIC_IP}" && -f "${FRESHVPS_ETC}/public_ip" ]]; then
    PUBLIC_IP="$(cat "${FRESHVPS_ETC}/public_ip")"
  fi
  PUBLIC_IP="${PUBLIC_IP:-$(detect_public_ip)}"
  if [[ ! ( "${DO_UPGRADE}" -eq 1 && -n "${PUBLIC_IP}" && "${NONINTERACTIVE}" -eq 1 ) ]]; then
    PUBLIC_IP="$(prompt_value "Public IPv4 of this VPS" "${PUBLIC_IP}")"
  fi
  [[ -n "${PUBLIC_IP}" ]] || die "PUBLIC_IP required"

  if [[ "${NONINTERACTIVE}" -eq 0 && "${DO_UPGRADE}" -eq 0 ]]; then
    pick_components_checklist
    local flow
    flow="$(tui_menu "How to proceed" \
      install_now "Install selected components now" \
      prepare_then "Prepare deps for selection, then install" \
      prepare_only "Only prepare deps (no install)")" || flow=install_now
    case "${flow}" in
      prepare_only)
        ACTION=prepare
        prepare_derive_from_components
        # still allow manual tweak
        if prompt_yes_no "Review/adjust prepare checklist?" n; then
          prepare_pick_checklist
          # re-apply component hints for docker/golang if user cleared them wrongly
          prepare_derive_from_components
        fi
        ;;
      prepare_then)
        INSTALL_AFTER_PREPARE=1
        prepare_derive_from_components
        if prompt_yes_no "Review/adjust prepare checklist?" n; then
          prepare_pick_checklist
          prepare_derive_from_components
        fi
        ;;
      *) ;;
    esac
  fi

  if [[ "${ENABLE_SINGBOX}" -eq 1 && "${ENABLE_VPN_USERS}" -eq 0 ]]; then
    warn "sing-box without vpn-users → empty users (VPN will accept nobody)"
  fi

  if [[ "${ENABLE_TELEGRAM}" -eq 1 && "${NONINTERACTIVE}" -eq 0 && "${DO_UPGRADE}" -eq 0 && "${ACTION}" != "prepare" ]]; then
    TELEGRAM_BOT_TOKEN="$(prompt_value "Telegram bot token" "${TELEGRAM_BOT_TOKEN}")"
    TELEGRAM_ADMIN_ID="$(prompt_value "Telegram admin chat id" "${TELEGRAM_ADMIN_ID}")"
  fi
}

write_ready_summary() {
  local f="${FRESHVPS_ETC}/READY.txt"
  {
    echo "FreshVPS ${VERSION} READY — $(date -Iseconds)"
    echo "Host: $(hostname)  IP: ${PUBLIC_IP}"
    [[ -f /etc/freshvps/singbox-config-variant ]] && echo "sing-box variant: $(cat /etc/freshvps/singbox-config-variant)"
    echo
    echo "NOTE: HY2 client links use insecure=1 (self-signed cert). Prefer VLESS+Reality when possible."
    echo
    if [[ -f /etc/freshvps/clients/operator/subscription.txt ]]; then
      cat /etc/freshvps/clients/operator/subscription.txt
    fi
    echo
    echo "CLI: freshvps-vpn | freshvps-doctor | freshvps-smoke | freshvps-tests"
  } >"${f}"
  chmod 600 "${f}"
  cat "${f}"
}

run_tests_action() {
  source "${FRESHVPS_ROOT}/modules/vps-tests.sh"
  module_vps_tests_install
  bash /opt/freshvps/scripts/vps-tests.sh
}

run_install_modules() {
  pkg_install ca-certificates curl wget openssl jq tar gzip coreutils
  mkdir -p /opt/freshvps/compose
  [[ -f "${FRESHVPS_ROOT}/compose/panels.yml" ]] && install -m 644 "${FRESHVPS_ROOT}/compose/panels.yml" /opt/freshvps/compose/panels.yml

  run_module_idempotent host-tools
  [[ "${ENABLE_HARDENING}" -eq 1 ]] && run_module_idempotent hardening
  [[ "${ENABLE_SINGBOX}" -eq 1 ]] && run_module_idempotent sing-box
  [[ "${ENABLE_BLOCKY}" -eq 1 ]] && run_module_idempotent blocky
  [[ "${ENABLE_VPN_USERS}" -eq 1 && "${ENABLE_SINGBOX}" -eq 1 ]] && run_module_idempotent vpn-users
  [[ "${ENABLE_VPN_API}" -eq 1 ]] && run_module_idempotent vpn-api
  [[ "${ENABLE_OPENSOHO}" -eq 1 ]] && run_module_idempotent opensoho
  [[ "${ENABLE_KUMA}" -eq 1 ]] && run_module_idempotent kuma
  [[ "${ENABLE_BESZEL}" -eq 1 ]] && run_module_idempotent beszel
  [[ "${ENABLE_LAMPAC}" -eq 1 ]] && run_module_idempotent lampac
  [[ "${ENABLE_TELEGRAM}" -eq 1 ]] && run_module_idempotent telegram
  [[ "${ENABLE_BACKUP}" -eq 1 ]] && run_module_idempotent backup
  [[ "${ENABLE_METRICS:-1}" -eq 1 ]] && run_module_idempotent metrics
  run_module_idempotent vps-tests
  # shellcheck source=/dev/null
  source "${FRESHVPS_ROOT}/modules/host-tools.sh"
  module_host_tools_install
  mark_done host-tools

  write_install_conf
  write_installed_version "${VERSION}"
  write_ready_summary
  if [[ -x /opt/freshvps/scripts/smoke-host.sh ]]; then
    info "Running freshvps-smoke…"
    bash /opt/freshvps/scripts/smoke-host.sh || warn "smoke reported failures — run freshvps-doctor"
  fi
  info "Finished — ${VERSION}"
}

main() {
  parse_args "$@"
  if [[ "${ROLE}" == "openwrt" ]]; then exec bash "${FRESHVPS_ROOT}/install-openwrt.sh"; fi
  if [[ "${ROLE}" == "edge-client" ]]; then
    extra=(); [[ -n "${CONFIG_FILE}" ]] && extra+=(--config "${CONFIG_FILE}")
    [[ "${WITH_MIKROTIK}" -eq 1 ]] && extra+=(--with-mikrotik)
    exec bash "${FRESHVPS_ROOT}/install-edge.sh" "${extra[@]}"
  fi

  pick_action_menu
  [[ "${ACTION}" == "quit" ]] && exit 0
  if [[ "${ACTION}" == "tests" ]]; then require_root; run_tests_action; exit 0; fi
  if [[ "${ACTION}" == "rebuild-bot" ]]; then
    require_root; require_debian; ensure_dirs
    source "${FRESHVPS_ROOT}/lib/prepare.sh"
    prepare_pick_tg_bot_mode
    [[ "${PREPARE_TG_BOT}" -eq 1 ]] && prepare_run_tg_bot
    if [[ -x /opt/freshvps/bin/freshvps-tg ]]; then
      systemctl restart freshvps-telegram-bot 2>/dev/null || true
      info "Bot binary ready: /opt/freshvps/bin/freshvps-tg"
    fi
    exit 0
  fi
  if [[ "${ACTION}" == "upgrade" ]]; then DO_UPGRADE=1; FRESHVPS_UPGRADE=1; export FRESHVPS_UPGRADE; fi

  require_root; require_debian; ensure_dirs; touch "${FRESHVPS_LOG}"
  [[ "${DO_UPGRADE}" -eq 1 ]] && { NONINTERACTIVE="${NONINTERACTIVE:-1}"; export NONINTERACTIVE; }

  # Standalone prepare from main menu
  if [[ "${ACTION}" == "prepare" ]]; then
    info "FreshVPS ${VERSION} action=prepare"
    if [[ "${NONINTERACTIVE}" -eq 0 ]]; then
      prepare_pick_checklist
    else
      PREPARE_APT_UPDATE=1 PREPARE_BASE_TOOLS=1
      # optional via env PREPARE_DOCKER=1 etc.
    fi
    prepare_run_selected
    exit 0
  fi

  info "FreshVPS ${VERSION} action=${ACTION:-install}"
  collect_settings

  # collect_settings may switch ACTION to prepare
  if [[ "${ACTION}" == "prepare" && "${INSTALL_AFTER_PREPARE}" -eq 0 ]]; then
    prepare_run_selected
    info "Prepare finished — run install when ready"
    exit 0
  fi

  if [[ "${INSTALL_AFTER_PREPARE}" -eq 1 ]]; then
    prepare_run_selected
  fi

  export PUBLIC_IP SSH_PORT SINGBOX_VLESS_PORT SINGBOX_HY2_PORT SINGBOX_REALITY_SNI
  export BLOCKY_DNS_PORT OPENSOHO_HTTP_PORT KUMA_PORT BESZEL_PORT
  export VPN_API_PORT VPN_API_BIND TELEGRAM_BOT_TOKEN TELEGRAM_ADMIN_ID BACKUP_REPO ENABLE_LAMPAC
  printf '%s\n' "${PUBLIC_IP}" >"${FRESHVPS_ETC}/public_ip"

  run_install_modules

  if [[ "${NONINTERACTIVE}" -eq 0 ]]; then
    prompt_yes_no "Run VPS network tests now?" n && run_tests_action || true
  fi
}

main "$@"
