# Bootstrap, panels compose, SSH keys (English)

**RU:** [ru/BOOTSTRAP.md](ru/BOOTSTRAP.md)

## Install without `git clone`

On a **Debian VPS**:

```bash
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh | sudo bash
```

Pinned ref:

```bash
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh | sudo bash -s -- --ref main
```

Config-driven:

```bash
# upload freshvps.conf first, then:
curl -fsSL .../bootstrap.sh | sudo bash -s -- --non-interactive --config /root/freshvps.conf
```

Bootstrap downloads a **GitHub tarball** (no git required). Falls back to `git clone` only if the tarball fails and git exists.

| Mode | When |
|------|------|
| `vps` | Debian Linux (default on server) |
| `edge-client` | `--mode edge-client` |
| `openwrt` | OpenWrt host |
| `plan` | macOS / control machine — unpack tree, no local VPN install |

## Component selection

**Unchanged:** modules stay optional (`ENABLE_*` / TUI / config). Nothing forces “install everything”.

Optional Docker stack is unified in [`compose/panels.yml`](../compose/panels.yml) with **profiles**: `opensoho`, `kuma`, `beszel`, `beszel-agent`, `lampac`. Installer modules still enable only what you chose; compose is the single definition of panel containers.

Data plane (**sing-box**, **Blocky**) remains on the host.

## Portable TUI

[`lib/tui.sh`](../lib/tui.sh): `whiptail` → `dialog` → macOS `osascript` → plain `read`. No mandatory extra packages on OpenWrt (plain prompts).

## Plan mode (macOS)

```bash
curl -fsSL .../bootstrap.sh | bash -s -- --mode plan --keep
```

Use the unpacked tree to edit `configs/` and `openwrt/site.conf`, then deploy over **SSH keys** to VPS/routers. Full multi-target wizard is evolving; plan mode is the safe entry.

## SSH keys (recommended)

Prefer keys over passwords for VPS and OpenWrt.

| Approach | When |
|----------|------|
| **One operator key** | Single admin, small blast radius acceptable |
| **Separate keys** (`vps`, `routers`) | Better isolation if a router is stolen |

```bash
ssh-keygen -t ed25519 -f ~/.ssh/freshvps_ed25519 -C "freshvps-operator"
ssh-copy-id -i ~/.ssh/freshvps_ed25519.pub root@VPS
# OpenWrt: append pubkey to /etc/dropbear/authorized_keys
```

### Passphrase on the key?

| Key without passphrase | Convenient for automation; **risk** if laptop disk/agent is compromised |
| Key **with** passphrase + **ssh-agent** | Unlock once per session; better default for a Mac control plane |

Recommendation: **passphrase + agent** on the Mac; on servers, `authorized_keys` only (no password SSH). Avoid putting the private key on routers.
