# FreshVPS

Modular bootstrap for a fresh **Debian** VPS: answer a few questions, get a hardened system with VPN, DNS filtering, OpenWrt management, monitoring, Telegram control, and backups.

**Version:** see [VERSION](VERSION)  
**Rules for contributors/agents:** [AGENTS.md](AGENTS.md)  
**Plan:** [docs/ROADMAP.md](docs/ROADMAP.md)

## v1 stack

| Component | Role |
|-----------|------|
| Hardening | SSH, fail2ban, firewall, unattended-upgrades, BBR |
| **sing-box** | VLESS + Reality + Hysteria2 |
| **OpenSOHO** | Central management of OpenWrt routers |
| **Blocky** | DNS + ad/tracker blocking (hybrid with routers) |
| **Uptime Kuma** | Uptime / HTTP checks + alerts |
| **Beszel** | Resource metrics |
| Telegram bot | Keys, status, notifications |
| Backups | restic (VPS + OpenWrt configs) |

## Requirements

- Fresh Debian 12 or 13
- Root (or sudo)
- Public IPv4 (recommended)

## Quick start

```bash
# On the VPS
git clone https://github.com/PavelNeyman/FreshVPS.git
cd FreshVPS
sudo bash install.sh
```

Non-interactive (optional config file):

```bash
sudo bash install.sh --config /path/to/freshvps.conf
```

## Layout

```
install.sh          # entrypoint
lib/                # shared helpers
modules/            # one module per concern
configs/            # templates (no secrets)
docs/               # ROADMAP, INSTALL, SECURITY
```

## License

TBD
