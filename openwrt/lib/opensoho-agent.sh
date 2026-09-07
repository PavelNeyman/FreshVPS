#!/bin/sh
# OpenSOHO uses OpenWISP config agent on OpenWrt

apply_opensoho_agent() {
    [ "${ENABLE_OPENSOHO:-0}" = "1" ] || { echo "[SKIP] OpenSOHO agent"; return 0; }
    [ -n "$OPENSOHO_URL" ] || { echo "[ERROR] OPENSOHO_URL required"; return 1; }
    [ -n "$OPENSOHO_SHARED_SECRET" ] || { echo "[ERROR] OPENSOHO_SHARED_SECRET required"; return 1; }

    # strip trailing slash
    _url="${OPENSOHO_URL%/}"

    if command -v apk >/dev/null 2>&1; then
        apk add openwisp-config openwisp-monitoring 2>/dev/null || \
            apk add openwisp-config 2>/dev/null || true
    else
        opkg update 2>/dev/null || true
        opkg install openwisp-config openwisp-monitoring 2>/dev/null || \
            opkg install openwisp-config 2>/dev/null || {
            echo "[WARN] openwisp-config package missing — install feed manually"
            return 1
        }
    fi

    # openwisp-config UCI (controller URL + shared secret)
    uci -q set openwisp.http=controller || uci set openwisp.http=controller
    # common keys across openwisp-config versions
    uci set openwisp.http.url="$_url"
    uci set openwisp.http.shared_secret="$OPENSOHO_SHARED_SECRET"
    uci -q set openwisp.http.verify_ssl='0'
    uci -q set openwisp.http.mac_address_management='1'
    uci commit openwisp 2>/dev/null || true

    [ -x /etc/init.d/openwisp-config ] && /etc/init.d/openwisp-config enable
    [ -x /etc/init.d/openwisp-config ] && /etc/init.d/openwisp-config restart
    [ -x /etc/init.d/openwisp-monitoring ] && /etc/init.d/openwisp-monitoring enable
    [ -x /etc/init.d/openwisp-monitoring ] && /etc/init.d/openwisp-monitoring restart 2>/dev/null || true

    echo "[OK] OpenSOHO/OpenWISP agent -> $_url"
    echo "     Ensure router can reach OpenSOHO (VPN or public URL)."
}
