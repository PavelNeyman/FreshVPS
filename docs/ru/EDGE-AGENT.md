# Агент FreshVPS (OpenWrt)

Исходящий агент за NAT → VPS. Не замена полному OpenSOHO, а **минимум**: heartbeat, uci get/show, reload wifi/network, logread, reboot.

Токен: `/etc/freshvps/secrets/edge_token`  
Файлы агента: `/opt/freshvps/edge/openwrt/`  
HTTPS снаружи: модуль **cloudflared** (Cloudflare Tunnel на `127.0.0.1:8787`).

Подробности: [../EDGE-AGENT.md](../EDGE-AGENT.md)
