# FreshVPS

Modular **greenfield** bootstrap for Debian VPS (+ OpenWrt / edge). Multi-user VLESS+Reality + Hysteria2, Blocky, panels, Telegram.

```bash
# Prefer SSH key on the VPS first, then:
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh | sudo bash

sudo freshvps-doctor
sudo freshvps-vpn link operator    # subscription (VLESS + HY2)
sudo freshvps-tests --default     # optional network probes
```

Docs: [INSTALL](docs/INSTALL.md) · [UPGRADE](docs/UPGRADE.md) · [SMOKE](docs/SMOKE.md) · [VPS-TESTS](docs/VPS-TESTS.md)  
RU: [docs/ru/](docs/ru/) · Version: [VERSION](VERSION)
