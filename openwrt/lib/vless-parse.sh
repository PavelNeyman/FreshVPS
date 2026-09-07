#!/bin/sh
# Ash-compatible vless:// parser — sets VLESS_* variables

vless_parse() {
    _url="$1"
    VLESS_UUID=; VLESS_HOST=; VLESS_PORT=; VLESS_SNI=; VLESS_PBK=; VLESS_SID=; VLESS_FP=; VLESS_FLOW=; VLESS_NAME=
    case "$_url" in
        vless://*) ;;
        *) return 1 ;;
    esac
    _rest="${_url#vless://}"
    case "$_rest" in
        *#*) VLESS_NAME="${_rest##*#}"; _rest="${_rest%%#*}" ;;
    esac
    VLESS_UUID="${_rest%%@*}"
    _rest="${_rest#*@}"
    case "$_rest" in
        *\?*) _hp="${_rest%%\?*}"; _qs="${_rest#*\?}" ;;
        *) _hp="$_rest"; _qs= ;;
    esac
    VLESS_HOST="${_hp%%:*}"
    VLESS_PORT="${_hp##*:}"
    VLESS_SNI="$(echo "$_qs" | tr '&' '\n' | sed -n 's/^sni=//p' | head -1)"
    VLESS_PBK="$(echo "$_qs" | tr '&' '\n' | sed -n 's/^pbk=//p' | head -1)"
    VLESS_SID="$(echo "$_qs" | tr '&' '\n' | sed -n 's/^sid=//p' | head -1)"
    VLESS_FP="$(echo "$_qs" | tr '&' '\n' | sed -n 's/^fp=//p' | head -1)"
    VLESS_FLOW="$(echo "$_qs" | tr '&' '\n' | sed -n 's/^flow=//p' | head -1)"
    [ -n "$VLESS_SNI" ] || VLESS_SNI=www.cloudflare.com
    [ -n "$VLESS_FP" ] || VLESS_FP=chrome
    [ -n "$VLESS_FLOW" ] || VLESS_FLOW=xtls-rprx-vision
    [ -n "$VLESS_PORT" ] || VLESS_PORT=443
    [ -n "$VLESS_UUID" ] && [ -n "$VLESS_HOST" ]
}
