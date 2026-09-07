#!/usr/bin/env bash
# FreshVPS bootstrap — no git required.
#
# On target VPS (Debian):
#   curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh | sudo bash
#
# With options:
#   curl -fsSL .../bootstrap.sh | sudo bash -s -- --non-interactive --config /root/freshvps.conf
#
# On macOS (control plane / plan mode — does not install VPN on the Mac):
#   curl -fsSL .../bootstrap.sh | bash -s -- --mode plan
#
# Env:
#   FRESHVPS_REF=main|v0.3.1|commit   FRESHVPS_TARBALL_URL=...   FRESHVPS_DIR=...
#
set -euo pipefail

REF="${FRESHVPS_REF:-main}"
REPO="PavelNeyman/FreshVPS"
TARBALL_URL="${FRESHVPS_TARBALL_URL:-https://codeload.github.com/${REPO}/tar.gz/refs/heads/${REF}}"
# tags: https://codeload.github.com/.../tar.gz/refs/tags/v0.3.1
MODE=""
EXTRA=()

usage() {
  cat <<'EOF'
FreshVPS bootstrap (git optional)

  --mode auto|vps|edge-client|openwrt|plan   detect or force
  --ref REF           git branch/tag name for tarball (default main)
  --dir DIR           unpack directory (default mktemp)
  --keep              do not delete unpack dir after run
  --non-interactive   pass through to install.sh
  --config FILE       pass through
  --with-mikrotik     edge only
  -h|--help
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --mode) MODE="$2"; shift 2 ;;
    --ref) REF="$2"; TARBALL_URL="https://codeload.github.com/${REPO}/tar.gz/refs/heads/${REF}"
           # if looks like version tag
           if [[ "${REF}" == v* ]]; then
             TARBALL_URL="https://codeload.github.com/${REPO}/tar.gz/refs/tags/${REF}"
           fi
           shift 2 ;;
    --dir) FRESHVPS_DIR="$2"; shift 2 ;;
    --keep) KEEP=1; shift ;;
    --non-interactive) EXTRA+=(--non-interactive); shift ;;
    --config) EXTRA+=(--config "$2"); shift 2 ;;
    --with-mikrotik) EXTRA+=(--with-mikrotik); shift ;;
    -h|--help) usage; exit 0 ;;
    *) EXTRA+=("$1"); shift ;;
  esac
done

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || { echo "Missing command: $1" >&2; exit 1; }
}

detect_mode() {
  if [[ -n "${MODE}" && "${MODE}" != "auto" ]]; then
    echo "${MODE}"
    return
  fi
  local u; u="$(uname -s)"
  case "${u}" in
    Darwin) echo plan; return ;;
    Linux)
      if [[ -f /etc/openwrt_release ]] || grep -qi openwrt /etc/os-release 2>/dev/null; then
        echo openwrt; return
      fi
      if [[ -f /etc/debian_version ]]; then
        # heuristic: if already a small edge hostname env, still default vps
        echo vps; return
      fi
      echo vps; return
      ;;
    *)
      echo "Unsupported OS: ${u}" >&2
      exit 1
      ;;
  esac
}

fetch_tree() {
  local dest="$1"
  need_cmd curl
  need_cmd tar
  mkdir -p "${dest}"
  echo "[bootstrap] downloading ${TARBALL_URL}"
  local tmp; tmp="$(mktemp)"
  if ! curl -fsSL "${TARBALL_URL}" -o "${tmp}"; then
    # fallback API zip needs unzip; try git if present
    if command -v git >/dev/null 2>&1; then
      echo "[bootstrap] tarball failed — git clone"
      git clone --depth 1 --branch "${REF}" "https://github.com/${REPO}.git" "${dest}/repo"
      echo "${dest}/repo"
      rm -f "${tmp}"
      return
    fi
    echo "[bootstrap] download failed" >&2
    exit 1
  fi
  tar -xzf "${tmp}" -C "${dest}"
  rm -f "${tmp}"
  # GitHub tarball extracts to Repo-ref/
  local top
  top="$(find "${dest}" -mindepth 1 -maxdepth 1 -type d | head -1)"
  echo "${top}"
}

main() {
  local mode work root
  mode="$(detect_mode)"
  echo "[bootstrap] mode=${mode} uname=$(uname -s)"

  if [[ "${mode}" == "openwrt" ]]; then
    echo "[bootstrap] OpenWrt: copy openwrt/ tree to the router and run install-openwrt.sh there."
    echo "  This host looks like OpenWrt — fetching tree for local use."
  fi

  if [[ "${mode}" == "plan" ]]; then
    work="${FRESHVPS_DIR:-$(mktemp -d /tmp/freshvps-plan.XXXXXX)}"
    root="$(fetch_tree "${work}")"
    echo "[bootstrap] plan mode on $(uname -s)"
    echo "  Unpacked: ${root}"
    echo "  Next: edit configs, generate site.conf, deploy via SSH keys (see docs/BOOTSTRAP.md)."
    if [[ -f "${root}/lib/tui.sh" ]]; then
      # shellcheck source=/dev/null
      source "${root}/lib/tui.sh"
      tui_msg "FreshVPS plan mode: tree is at ${root}. Use SSH keys to push configs to VPS/OpenWrt (docs/BOOTSTRAP.md)."
    fi
    if [[ "${KEEP:-0}" -ne 1 && -z "${FRESHVPS_DIR:-}" ]]; then
      echo "[bootstrap] keeping ${work} for plan (set FRESHVPS_DIR or --keep)"
    fi
    exit 0
  fi

  if [[ "$(id -u)" -ne 0 && "${mode}" != "plan" ]]; then
    echo "[bootstrap] re-exec with sudo"
    exec sudo -E bash "$0" --mode "${mode}" ${KEEP:+--keep} "${EXTRA[@]+"${EXTRA[@]}"}"
  fi

  work="${FRESHVPS_DIR:-$(mktemp -d /tmp/freshvps.XXXXXX)}"
  root="$(fetch_tree "${work}")"
  cd "${root}"

  case "${mode}" in
    vps)
      bash install.sh --role vps "${EXTRA[@]+"${EXTRA[@]}"}"
      ;;
    edge-client)
      bash install.sh --role edge-client "${EXTRA[@]+"${EXTRA[@]}"}"
      ;;
    openwrt)
      bash install-openwrt.sh || sh openwrt/install-openwrt.sh
      ;;
    *)
      echo "Unknown mode: ${mode}" >&2
      exit 1
      ;;
  esac

  if [[ "${KEEP:-0}" -ne 1 && -z "${FRESHVPS_DIR:-}" ]]; then
    rm -rf "${work}"
  fi
}

main
