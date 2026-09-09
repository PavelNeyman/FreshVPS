#!/usr/bin/env bash
# Netductor bootstrap — no git required (legacy name: FreshVPS).
#
#   curl -fsSL .../bootstrap.sh -o /tmp/nd.sh && sudo bash /tmp/nd.sh
#
set -euo pipefail

REF="${NETDUCTOR_REF:-${FRESHVPS_REF:-main}}"
REPO="${NETDUCTOR_REPO:-PavelNeyman/netductor}"
TARBALL_URL="${NETDUCTOR_TARBALL_URL:-${FRESHVPS_TARBALL_URL:-https://codeload.github.com/${REPO}/tar.gz/refs/heads/${REF}}}"
MODE=""
EXTRA=()
KEEP=0

usage() {
  cat <<'EOF'
Netductor bootstrap

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

fetch_tree() {
  local dest="$1" top tmp
  need_cmd curl; need_cmd tar
  mkdir -p "${dest}"
  echo "[bootstrap] downloading ${TARBALL_URL}" >&2
  tmp="$(mktemp)"
  if ! curl -fsSL "${TARBALL_URL}" -o "${tmp}"; then
    if command -v git >/dev/null 2>&1; then
      echo "[bootstrap] codeload failed — falling back to git clone" >&2
      git clone --depth 1 --branch "${REF}" "https://github.com/${REPO}.git" "${dest}/repo"
      rm -f "${tmp}"
      printf '%s\n' "${dest}/repo"
      return
    fi
    echo "[bootstrap] download failed" >&2; exit 1
  fi
  tar -xzf "${tmp}" -C "${dest}"
  rm -f "${tmp}"
  top="$(find "${dest}" -mindepth 1 -maxdepth 1 -type d | head -1)"
  if [[ -z "${top}" || ! -d "${top}" ]]; then
    echo "[bootstrap] empty or invalid tarball under ${dest}" >&2
    exit 1
  fi
  printf '%s\n' "${top}"
}

main() {
  local mode work root
  mode="$(detect_mode)"
  echo "[bootstrap] mode=${mode} repo=${REPO}"

  if [[ -f /var/lib/freshvps/installed_version ]]; then
    echo "[bootstrap] existing install: $(cat /var/lib/freshvps/installed_version)"
  fi

  if [[ "${mode}" == "plan" ]]; then
    work="${FRESHVPS_DIR:-$(mktemp -d /tmp/netductor-plan.XXXXXX)}"
    root="$(fetch_tree "${work}")"
    echo "[bootstrap] plan tree: ${root}"
    if [[ -f "${root}/scripts/plan.sh" ]]; then
      bash "${root}/scripts/plan.sh" "${root}"
    else
      echo "Tree at ${root} — see docs/"
    fi
    exit 0
  fi

  if [[ "$(id -u)" -ne 0 ]]; then
    echo "[bootstrap] need root for mode=${mode}"
    exit 1
  fi

  work="${FRESHVPS_DIR:-$(mktemp -d /tmp/netductor.XXXXXX)}"
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
