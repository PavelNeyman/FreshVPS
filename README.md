# FreshVPS

Modular **greenfield** bootstrap for Debian VPS (+ OpenWrt / edge).

```bash
# 1) SSH key on the VPS first
# 2) Install (prefer file):
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh -o /tmp/fv.sh
sudo bash /tmp/fv.sh

sudo freshvps-doctor
sudo freshvps-smoke
sudo freshvps-vpn link operator
```

**Docs (EN + RU, same meaning):** [docs/README.md](docs/README.md) · [docs/ru/README.md](docs/ru/README.md)
