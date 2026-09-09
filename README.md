# Netductor

> Formerly **FreshVPS**. Repository name may still be `FreshVPS` during migration.

Personal **network control plane** for a Debian VPS (+ OpenWrt edge): VPN, DNS, Admin, Telegram, outbound router agents.

```bash
# Current bootstrap (legacy entrypoint, still valid):
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh -o /tmp/fv.sh
sudo bash /tmp/fv.sh

sudo freshvps-doctor
sudo freshvps-vpn link operator
```

**Direction:** single Go binary `netductor` from GitHub Releases (see [docs/ROADMAP.md](docs/ROADMAP.md) Phase G).

**Docs:** [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) · [docs/README.md](docs/README.md) · [docs/ru/README.md](docs/ru/README.md) · [AGENTS.md](AGENTS.md)
