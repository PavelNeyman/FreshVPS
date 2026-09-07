#!/bin/sh
# Idempotent LAN/WAN/DHCP/Wi-Fi for Cudy-like OpenWrt (sourced by install-openwrt.sh)
# Expects: site variables + uci-helpers + NETWORK_CHANGED DHCP_CHANGED WIRELESS_CHANGED

apply_site_network() {
    LAN_IP="${LAN_NETWORK}.${LAN_ROUTER_HOST}"

    echo
    echo "=== NETWORK ==="
    check_set network.lan.device 'br-lan' NETWORK_CHANGED
    check_set network.lan.proto 'static' NETWORK_CHANGED
    check_set network.lan.ipaddr "$LAN_IP" NETWORK_CHANGED
    check_set network.lan.netmask "$LAN_NETMASK" NETWORK_CHANGED
    check_set network.wan.device "$WAN_DEVICE" NETWORK_CHANGED
    check_set network.wan.proto "$WAN_PROTO" NETWORK_CHANGED

    echo
    echo "=== DHCP ==="
    check_set dhcp.lan.start "$DHCP_POOL_START" DHCP_CHANGED
    check_set dhcp.lan.limit "$DHCP_POOL_SIZE" DHCP_CHANGED
    check_set dhcp.lan.leasetime "$DHCP_LEASETIME" DHCP_CHANGED

    echo
    echo "=== WIFI 2.4 ==="
    check_set wireless.radio0.country "$WIFI_COUNTRY" WIRELESS_CHANGED
    check_set wireless.radio0.channel "$WIFI_2G_CHANNEL" WIRELESS_CHANGED
    check_set wireless.radio0.htmode "$WIFI_2G_HTMODE" WIRELESS_CHANGED
    check_set wireless.radio0.disabled '0' WIRELESS_CHANGED
    check_set wireless.default_radio0.device 'radio0' WIRELESS_CHANGED
    check_set wireless.default_radio0.network 'lan' WIRELESS_CHANGED
    check_set wireless.default_radio0.mode 'ap' WIRELESS_CHANGED
    check_set wireless.default_radio0.ssid "$WIFI_2G_SSID" WIRELESS_CHANGED
    check_set wireless.default_radio0.encryption "$WIFI_ENCRYPTION" WIRELESS_CHANGED
    check_set_secret wireless.default_radio0.key "$WIFI_PASSWORD" WIRELESS_CHANGED
    check_set wireless.default_radio0.disabled '0' WIRELESS_CHANGED

    echo
    echo "=== WIFI 5 ==="
    check_set wireless.radio1.country "$WIFI_COUNTRY" WIRELESS_CHANGED
    check_set wireless.radio1.channel "$WIFI_5G_CHANNEL" WIRELESS_CHANGED
    check_set wireless.radio1.htmode "$WIFI_5G_HTMODE" WIRELESS_CHANGED
    check_set wireless.radio1.disabled '0' WIRELESS_CHANGED
    check_set wireless.default_radio1.device 'radio1' WIRELESS_CHANGED
    check_set wireless.default_radio1.network 'lan' WIRELESS_CHANGED
    check_set wireless.default_radio1.mode 'ap' WIRELESS_CHANGED
    check_set wireless.default_radio1.ssid "$WIFI_5G_SSID" WIRELESS_CHANGED
    check_set wireless.default_radio1.encryption "$WIFI_ENCRYPTION" WIRELESS_CHANGED
    check_set_secret wireless.default_radio1.key "$WIFI_PASSWORD" WIRELESS_CHANGED
    check_set wireless.default_radio1.disabled '0' WIRELESS_CHANGED
}

commit_and_reload_site() {
    if [ "$NETWORK_CHANGED" -eq 0 ] && [ "$DHCP_CHANGED" -eq 0 ] && [ "$WIRELESS_CHANGED" -eq 0 ]; then
        echo "[OK] Site network: no changes"
        return 0
    fi

    BACKUP_TS="$(date +%Y%m%d-%H%M%S)"
    [ "$NETWORK_CHANGED" -eq 1 ] && cp /etc/config/network "/etc/config/network.bak.$BACKUP_TS" && echo "[BACKUP] network"
    [ "$DHCP_CHANGED" -eq 1 ] && cp /etc/config/dhcp "/etc/config/dhcp.bak.$BACKUP_TS" && echo "[BACKUP] dhcp"
    [ "$WIRELESS_CHANGED" -eq 1 ] && cp /etc/config/wireless "/etc/config/wireless.bak.$BACKUP_TS" && echo "[BACKUP] wireless"

    [ "$NETWORK_CHANGED" -eq 1 ] && uci commit network && echo "[COMMIT] network"
    [ "$DHCP_CHANGED" -eq 1 ] && uci commit dhcp && echo "[COMMIT] dhcp"
    [ "$WIRELESS_CHANGED" -eq 1 ] && uci commit wireless && echo "[COMMIT] wireless"

    # Order: dhcp/wifi first, network last (can drop SSH briefly)
    [ "$DHCP_CHANGED" -eq 1 ] && /etc/init.d/dnsmasq restart
    [ "$WIRELESS_CHANGED" -eq 1 ] && wifi reload
    [ "$NETWORK_CHANGED" -eq 1 ] && /etc/init.d/network reload
    echo "[OK] Site configuration applied"
}
