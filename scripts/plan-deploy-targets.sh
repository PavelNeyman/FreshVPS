#!/usr/bin/env bash
# Deploy all targets from a file (see configs/targets.example)
# Usage: bash scripts/plan-deploy-targets.sh [targets_file]
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FILE="${1:-${HOME}/.freshvps/targets}"
SSH_OPTS=(-o StrictHostKeyChecking=accept-new -o BatchMode=yes -o ConnectTimeout=15)

[[ -f "${FILE}" ]] || { echo "No targets file: ${FILE}"; echo "Copy configs/targets.example"; exit 1; }

ok=0
fail=0
while IFS= read -r line || [[ -n "${line}" ]]; do
  [[ -z "${line}" || "${line}" =~ ^# ]] && continue
  IFS='|' read -r kind host conf <<<"${line}"
  kind="$(echo "${kind}" | xargs)"
  host="$(echo "${host}" | xargs)"
  conf="$(echo "${conf:-}" | xargs)"
  echo "=== ${kind} → ${host} ==="
  set +e
  case "${kind}" in
    vps)
      ssh "${SSH_OPTS[@]}" "${host}" 'curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh | sudo bash -s -- --upgrade'
      rc=$?
      ;;
    openwrt)
      if [[ -z "${conf}" || ! -f "${conf}" ]]; then
        echo "FAIL openwrt needs site.conf path"; fail=$((fail+1)); continue
      fi
      scp "${SSH_OPTS[@]}" -r "${ROOT}/openwrt" "${host}:/root/freshvps-openwrt"
      scp "${SSH_OPTS[@]}" "${conf}" "${host}:/root/freshvps-openwrt/site.conf"
      ssh "${SSH_OPTS[@]}" "${host}" 'cd /root/freshvps-openwrt && sh install-openwrt.sh'
      rc=$?
      ;;
    *)
      echo "unknown kind: ${kind}"; rc=1
      ;;
  esac
  set -e
  if [[ "${rc}" -eq 0 ]]; then
    echo "OK ${kind} ${host}"
    ok=$((ok+1))
  else
    echo "FAIL ${kind} ${host} (rc=${rc}) — check SSH keys / BatchMode"
    fail=$((fail+1))
  fi
done <"${FILE}"

echo "Summary: ok=${ok} fail=${fail}"
[[ "${fail}" -eq 0 ]]
