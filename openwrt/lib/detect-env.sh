#!/bin/sh
# Probe interfaces/radios; set WAN_DEVICE / radio hints if unset or auto

detect_openwrt_env() {
    BOARD="$(cat /tmp/sysinfo/board_name 2>/dev/null || true)"
    MODEL="$(cat /tmp/sysinfo/model 2>/dev/null || true)"
    echo "[INFO] board=${BOARD:-?} model=${MODEL:-?}"

    # Radios: list wireless devices
    RADIOS="$(uci -q show wireless 2>/dev/null | sed -n "s/^wireless\.\(radio[0-9]*\)=wifi-device/\1/p" | sort -u)"
    if [ -z "$RADIOS" ]; then
        RADIOS="$(ls -1 /sys/class/ieee80211 2>/dev/null | sed 's/^/radio/' || true)"
    fi
    echo "[INFO] radios: ${RADIOS:-none}"

    # WAN: prefer existing uci wan device, else heuristic
    CUR_WAN="$(uci -q get network.wan.device 2>/dev/null || true)"
    if [ -n "$WAN_DEVICE" ] && [ "$WAN_DEVICE" != "auto" ]; then
        echo "[INFO] WAN_DEVICE from config: $WAN_DEVICE"
    elif [ -n "$CUR_WAN" ]; then
        WAN_DEVICE="$CUR_WAN"
        echo "[INFO] WAN_DEVICE from uci: $WAN_DEVICE"
    else
        # Prefer vlan/wan-looking interfaces that are not br-lan
        WAN_DEVICE=""
        for cand in eth0.2 eth1 wan wan.1 eth0; do
            if [ -d "/sys/class/net/$cand" ]; then
                WAN_DEVICE="$cand"
                break
            fi
        done
        if [ -z "$WAN_DEVICE" ]; then
            WAN_DEVICE="$(ip -o link show 2>/dev/null | awk -F': ' '$2 !~ /lo|br-lan|wlan|docker|veth/ {print $2; exit}')"
        fi
        WAN_DEVICE="${WAN_DEVICE:-eth0}"
        echo "[INFO] WAN_DEVICE detected: $WAN_DEVICE"
    fi

    # Profile hint (non-binding)
    case "$BOARD$model" in
        *tr1200*|*TR1200*|*cudy*) PROFILE_HINT=cudy-tr1200 ;;
        *) PROFILE_HINT=generic ;;
    esac
    echo "[INFO] profile hint: $PROFILE_HINT"

    # Map first two radios for 2g/5g section names used by site-network
    set -- $RADIOS
    RADIO_2G="${1:-radio0}"
    RADIO_5G="${2:-radio1}"
    IFACE_2G="default_${RADIO_2G}"
    IFACE_5G="default_${RADIO_5G}"
    # Legacy Cudy uses default_radio0 / default_radio1 — keep if present
    if uci -q get wireless.default_radio0 >/dev/null 2>&1; then
        IFACE_2G=default_radio0
        RADIO_2G=radio0
    fi
    if uci -q get wireless.default_radio1 >/dev/null 2>&1; then
        IFACE_5G=default_radio1
        RADIO_5G=radio1
    fi
    export WAN_DEVICE RADIO_2G RADIO_5G IFACE_2G IFACE_5G PROFILE_HINT BOARD MODEL
}
