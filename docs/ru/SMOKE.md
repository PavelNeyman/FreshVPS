# Smoke-тест (русский)

**EN:** [../SMOKE.md](../SMOKE.md)

1. `install.sh` exit 0, есть `/etc/freshvps/READY.txt`
2. `sing-box` + `blocky` active
3. `operator` в `freshvps-vpn list`; есть `subscription.txt` (vless + hysteria2)
4. `sing-box check` OK; слушатели 443/8443/53
5. `dig @127.0.0.1` работает
6. Импорт subscription на телефон
7. `add` / `revoke` тестового пользователя
8. API session + curl localhost
9. Telegram `/status` `/ready` (Go-бот)
10. Reboot — сервисы поднимаются
