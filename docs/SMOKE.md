# Smoke-test checklist (clean Debian VPS)

1. Clone repo, run `sudo bash install.sh --config configs/freshvps.conf.example --non-interactive` (set PUBLIC_IP first).
2. `systemctl is-active sing-box blocky` → active
3. `ss -lntup | grep -E '443|8443|53|4000|3001|8090|8091'` — expected listeners
4. `dig @127.0.0.1 example.com` and blocked domain (ads) returns NXDOMAIN or empty
5. Open Kuma / Beszel HTTP UIs, complete first-run setup
6. Read `/etc/freshvps/secrets/*` (permissions 600)
7. `ufw status` — SSH + service ports
8. `/opt/freshvps-backup/backup.sh` exits 0
9. Optional: Telegram notify after writing token files
10. Reboot VPS; services come back

Log issues against modules; fix before declaring 0.2.0.
