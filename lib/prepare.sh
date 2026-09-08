#!/usr/bin/env bash
# VPS preparation helpers — install only what is needed.
# shellcheck disable=SC2154

PREPARE_APT_UPDATE=${PREPARE_APT_UPDATE:-0}
PREPARE_APT_UPGRADE=${PREPARE_APT_UPGRADE:-0}
PREPARE_BASE_TOOLS=${PREPARE_BASE_TOOLS:-0}
PREPARE_DOCKER=${PREPARE_DOCKER:-0}
PREPARE_GOLANG=${PREPARE_GOLANG:-0}
PREPARE_RESTIC=${PREPARE_RESTIC:-0}
PREPARE_DNSUTILS=${PREPARE_DNSUTILS:-0}
PREPARE_TG_BOT=${PREPARE_TG_BOT:-0}

prepare_derive_from_components() {
  PREPARE_APT_UPDATE=1
  PREPARE_BASE_TOOLS=1
  PREPARE_DNSUTILS=1
  if [[ "${ENABLE_OPENSOHO:-0}" -eq 1 || "${ENABLE_KUMA:-0}" -eq 1 \
     || "${ENABLE_BESZEL:-0}" -eq 1 || "${ENABLE_LAMPAC:-0}" -eq 1 ]]; then
    PREPARE_DOCKER=1
  fi
  if [[ "${ENABLE_TELEGRAM:-0}" -eq 1 ]]; then
    # Prefer prebuilt bot step over installing golang on VPS
    PREPARE_TG_BOT=1
    if ! prepare_tg_prebuilt_available && [[ "${PREPARE_TG_BOT}" -eq 0 ]]; then
      PREPARE_GOLANG=1
    fi
  fi
  if [[ "${ENABLE_BACKUP:-0}" -eq 1 ]]; then
    PREPARE_RESTIC=1
  fi
}

prepare_tg_prebuilt_available() {
  local arch bin
  arch="$(arch_go 2>/dev/null || echo amd64)"
  bin="${FRESHVPS_TG_BIN:-}"
  [[ -n "${bin}" && -x "${bin}" ]] && return 0
  [[ -x /opt/freshvps/bin/freshvps-tg ]] && return 0
  [[ -n "${FRESHVPS_ROOT:-}" && -x "${FRESHVPS_ROOT}/dist/freshvps-tg-linux-${arch}" ]] && return 0
  return 1
}

# Download or build bot into /opt/freshvps/bin/freshvps-tg (no systemd)
prepare_run_tg_bot() {
  local dest=/opt/freshvps/bin/freshvps-tg arch ver url tmp
  arch="$(arch_go)"
  mkdir -p /opt/freshvps/bin
  info "prepare: telegram bot binary (linux-${arch})"

  if [[ -n "${FRESHVPS_TG_BIN:-}" && -f "${FRESHVPS_TG_BIN}" ]]; then
    install -m 755 "${FRESHVPS_TG_BIN}" "${dest}"
    info "prepare: tg bot from FRESHVPS_TG_BIN"
    return 0
  fi
  if [[ -n "${FRESHVPS_ROOT:-}" && -f "${FRESHVPS_ROOT}/dist/freshvps-tg-linux-${arch}" ]]; then
    install -m 755 "${FRESHVPS_ROOT}/dist/freshvps-tg-linux-${arch}" "${dest}"
    info "prepare: tg bot from dist/"
    return 0
  fi

  ver="$(cat "${FRESHVPS_ROOT:-/}/VERSION" 2>/dev/null || echo "")"
  tmp="$(mktemp)"
  for url in \
    "https://github.com/PavelNeyman/FreshVPS/releases/download/v${ver}/freshvps-tg-linux-${arch}" \
    "https://github.com/PavelNeyman/FreshVPS/releases/latest/download/freshvps-tg-linux-${arch}"
  do
    [[ -z "${ver}" && "${url}" == *"/v/"* ]] && continue
    if curl -fsSL "${url}" -o "${tmp}" 2>/dev/null && [[ -s "${tmp}" ]]; then
      # reject HTML error pages
      if head -c 2 "${tmp}" | grep -q $'\x7fE'; then
        :
      elif file "${tmp}" 2>/dev/null | grep -qi elf; then
        :
      elif [[ $(wc -c <"${tmp}") -lt 100000 ]]; then
        continue
      fi
      install -m 755 "${tmp}" "${dest}"
      rm -f "${tmp}"
      info "prepare: tg bot downloaded (${url##*/})"
      return 0
    fi
  done
  rm -f "${tmp}"

  if command -v go >/dev/null 2>&1 || pkg_install golang-go 2>/dev/null; then
    if [[ -d "${FRESHVPS_ROOT}/cmd/freshvps-tg" ]]; then
      info "prepare: building tg bot with Go"
      ( cd "${FRESHVPS_ROOT}/cmd/freshvps-tg" && go build -o "${dest}" -trimpath -ldflags='-s -w' . )
      chmod 755 "${dest}"
      return 0
    fi
  fi
  warn "prepare: could not obtain tg bot binary — module will try again at install"
}

prepare_run_apt_update() {
  export DEBIAN_FRONTEND=noninteractive
  info "prepare: apt-get update"
  apt-get update -y
  mkdir -p "${FRESHVPS_STATE_DIR}"
  touch "${FRESHVPS_STATE_DIR}/apt-updated"
}

prepare_run_apt_upgrade() {
  export DEBIAN_FRONTEND=noninteractive
  info "prepare: apt-get upgrade (safe)"
  apt-get upgrade -y
}

prepare_run_base_tools() {
  info "prepare: base tools (curl jq tar openssl…)"
  pkg_install ca-certificates curl wget openssl jq tar gzip coreutils
}

prepare_run_dnsutils() {
  info "prepare: dnsutils"
  pkg_install dnsutils
}

prepare_run_restic() {
  info "prepare: restic"
  pkg_install restic
}

prepare_run_golang() {
  if command -v go >/dev/null 2>&1; then
    info "prepare: golang already present ($(go version 2>/dev/null || true))"
    return 0
  fi
  info "prepare: golang-go"
  pkg_install golang-go
}

prepare_run_docker() {
  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    info "prepare: docker already present"
    systemctl enable --now docker 2>/dev/null || true
    return 0
  fi
  info "prepare: Docker CE"
  pkg_install ca-certificates curl
  curl -fsSL https://get.docker.com | sh
  systemctl enable --now docker
  if ! docker compose version >/dev/null 2>&1; then
    pkg_install docker-compose-plugin 2>/dev/null || true
  fi
}

prepare_run_selected() {
  require_root
  require_debian
  ensure_dirs
  [[ "${PREPARE_APT_UPDATE}" -eq 1 ]] && prepare_run_apt_update
  [[ "${PREPARE_APT_UPGRADE}" -eq 1 ]] && prepare_run_apt_upgrade
  [[ "${PREPARE_BASE_TOOLS}" -eq 1 ]] && prepare_run_base_tools
  [[ "${PREPARE_DNSUTILS}" -eq 1 ]] && prepare_run_dnsutils
  [[ "${PREPARE_RESTIC}" -eq 1 ]] && prepare_run_restic
  [[ "${PREPARE_GOLANG}" -eq 1 ]] && prepare_run_golang
  [[ "${PREPARE_DOCKER}" -eq 1 ]] && prepare_run_docker
  [[ "${PREPARE_TG_BOT}" -eq 1 ]] && prepare_run_tg_bot
  info "prepare: done"
}

prepare_pick_checklist() {
  local result item
  PREPARE_APT_UPDATE=0 PREPARE_APT_UPGRADE=0 PREPARE_BASE_TOOLS=0
  PREPARE_DOCKER=0 PREPARE_GOLANG=0 PREPARE_RESTIC=0 PREPARE_DNSUTILS=0 PREPARE_TG_BOT=0

  if [[ "${NONINTERACTIVE:-0}" -eq 1 ]]; then
    PREPARE_APT_UPDATE=${PREPARE_APT_UPDATE:-1}
    PREPARE_BASE_TOOLS=${PREPARE_BASE_TOOLS:-1}
    return 0
  fi

  result="$(tui_checklist "VPS preparation — select steps" \
    apt_update "apt-get update" 1 \
    apt_upgrade "apt-get upgrade (optional, slower)" 0 \
    base_tools "curl jq tar openssl ca-certificates" 1 \
    dnsutils "dnsutils (dig)" 1 \
    docker "Docker CE + compose plugin" 0 \
    tg_bot "Telegram bot binary (download/build, no golang needed if prebuilt)" 1 \
    golang "golang-go (only if compiling bot yourself)" 0 \
    restic "restic binary" 0 \
    )" || true

  for item in ${result}; do
    case "${item}" in
      apt_update) PREPARE_APT_UPDATE=1 ;;
      apt_upgrade) PREPARE_APT_UPGRADE=1 ;;
      base_tools) PREPARE_BASE_TOOLS=1 ;;
      dnsutils) PREPARE_DNSUTILS=1 ;;
      docker) PREPARE_DOCKER=1 ;;
      tg_bot) PREPARE_TG_BOT=1 ;;
      golang) PREPARE_GOLANG=1 ;;
      restic) PREPARE_RESTIC=1 ;;
    esac
  done
}
