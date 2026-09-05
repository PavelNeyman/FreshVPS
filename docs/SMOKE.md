# Smoke-test checklist (clean Debian VPS)

1. Set `PUBLIC_IP` in config; run `sudo bash install.sh --config … --non-interactive` (or interactive).
2. Install exits 0; **`/etc/freshvps/READY.txt`** exists and lists operator `vless://` link.
3. `systemctl is-active sing-box blocky` → active.
4. `freshvps-vpn list` shows `operator` on.
5. `test -f /etc/freshvps/clients/operator/link.txt` and `qr.png`.
6. `sing-box check -c /usr/local/etc/sing-box/config.json` → OK.
7. `ss -lntup | grep -E '443|8443|53'` — listeners present.
8. `dig @127.0.0.1 example.com +short` works; a known ad domain is NXDOMAIN/blocked (HaGeZi).
9. Import operator link on a phone; traffic works; DNS preferably via tunnel.
10. `freshvps-vpn add smoke1 test && freshvps-vpn revoke smoke1`.
11. `freshvps-vpn session 1` prints token; `curl -s -H "Authorization: Bearer TOKEN" http://127.0.0.1:8787/vpn/users` (if API enabled).
12. Uptime Kuma / Beszel HTTP open; complete first-run UI.
13. `ufw status` — SSH + service ports.
14. `/opt/freshvps-backup/backup.sh` exits 0.
15. Optional: Telegram `/status` `/ready`.
16. Reboot; `sing-box` `blocky` and timers come back.

Log failures against modules; fix before calling the release battle-tested.
