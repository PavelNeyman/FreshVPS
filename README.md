# Netductor

Personal **network control plane** for a Debian VPS (+ OpenWrt edge): VPN, DNS, Admin, Telegram, outbound router agents.

Formerly **FreshVPS** (paths `/opt/freshvps`, `freshvps-*` still valid during migration).

```bash
# Bootstrap (Debian VPS)
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/netductor/main/bootstrap.sh -o /tmp/nd.sh
sudo bash /tmp/nd.sh

# Or CLI from release
curl -fsSL -o /usr/local/bin/netductor \
  https://github.com/PavelNeyman/netductor/releases/download/v0.7.0-dev/netductor-linux-amd64
chmod 755 /usr/local/bin/netductor
netductor version
```

**Docs:** [ARCHITECTURE](docs/ARCHITECTURE.md) · [ROADMAP](docs/ROADMAP.md) · [RELEASES](docs/RELEASES.md) · [AGENTS.md](AGENTS.md)
