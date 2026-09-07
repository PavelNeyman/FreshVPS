#!/usr/bin/env bash
# Parse vless:// UUID@host:port?query#name into globals VLESS_*
# shellcheck disable=SC2034

vless_parse() {
  local url="$1"
  VLESS_UUID="" VLESS_HOST="" VLESS_PORT="" VLESS_SNI="" VLESS_PBK="" VLESS_SID="" VLESS_FP="" VLESS_FLOW="" VLESS_NAME=""
  [[ "${url}" == vless://* ]] || return 1
  local rest="${url#vless://}"
  VLESS_NAME=""
  if [[ "${rest}" == *#* ]]; then
    VLESS_NAME="${rest##*#}"
    rest="${rest%%#*}"
  fi
  VLESS_UUID="${rest%%@*}"
  rest="${rest#*@}"
  local hostport qs
  if [[ "${rest}" == *\?* ]]; then
    hostport="${rest%%\?*}"
    qs="${rest#*\?}"
  else
    hostport="${rest}"
    qs=""
  fi
  # IPv6 [addr]:port
  if [[ "${hostport}" == \[*\]* ]]; then
    VLESS_HOST="${hostport#\[}"
    VLESS_HOST="${VLESS_HOST%%\]*}"
    VLESS_PORT="${hostport##*:}"
  else
    VLESS_HOST="${hostport%%:*}"
    VLESS_PORT="${hostport##*:}"
  fi
  VLESS_SNI="$(printf '%s' "${qs}" | tr '&' '\n' | sed -n 's/^sni=//p' | head -1)"
  VLESS_PBK="$(printf '%s' "${qs}" | tr '&' '\n' | sed -n 's/^pbk=//p' | head -1)"
  VLESS_SID="$(printf '%s' "${qs}" | tr '&' '\n' | sed -n 's/^sid=//p' | head -1)"
  VLESS_FP="$(printf '%s' "${qs}" | tr '&' '\n' | sed -n 's/^fp=//p' | head -1)"
  VLESS_FLOW="$(printf '%s' "${qs}" | tr '&' '\n' | sed -n 's/^flow=//p' | head -1)"
  VLESS_SNI="${VLESS_SNI:-www.cloudflare.com}"
  VLESS_FP="${VLESS_FP:-chrome}"
  VLESS_FLOW="${VLESS_FLOW:-xtls-rprx-vision}"
  VLESS_PORT="${VLESS_PORT:-443}"
  [[ -n "${VLESS_UUID}" && -n "${VLESS_HOST}" ]]
}
