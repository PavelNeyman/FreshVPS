# FreshVPS installation (English)

**RU:** [ru/INSTALL.md](ru/INSTALL.md)

Produces a **ready-to-use** system: multi-user VPN (VLESS+Reality + Hysteria2), Blocky, monitoring, operator CLI / Telegram (Go) / API.

## Requirements

- Debian 12/13 (fresh VPS recommended)
- Root; outbound HTTPS (GitHub, Docker Hub / mirrors)

## Roles

| Role | Command |
|------|---------|
| VPS (default) | `sudo bash install.sh` |
| Edge SBC | `sudo bash install.sh --role edge-client` |
| OpenWrt help | `bash install.sh --role openwrt` then run on router |

## Interactive VPS

```bash
git clone https://github.com/PavelNeyman/FreshVPS.git
cd FreshVPS
sudo bash install.sh
```

Optional: `apt-get install -y whiptail`.

## Non-interactive

```bash
cp configs/freshvps.conf.example /root/freshvps.conf
# PUBLIC_IP, TELEGRAM_BOT_TOKEN, TELEGRAM_ADMIN_ID
sudo bash install.sh --config /root/freshvps.conf --non-interactive
```

## After install

Read **`/etc/freshvps/READY.txt`**.

| Path | Purpose |
|------|---------|
| `/etc/freshvps/clients/<name>/subscription.txt` | VLESS + HY2 links (handoff) |
| `/etc/freshvps/clients/<name>/qr.png` | VLESS QR |
| `/usr/local/bin/freshvps-vpn` | User CLI |
| `/opt/freshvps/bin/freshvps-tg` | Go Telegram bot binary |

### VPN users

```bash
freshvps-vpn add alice "phone"
cat /etc/freshvps/clients/alice/subscription.txt
freshvps-vpn list
freshvps-vpn session 72
```

Ports: **443** TCP VLESS+Reality, **8443** UDP Hysteria2 (per-user passwords).

Client DNS: prefer **remote/server DNS** (Blocky on VPS).

### Telegram

`/status` `/vpn_add` `/vpn_list` `/vpn_link` `/vpn_disable` `/vpn_enable` `/vpn_revoke` `/session` `/ready`

### Admin API

Default `127.0.0.1:8787`. Session via `freshvps-vpn session`. See [SHORTCUT-IOS.md](SHORTCUT-IOS.md).

### OpenWrt (on the router)

```bash
scp -r openwrt root@ROUTER:/root/freshvps-openwrt
# site.conf: WIFI_PASSWORD, LAN; WAN_DEVICE=auto or eth0.2 (Cudy TR1200)
sh install-openwrt.sh
```

### Panels (localhost)

```bash
ssh -L 8090:127.0.0.1:8090 -L 3001:127.0.0.1:3001 -L 8091:127.0.0.1:8091 root@VPS
```

Lampac is **optional** (`ENABLE_LAMPAC=1`).

## Uninstall

`sudo bash uninstall.sh` · `sudo bash uninstall.sh --purge-secrets`
