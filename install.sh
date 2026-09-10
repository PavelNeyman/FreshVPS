#!/usr/bin/env bash
# Deprecated entrypoint — Netductor install is Go.
# Prefer: netductor install
set -euo pipefail
if ! command -v netductor >/dev/null 2>&1; then
  echo "netductor not in PATH — install release binary first" >&2
  exit 1
fi
exec netductor install "$@"
