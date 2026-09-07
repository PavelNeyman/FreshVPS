# FreshVPS

Modular **greenfield** bootstrap for Debian VPS (+ OpenWrt / edge).

```bash
# 1) SSH key on the VPS first
# 2) Install (prefer file — safest):
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh -o /tmp/fv.sh
sudo bash /tmp/fv.sh

# Also: curl … | sudo bash

sudo freshvps-doctor
sudo freshvps-smoke
sudo freshvps-vpn link operator
```

Tests: [Human](docs/TEST-PLAN-HUMAN.md) · [AI/SSH](docs/TEST-PLAN-AI.md)  
Docs: [INSTALL](docs/INSTALL.md) · [UPGRADE](docs/UPGRADE.md) · [SECURITY](docs/SECURITY.md) · [RU](docs/ru/)
