# FreshVPS installation

## Requirements

- Debian 12 or 13 (fresh VPS recommended)
- Root access
- Outbound HTTPS to GitHub / Docker Hub / ghcr.io

## Interactive

```bash
git clone https://github.com/PavelNeyman/FreshVPS.git
cd FreshVPS
sudo bash install.sh
```

Optional: `apt-get install -y whiptail` for dialog UI.

## Non-interactive

```bash
cp configs/freshvps.conf.example /root/freshvps.conf
# edit values (PUBLIC_IP, TELEGRAM_*, ports)
sudo bash install.sh --config /root/freshvps.conf --non-interactive
```

## After install

| Path | Purpose |
|------|---------|
| `/etc/freshvps/secrets/` | UUID, Reality keys, passwords, Telegram token |
| `/var/lib/freshvps/` | Module done markers |
| `/var/log/freshvps.log` | Install log |
| `/usr/local/etc/sing-box/config.json` | VPN config |
| `/etc/blocky/config.yml` | DNS |
| `http://IP:3001` | Uptime Kuma setup |
| `http://IP:8091` | Beszel hub |
| `http://IP:8090` | OpenSOHO |

### VPN clients

```bash
sudo /opt/freshvps-telegram/vpn-export.sh
# or
sudo cat /etc/freshvps/secrets/singbox_uuid
sudo cat /etc/freshvps/secrets/singbox_reality_public
sudo cat /etc/freshvps/secrets/singbox_short_id
sudo cat /etc/freshvps/secrets/singbox_hy2_password
```

Ports: **443** VLESS+Reality, **8443** Hysteria2.

### OpenWrt / OpenSOHO

1. Shared secret: `/etc/freshvps/secrets/opensoho_shared_secret`
2. On router: `openwisp-config` + `openwisp-monitoring`
3. Server URL: `http://VPS_IP:8090` (no trailing slash)
4. Admin user: `docker exec -it opensoho ./opensoho superuser upsert EMAIL PASS` (Docker install)

### Beszel agent

1. Open hub UI, create account, Add System → copy KEY + TOKEN
2. `export BESZEL_KEY='...' BESZEL_TOKEN='...'`
3. `sudo bash /opt/beszel/enable-agent.sh`

### Telegram

Bot commands (admin chat only): `/status` `/vpn` `/help`

```bash
sudo systemctl status freshvps-telegram-bot
/opt/freshvps-telegram/notify.sh "test"
```

### Backups

```bash
sudo /opt/freshvps-backup/backup.sh
systemctl list-timers freshvps-backup.timer
```

## Uninstall modules

```bash
sudo bash uninstall.sh              # all service modules
sudo bash uninstall.sh sing-box kuma # selected
sudo bash uninstall.sh --purge-secrets
```

Hardening (SSH/UFW) is not rolled back automatically.

## Re-run

Markers: `/var/lib/freshvps/modules/*.done`. Remove a marker to re-run that module path.
