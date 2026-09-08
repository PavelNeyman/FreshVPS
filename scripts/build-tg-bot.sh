#!/usr/bin/env bash
# Build freshvps-tg for Linux on any host with Go (including macOS).
# Usage:
#   ./scripts/build-tg-bot.sh              # both amd64 + arm64 → dist/
#   ./scripts/build-tg-bot.sh amd64        # one arch
#   ./scripts/build-tg-bot.sh amd64 arm64
#
# Deploy to VPS (match VPS arch: uname -m → x86_64=amd64, aarch64=arm64):
#   scp dist/freshvps-tg-linux-amd64 root@VPS:/opt/freshvps/bin/freshvps-tg
#   # or before install:
#   scp dist/freshvps-tg-linux-amd64 root@VPS:/root/freshvps-tg
#   ssh root@VPS 'export FRESHVPS_TG_BIN=/root/freshvps-tg; bash install …'
#
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="${ROOT}/dist"
SRC="${ROOT}/cmd/freshvps-tg"

need_go() {
  command -v go >/dev/null 2>&1 || {
    echo "Go toolchain required (https://go.dev/dl/)" >&2
    exit 1
  }
}

build_one() {
  local arch="$1" dest
  dest="${OUT}/freshvps-tg-linux-${arch}"
  echo "Building linux/${arch} → ${dest}"
  mkdir -p "${OUT}"
  (
    cd "${SRC}"
    CGO_ENABLED=0 GOOS=linux GOARCH="${arch}" go build -trimpath -ldflags='-s -w' -o "${dest}" .
  )
  chmod 755 "${dest}"
  ls -la "${dest}"
}

main() {
  need_go
  local targets=("$@")
  if [[ ${#targets[@]} -eq 0 ]]; then
    targets=(amd64 arm64)
  fi
  local t
  for t in "${targets[@]}"; do
    case "${t}" in
      amd64|arm64|arm) build_one "${t}" ;;
      *) echo "Unknown arch: ${t} (use amd64|arm64)" >&2; exit 1 ;;
    esac
  done
  echo "Done. Optional: commit dist/ or scp to VPS (see header comments)."
}

main "$@"
