# FreshVPS installation (English)

**RU:** [ru/INSTALL.md](ru/INSTALL.md)

Produces a **ready-to-use** system: multi-user VPN (VLESS+Reality + Hysteria2), Blocky, optional panels, operator CLI / Telegram / API.

## Requirements

- Debian 12/13 (fresh VPS recommended)
- Root; outbound HTTPS (GitHub, Docker Hub / mirrors)
- **SSH public key** in `authorized_keys` before install (password auth is disabled when keys exist)

## Recommended install (bootstrap)

GitHub serves a live `tar.gz` of the repo (no Actions needed).

```bash
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh -o /tmp/fv.sh
sudo bash /tmp/fv.sh
```

Also works: `curl … | sudo bash` (prefer the file form).

Upgrade later: `sudo bash /tmp/fv.sh --upgrade`

## Roles

| Role | How |
|------|-----|
| VPS (default on Debian) | bootstrap / `install.sh --role vps` |
| Edge SBC | `install.sh --role edge-client` or bootstrap `--mode edge-client` |
| OpenWrt | run `openwrt/install-openwrt.sh` **on the router** |
| macOS control plane | bootstrap `--mode plan` (no VPN on Mac) |

## Non-interactive

```bash
cp configs/freshvps.conf.example /root/freshvps.conf
# set PUBLIC_IP, TELEGRAM_*, ENABLE_* as needed
sudo bash /tmp/fv.sh --non-interactive --config /root/freshvps.conf
# or from a git tree: sudo bash install.sh --config … --non-interactive
```

## After install

Read **`/etc/freshvps/READY.txt`**.

| Path / command | Purpose |
|----------------|---------|
| `/etc/freshvps/clients/<name>/subscription.txt` | VLESS + HY2 (recommended handoff) |
| `freshvps-vpn link <name>` | Prints subscription by default |
| `freshvps-doctor` / `freshvps-smoke` | Health checks |
| `freshvps-tests` | Optional network probes |

Ports: **443** TCP VLESS+Reality, **8443** UDP Hysteria2. Client DNS: prefer remote/server (Blocky on VPS).

### VPN users

```bash
freshvps-vpn add alice "phone"
freshvps-vpn link alice
freshvps-vpn list
freshvps-vpn session 72
```

See [VPN-USERS.md](VPN-USERS.md).

### Telegram

Operator-only: `/status` `/vpn_add` `/vpn_list` `/vpn_link` `/vpn_disable` `/vpn_enable` `/vpn_revoke` `/session` `/ready`

### Admin API

Default `127.0.0.1:8787`. Auth: Bearer **session** from `freshvps-vpn session`. See [SHORTCUT-IOS.md](SHORTCUT-IOS.md).

### OpenWrt (on the router)

```bash
scp -r openwrt root@ROUTER:/root/freshvps-openwrt
# edit site.conf: WIFI_PASSWORD, LAN; WAN_DEVICE=auto or eth0.2 (Cudy TR1200)
ssh root@ROUTER 'cd /root/freshvps-openwrt && sh install-openwrt.sh'
```

### Panels (localhost)

```bash
ssh -L 8090:127.0.0.1:8090 -L 3001:127.0.0.1:3001 -L 8091:127.0.0.1:8091 root@VPS
```

Lampac is **optional** (`ENABLE_LAMPAC=1`).

## Uninstall

`sudo bash uninstall.sh` · `sudo bash uninstall.sh --purge-secrets`
