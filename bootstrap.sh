#!/usr/bin/env bash
# FreshVPS bootstrap — no git required.
#
# Prefer:
#   curl -fsSL .../bootstrap.sh -o /tmp/fv.sh && sudo bash /tmp/fv.sh
# Also works:
#   curl -fsSL .../bootstrap.sh | sudo bash
#
set -euo pipefail

REF="${FRESHVPS_REF:-main}"
REPO="PavelNeyman/FreshVPS"
TARBALL_URL="${FRESHVPS_TARBALL_URL:-https://codeload.github.com/${REPO}/tar.gz/refs/heads/${REF}}"
MODE=""
EXTRA=()
KEEP=0

usage() {
  cat <<'EOF'
FreshVPS bootstrap

  --mode auto|vps|edge-client|openwrt|plan
  --ref REF | --dir DIR | --keep
  --upgrade | --force | --force-module NAME
  --non-interactive | --config FILE
  --with-mikrotik
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --mode) MODE="$2"; shift 2 ;;
    --ref)
      REF="$2"
      if [[ "${REF}" == v* ]]; then
        TARBALL_URL="https://codeload.github.com/${REPO}/tar.gz/refs/tags/${REF}"
      else
        TARBALL_URL="https://codeload.github.com/${REPO}/tar.gz/refs/heads/${REF}"
      fi
      shift 2
      ;;
    --dir) FRESHVPS_DIR="$2"; shift 2 ;;
    --keep) KEEP=1; shift ;;
    --upgrade) EXTRA+=(--upgrade); shift ;;
    --force) EXTRA+=(--force); shift ;;
    --force-module) EXTRA+=(--force-module "$2"); shift 2 ;;
    --non-interactive) EXTRA+=(--non-interactive); shift ;;
    --config) EXTRA+=(--config "$2"); shift 2 ;;
    --with-mikrotik) EXTRA+=(--with-mikrotik); shift ;;
    -h|--help) usage; exit 0 ;;
    *) EXTRA+=("$1"); shift ;;
  esac
done

need_cmd() { command -v "$1" >/dev/null 2>&1 || { echo "Missing: $1" >&2; exit 1; }; }

detect_mode() {
  if [[ -n "${MODE}" && "${MODE}" != "auto" ]]; then echo "${MODE}"; return; fi
  case "$(uname -s)" in
    Darwin) echo plan ;;
    Linux)
      if [[ -f /etc/openwrt_release ]] || grep -qi openwrt /etc/os-release 2>/dev/null; then echo openwrt
      elif [[ -f /etc/debian_version ]]; then echo vps
      else
        echo "[bootstrap] Linux without Debian/OpenWrt — use --mode explicitly" >&2
        exit 1
      fi
      ;;
    *) echo "Unsupported OS" >&2; exit 1 ;;
  esac
}

# If we were started via curl|bash, $0 is not a real file — materialize once.
ensure_real_script() {
  local self="${BASH_SOURCE[0]:-$0}"
  if [[ -f "${self}" && -r "${self}" && "${self}" != "-" && "${self}" != "bash" ]]; then
    echo "${self}"
    return
  fi
  local copy
  copy="$(mktemp /tmp/freshvps-bootstrap.XXXXXX)"
  cat >"${copy}"
  chmod 700 "${copy}"
  echo "${copy}"
}

fetch_tree() {
  local dest="$1" top tmp
  need_cmd curl; need_cmd tar
  mkdir -p "${dest}"
  echo "[bootstrap] downloading ${TARBALL_URL}"
  tmp="$(mktemp)"
  if ! curl -fsSL "${TARBALL_URL}" -o "${tmp}"; then
    if command -v git >/dev/null 2>&1; then
      git clone --depth 1 --branch "${REF}" "https://github.com/${REPO}.git" "${dest}/repo"
      echo "${dest}/repo"; rm -f "${tmp}"; return
    fi
    echo "[bootstrap] download failed" >&2; exit 1
  fi
  tar -xzf "${tmp}" -C "${dest}"
  rm -f "${tmp}"
  top="$(find "${dest}" -mindepth 1 -maxdepth 1 -type d | head -1)"
  echo "${top}"
}

main() {
  local mode work root script_path
  mode="$(detect_mode)"
  echo "[bootstrap] mode=${mode}"

  if [[ -f /var/lib/freshvps/installed_version ]]; then
    echo "[bootstrap] existing install: $(cat /var/lib/freshvps/installed_version)"
  fi

  if [[ "${mode}" == "plan" ]]; then
    work="${FRESHVPS_DIR:-$(mktemp -d /tmp/freshvps-plan.XXXXXX)}"
    root="$(fetch_tree "${work}")"
    echo "[bootstrap] plan tree: ${root}"
    if [[ -f "${root}/scripts/plan.sh" ]]; then
      bash "${root}/scripts/plan.sh" "${root}"
    else
      echo "Tree at ${root} — see docs/BOOTSTRAP.md"
    fi
    exit 0
  fi

  if [[ "$(id -u)" -ne 0 ]]; then
    echo "[bootstrap] need root for mode=${mode}"
    echo "  curl -fsSL .../bootstrap.sh -o /tmp/fv.sh && sudo bash /tmp/fv.sh --mode ${mode} ..."
    exit 1
  fi

  work="${FRESHVPS_DIR:-$(mktemp -d /tmp/freshvps.XXXXXX)}"
  root="$(fetch_tree "${work}")"
  cd "${root}"

  case "${mode}" in
    vps) bash install.sh --role vps "${EXTRA[@]+"${EXTRA[@]}"}" ;;
    edge-client) bash install.sh --role edge-client "${EXTRA[@]+"${EXTRA[@]}"}" ;;
    openwrt) sh openwrt/install-openwrt.sh ;;
    *) echo "Unknown mode ${mode}" >&2; exit 1 ;;
  esac

  if [[ "${KEEP}" -ne 1 && -z "${FRESHVPS_DIR:-}" ]]; then
    rm -rf "${work}"
  fi
}

main
