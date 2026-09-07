# FreshVPS

Modular bootstrap for a **Debian** VPS (optional OpenWrt / edge). Operator VPN (**VLESS+Reality** + **Hysteria2**), Blocky, panels, Telegram (Go).

**Version:** [VERSION](VERSION) · **Rules:** [AGENTS.md](AGENTS.md)

**Docs:** [INSTALL](docs/INSTALL.md) · [BOOTSTRAP](docs/BOOTSTRAP.md) · [VPN users](docs/VPN-USERS.md)  
**RU:** [docs/ru/](docs/ru/)

## Quick start (no git required)

```bash
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh | sudo bash
```

Or clone and `sudo bash install.sh`. Component selection stays optional (TUI / config / `ENABLE_*`).

OpenWrt: copy `openwrt/` to the router → `sh install-openwrt.sh`.

Panels compose (profiles): `compose/panels.yml` → `/opt/freshvps/compose/panels.yml` after install.
