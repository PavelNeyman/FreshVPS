# FreshVPS

Modular bootstrap for a **Debian** VPS (and optional OpenWrt / edge). After install: operator VPN (**VLESS+Reality** + **Hysteria2** multi-user), Blocky, monitoring, Telegram (Go).

**Version:** [VERSION](VERSION) · **Rules:** [AGENTS.md](AGENTS.md)

**Docs EN:** [docs/INSTALL.md](docs/INSTALL.md) · [docs/VPN-USERS.md](docs/VPN-USERS.md) · [docs/ROADMAP.md](docs/ROADMAP.md)  
**Документация RU:** [docs/ru/INSTALL.md](docs/ru/INSTALL.md) · [docs/ru/VPN-USERS.md](docs/ru/VPN-USERS.md)

## Quick start (VPS)

```bash
git clone https://github.com/PavelNeyman/FreshVPS.git && cd FreshVPS
sudo bash install.sh
sudo freshvps-doctor
```

OpenWrt: copy `openwrt/` to the router, edit `site.conf`, run `sh install-openwrt.sh`.

Tests: `bash scripts/check.sh && bash tests/unit/test_vless_parse.sh && bash tests/unit/test_vpn_registry.sh`
