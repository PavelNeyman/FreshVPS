# Bootstrap

GitHub builds the repo `tar.gz` on the fly (no Actions required).

## VPS (Debian)

**Preferred** (avoids curl|bash quirks):

```bash
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh -o /tmp/fv.sh
sudo bash /tmp/fv.sh
```

Also supported: `curl … | sudo bash` (must already be able to use sudo/root).

Upgrade: `sudo bash /tmp/fv.sh --upgrade`

After install: `freshvps-doctor` · `freshvps-smoke` · `freshvps-vpn link operator`

## macOS (plan)

```bash
curl -fsSL …/bootstrap.sh | bash -s -- --mode plan --keep
```

Does not install VPN on the Mac.

## Modes

| OS | Default mode |
|----|----------------|
| Darwin | plan |
| Debian Linux | vps |
| OpenWrt | openwrt |

Override: `--mode vps|edge-client|openwrt|plan`
