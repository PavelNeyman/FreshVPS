# FreshVPS

Modular bootstrap for a fresh **Debian** VPS. After install you get a **ready-to-use** system: operator VPN profile + QR, Blocky DNS, monitoring, Telegram operator panel, optional admin API.

**Version:** [VERSION](VERSION) · **Rules:** [AGENTS.md](AGENTS.md) · **Plan:** [docs/ROADMAP.md](docs/ROADMAP.md)

## Stack

| Component | Role |
|-----------|------|
| Hardening | SSH, fail2ban, firewall, BBR |
| **sing-box** | VLESS+Reality + Hysteria2, multi-user |
| **Blocky** | DNS filtering |
| **OpenSOHO** | OpenWrt central mgmt |
| **Uptime Kuma / Beszel** | Uptime + metrics |
| Telegram | Operator-only VPN admin |
| restic | Backups scaffolding |

## Quick start

```bash
git clone https://github.com/PavelNeyman/FreshVPS.git
cd FreshVPS
sudo bash install.sh
# read /etc/freshvps/READY.txt
sudo freshvps-doctor
```

Non-interactive: `sudo bash install.sh --config configs/freshvps.conf.example --non-interactive` (set `PUBLIC_IP`).

Static checks (no root): `bash scripts/check.sh`

## Docs

- [INSTALL.md](docs/INSTALL.md) · [VPN-USERS.md](docs/VPN-USERS.md) · [SHORTCUT-IOS.md](docs/SHORTCUT-IOS.md)
- [SMOKE.md](docs/SMOKE.md) · [VPS-PROVIDERS.md](docs/VPS-PROVIDERS.md) · [SECURITY.md](docs/SECURITY.md)

## Layout

```
install.sh  modules/  lib/  bin/  runtime/  configs/  docs/  scripts/
```
