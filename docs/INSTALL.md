# FreshVPS installation

Produces a **ready-to-use** system: VPN with operator profile + QR, Blocky, monitoring, operator CLI/Telegram/API.

## Requirements

- Debian 12 or 13 (fresh VPS recommended)
- Root
- Outbound HTTPS (GitHub, Docker Hub, ghcr.io)

## Interactive

```bash
git clone https://github.com/PavelNeyman/FreshVPS.git
cd FreshVPS
sudo bash install.sh
```

Optional: `apt-get install -y whiptail`.

## Non-interactive

```bash
cp configs/freshvps.conf.example /root/freshvps.conf
# set PUBLIC_IP, TELEGRAM_BOT_TOKEN, TELEGRAM_ADMIN_ID
sudo bash install.sh --config /root/freshvps.conf --non-interactive
```

## After install

Read **`/etc/freshvps/READY.txt`** (also printed at the end of install).

| Path | Purpose |
|------|---------|
| `/etc/freshvps/READY.txt` | Operator summary |
| `/etc/freshvps/clients/operator/` | Your VLESS link + QR |
| `/etc/freshvps/vpn-users.json` | User registry |
| `/etc/freshvps/secrets/` | Reality keys, HY2, Telegram, … |
| `/usr/local/bin/freshvps-vpn` | User management CLI |

### VPN users

```bash
freshvps-vpn add alice "phone"
freshvps-vpn list
freshvps-vpn link alice
freshvps-vpn disable alice
freshvps-vpn revoke alice
freshvps-vpn session 72
```

Ports: **443** VLESS+Reality, **8443** Hysteria2 (shared password in secrets).

In the client, prefer **remote/server DNS** so queries use Blocky on the VPS.

### Telegram (operator only)

`/status` `/vpn_add` `/vpn_list` `/vpn_link` `/vpn_disable` `/vpn_enable` `/vpn_revoke` `/session` `/ready`

### Admin API (optional Shortcuts)

Default: `127.0.0.1:8787`.

```bash
freshvps-vpn session 72
# SSH tunnel: ssh -L 8787:127.0.0.1:8787 user@VPS
# or rebind: freshvps-vpn api-bind detect   # careful if IP is public
```

Shortcut build: [SHORTCUT-IOS.md](SHORTCUT-IOS.md).

### OpenSOHO

1. Secret: `/etc/freshvps/secrets/opensoho_shared_secret`
2. Router: `openwisp-config` + monitoring → `http://VPS_IP:8090`
3. Admin: `docker exec -it opensoho ./opensoho superuser upsert EMAIL PASS`

### Beszel

Hub UI → KEY/TOKEN → `sudo bash /opt/beszel/enable-agent.sh`

### Backups

```bash
sudo /opt/freshvps-backup/backup.sh
```

## Uninstall

```bash
sudo bash uninstall.sh
sudo bash uninstall.sh --purge-secrets
```
