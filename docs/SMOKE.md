# Smoke-test checklist (English)

**RU:** [ru/SMOKE.md](ru/SMOKE.md)

1. Install exits 0; `/etc/freshvps/READY.txt` present.
2. `systemctl is-active sing-box blocky` → active.
3. `freshvps-vpn list` shows `operator`; `subscription.txt` has `vless://` and `hysteria2://`.
4. `sing-box check -c /usr/local/etc/sing-box/config.json` OK.
5. Listeners on 443, 8443, 53.
6. `dig @127.0.0.1 example.com` works.
7. Import subscription on a phone; traffic works.
8. `freshvps-vpn add smoke1 test && freshvps-vpn revoke smoke1`.
9. API: `freshvps-vpn session 1` + curl `127.0.0.1:8787`.
10. Telegram `/status` `/ready` (if bot enabled).
11. Panels via SSH tunnel; UFW sane; backup script present.
12. Reboot; services return.

Automated host slice: `sudo freshvps-smoke` / `freshvps-doctor`.
