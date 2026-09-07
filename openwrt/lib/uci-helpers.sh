#!/bin/sh
# Shared UCI helpers for OpenWrt (ash-compatible)
# shellcheck disable=SC2039

uci_get() {
    uci -q get "$1" 2>/dev/null || true
}

# check_set SECTION TARGET FLAG_VAR
check_set() {
    _section="$1"
    _target="$2"
    _flag="$3"
    _current="$(uci_get "$_section")"
    if [ "$_current" = "$_target" ]; then
        echo "[OK]     $_section = $_current"
    else
        echo "[CHANGE] $_section: '$_current' -> '$_target'"
        uci set "$_section=$_target"
        eval "$_flag=1"
    fi
}

check_set_secret() {
    _section="$1"
    _target="$2"
    _flag="$3"
    _current="$(uci_get "$_section")"
    if [ "$_current" = "$_target" ]; then
        echo "[OK]     $_section = <unchanged>"
    else
        echo "[CHANGE] $_section: <secret differs>"
        uci set "$_section=$_target"
        eval "$_flag=1"
    fi
}
