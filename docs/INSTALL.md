# FreshVPS installation

## Requirements

- Debian 12 or 13 (fresh VPS recommended)
- Root access
- Outbound HTTPS to GitHub / Docker Hub

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
# edit values
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
| `http://IP:8090` | OpenSOHO (if binary started) |

### VPN clients

Read secrets:

```bash
sudo cat /etc/freshvps/secrets/singbox_uuid
sudo cat /etc/freshvps/secrets/singbox_reality_public
sudo cat /etc/freshvps/secrets/singbox_short_id
sudo cat /etc/freshvps/secrets/singbox_hy2_password
```

Build client profiles with server public IP, ports 443 (VLESS Reality) and 8443 (Hysteria2).

### OpenWrt

1. Shared secret: `/etc/freshvps/secrets/opensoho_shared_secret`
2. On router: install `openwisp-config` / `openwisp-monitoring`
3. Point controller URL to `http://VPS_IP:8090`

### Telegram

```bash
/opt/freshvps-telegram/status.sh
/opt/freshvps-telegram/notify.sh "test"
```

### Backups

```bash
sudo /opt/freshvps-backup/backup.sh
systemctl list-timers freshvps-backup.timer
```

## Re-run

Modules are mostly idempotent; markers live under `/var/lib/freshvps/modules/`. Remove a marker to force re-run of that module’s install function path (still safe for most packages).
