#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
# shellcheck source=/dev/null
source "${ROOT}/lib/vless-parse.sh"

URL='vless://11111111-1111-1111-1111-111111111111@1.2.3.4:443?encryption=none&flow=xtls-rprx-vision&security=reality&sni=www.example.com&fp=chrome&pbk=PUB&sid=abcd&type=tcp#alice'
vless_parse "${URL}" || { echo "parse failed"; exit 1; }
[[ "${VLESS_UUID}" == "11111111-1111-1111-1111-111111111111" ]] || { echo "uuid"; exit 1; }
[[ "${VLESS_HOST}" == "1.2.3.4" ]] || { echo "host"; exit 1; }
[[ "${VLESS_PORT}" == "443" ]] || { echo "port"; exit 1; }
[[ "${VLESS_SNI}" == "www.example.com" ]] || { echo "sni"; exit 1; }
[[ "${VLESS_PBK}" == "PUB" ]] || { echo "pbk"; exit 1; }
[[ "${VLESS_SID}" == "abcd" ]] || { echo "sid"; exit 1; }
[[ "${VLESS_NAME}" == "alice" ]] || { echo "name"; exit 1; }
echo "OK test_vless_parse"
