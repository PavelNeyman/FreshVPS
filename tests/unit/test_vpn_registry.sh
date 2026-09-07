#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
TMP="$(mktemp -d)"
export FRESHVPS_ETC="${TMP}"
export FRESHVPS_STATE_DIR="${TMP}/state"
mkdir -p "${TMP}/secrets"
# fake secrets for apply skip
echo priv >"${TMP}/secrets/singbox_reality_private"
echo pub >"${TMP}/secrets/singbox_reality_public"
echo abcd >"${TMP}/secrets/singbox_short_id"
# shellcheck source=/dev/null
source "${ROOT}/lib/vpn-core.sh"
# stub apply to avoid sing-box
vpn_apply_config() { :; }
vpn_ensure_dirs
vpn_add_user "alice" "test" >/dev/null
vpn_user_exists alice
vpn_list_users | grep -q alice
[[ -f "${TMP}/clients/alice/subscription.txt" ]]
grep -q 'vless://' "${TMP}/clients/alice/subscription.txt"
grep -q 'hysteria2://' "${TMP}/clients/alice/subscription.txt"
vpn_revoke_user alice
if vpn_user_exists alice; then echo "revoke failed"; exit 1; fi
rm -rf "${TMP}"
echo "OK test_vpn_registry"
