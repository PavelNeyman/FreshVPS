#!/usr/bin/env bash
# FreshVPS VPS diagnostics — download to temp, run, summarize, wipe traces
# Usage:
#   sudo bash scripts/vps-tests.sh              # interactive menu
#   sudo bash scripts/vps-tests.sh --all        # all tests (long)
#   sudo bash scripts/vps-tests.sh --list
#   sudo bash scripts/vps-tests.sh ipregion yabs sysbench
#
# shellcheck disable=SC2034
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RESULTS_TSV=""
WORKDIR=""
export DEBIAN_FRONTEND=noninteractive

cleanup() {
  if [[ -n "${WORKDIR:-}" && -d "${WORKDIR}" ]]; then
    rm -rf "${WORKDIR}"
  fi
}
trap cleanup EXIT

log() { printf '[%s] %s\n' "$(date -Iseconds)" "$*"; }

need_root_hint() {
  if [[ "${EUID}" -ne 0 ]]; then
    log "WARN: some tests work better as root (packages). Continuing as $(id -un)."
  fi
}

ensure_pkgs() {
  local pkgs=()
  command -v curl >/dev/null || pkgs+=(curl)
  command -v wget >/dev/null || pkgs+=(wget)
  command -v jq >/dev/null || pkgs+=(jq)
  if [[ ${#pkgs[@]} -gt 0 && "${EUID}" -eq 0 ]]; then
    apt-get update -qq && apt-get install -y -qq --no-install-recommends "${pkgs[@]}" || true
  fi
}

# name|description|default_on
TEST_CATALOG=(
  "ipregion|IP region (ipregion.vrnt.xyz)|1"
  "censor_geoblock|Censorcheck geoblock|1"
  "censor_dpi|Censorcheck DPI (RU servers)|0"
  "iperf_ru|Russian iPerf3 servers (itdoginfo)|1"
  "yabs|YABS (disk/net, IPv4)|1"
  "ip_check_place|IP.Check.Place blocks abroad|1"
  "bench|bench.sh (params + foreign ISP)|0"
  "ipquality|IPQuality Check.Place -EI|0"
  "sysbench|sysbench cpu 1 thread|1"
)

record() {
  local id="$1" status="$2" note="$3"
  printf '%s\t%s\t%s\n' "${id}" "${status}" "${note}" >>"${RESULTS_TSV}"
}

run_captured() {
  local id="$1"
  shift
  local logf="${WORKDIR}/${id}.log"
  local rc=0
  set +e
  ("$@") >"${logf}" 2>&1
  rc=$?
  set -e
  if [[ "${rc}" -eq 0 ]]; then
    record "${id}" "OK" "log ${logf}"
  else
    record "${id}" "FAIL(${rc})" "see ${logf}"
  fi
  # show tail for operator
  tail -n 30 "${logf}" || true
  return 0
}

test_ipregion() {
  log "=== ipregion ==="
  run_captured ipregion bash -c 'wget -qO- https://ipregion.vrnt.xyz | bash'
}

test_censor_geoblock() {
  log "=== censorcheck geoblock ==="
  run_captured censor_geoblock bash -c 'wget -qO- https://github.com/vernette/censorcheck/raw/master/censorcheck.sh | bash -s -- --mode geoblock'
}

test_censor_dpi() {
  log "=== censorcheck dpi ==="
  run_captured censor_dpi bash -c 'wget -qO- https://github.com/vernette/censorcheck/raw/master/censorcheck.sh | bash -s -- --mode dpi'
}

test_iperf_ru() {
  log "=== russian iperf3 ==="
  run_captured iperf_ru bash -c 'wget -qO- https://github.com/itdoginfo/russian-iperf3-servers/raw/main/speedtest.sh | bash'
}

test_yabs() {
  log "=== yabs ==="
  run_captured yabs bash -c 'curl -sL yabs.sh | bash -s -- -4'
}

test_ip_check_place() {
  log "=== IP.Check.Place ==="
  run_captured ip_check_place bash -c 'curl -Ls IP.Check.Place | bash -s -- -l en'
}

test_bench() {
  log "=== bench.sh ==="
  run_captured bench bash -c 'wget -qO- bench.sh | bash'
}

test_ipquality() {
  log "=== IPQuality ==="
  run_captured ipquality bash -c 'curl -Ls https://Check.Place | bash -s -- -EI'
}

test_sysbench() {
  log "=== sysbench cpu ==="
  if ! command -v sysbench >/dev/null 2>&1; then
    if [[ "${EUID}" -eq 0 ]]; then
      apt-get install -y -qq sysbench || true
    fi
  fi
  if command -v sysbench >/dev/null 2>&1; then
    run_captured sysbench sysbench cpu run --threads=1
  else
    record sysbench SKIP "sysbench not installed"
    log "SKIP sysbench"
  fi
}

run_one() {
  case "$1" in
    ipregion) test_ipregion ;;
    censor_geoblock) test_censor_geoblock ;;
    censor_dpi) test_censor_dpi ;;
    iperf_ru) test_iperf_ru ;;
    yabs) test_yabs ;;
    ip_check_place) test_ip_check_place ;;
    bench) test_bench ;;
    ipquality) test_ipquality ;;
    sysbench) test_sysbench ;;
    *) log "Unknown test: $1"; record "$1" SKIP "unknown" ;;
  esac
}

print_summary() {
  echo
  echo "============================================================"
  echo " FreshVPS VPS tests — summary"
  echo "============================================================"
  printf '%-18s %-12s %s\n' "TEST" "STATUS" "NOTE"
  printf '%-18s %-12s %s\n' "----" "------" "----"
  local line id st note
  while IFS=$'\t' read -r id st note; do
    [[ -z "${id}" ]] && continue
    printf '%-18s %-12s %s\n' "${id}" "${st}" "${note}"
  done <"${RESULTS_TSV}"
  echo "============================================================"
  echo "Logs were under ${WORKDIR} and are removed on exit."
}

list_tests() {
  local row id desc def
  for row in "${TEST_CATALOG[@]}"; do
    IFS='|' read -r id desc def <<<"${row}"
    printf '%-18s [%s] %s\n' "${id}" "$([ "${def}" = 1 ] && echo on || echo off)" "${desc}"
  done
}

interactive_menu() {
  local selected=()
  if command -v whiptail >/dev/null 2>&1; then
    local args=()
    local row id desc def onoff
    for row in "${TEST_CATALOG[@]}"; do
      IFS='|' read -r id desc def <<<"${row}"
      onoff=OFF
      [[ "${def}" == "1" ]] && onoff=ON
      args+=("${id}" "${desc}" "${onoff}")
    done
    local out
    out="$(whiptail --title "FreshVPS VPS tests" --checklist \
      "Select tests (space to toggle). Long tests may take many minutes." 20 78 12 \
      "${args[@]}" 3>&1 1>&2 2>&3)" || true
    # shellcheck disable=SC2206
    selected=( ${out//\"/} )
  else
    echo "Available tests:"
    list_tests
    echo
    read -r -p "Enter test ids (space-separated) or 'default': " line
    if [[ "${line}" == "default" || -z "${line}" ]]; then
      local row id desc def
      for row in "${TEST_CATALOG[@]}"; do
        IFS='|' read -r id desc def <<<"${row}"
        [[ "${def}" == "1" ]] && selected+=("${id}")
      done
    else
      # shellcheck disable=SC2206
      selected=( ${line} )
    fi
  fi
  if [[ ${#selected[@]} -eq 0 ]]; then
    log "Nothing selected."
    exit 0
  fi
  local t
  for t in "${selected[@]}"; do
    run_one "${t}"
  done
}

main() {
  need_root_hint
  WORKDIR="$(mktemp -d /tmp/freshvps-tests.XXXXXX)"
  RESULTS_TSV="${WORKDIR}/results.tsv"
  : >"${RESULTS_TSV}"
  chmod 700 "${WORKDIR}"
  ensure_pkgs

  if [[ $# -eq 0 ]]; then
    interactive_menu
  elif [[ "$1" == "--list" ]]; then
    list_tests
    exit 0
  elif [[ "$1" == "--all" ]]; then
    local row id desc def
    for row in "${TEST_CATALOG[@]}"; do
      IFS='|' read -r id desc def <<<"${row}"
      run_one "${id}"
    done
  elif [[ "$1" == "--default" ]]; then
    local row id desc def
    for row in "${TEST_CATALOG[@]}"; do
      IFS='|' read -r id desc def <<<"${row}"
      [[ "${def}" == "1" ]] && run_one "${id}"
    done
  else
    local t
    for t in "$@"; do
      run_one "${t}"
    done
  fi

  print_summary
  # optional copy summary only to stdout already; wipe logs via trap
}

main "$@"
