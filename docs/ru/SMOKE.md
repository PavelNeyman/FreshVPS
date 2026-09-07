# Smoke-чеклист (русский)

**EN:** [../SMOKE.md](../SMOKE.md)

1. Установка завершилась с кодом 0; есть `/etc/freshvps/READY.txt`.
2. `systemctl is-active sing-box blocky` → active.
3. `freshvps-vpn list` показывает `operator`; в `subscription.txt` есть `vless://` и `hysteria2://`.
4. `sing-box check -c /usr/local/etc/sing-box/config.json` OK.
5. Слушают порты 443, 8443, 53.
6. `dig @127.0.0.1 example.com` работает.
7. Импорт subscription на телефоне; трафик идёт.
8. `freshvps-vpn add smoke1 test && freshvps-vpn revoke smoke1`.
9. API: `freshvps-vpn session 1` + curl на `127.0.0.1:8787`.
10. Telegram `/status` `/ready` (если бот включён).
11. Панели через SSH-туннель; UFW в порядке; backup-скрипт на месте.
12. После reboot сервисы поднимаются.

Автоматический срез на сервере: `sudo freshvps-smoke` / `freshvps-doctor`.
