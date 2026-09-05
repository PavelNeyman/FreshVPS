# Multi-user VPN & admin control

> Status: **Implemented in tree** (smoke-test pending on real VPS).
> Scope: people/devices → **VPS** only (not OpenWrt).

## Goals

- Only the operator issues access.
- Users never talk to Telegram; operator forwards link/QR.
- Automation: registry → sing-box → Blocky DNS path → reload → artifacts.
- Surfaces: **CLI**, **Telegram**, **API** (+ optional **Shortcuts** built on device — see [SHORTCUT-IOS.md](SHORTCUT-IOS.md)).

## On-disk layout

| Path | Purpose |
|------|---------|
| `/etc/freshvps/vpn-users.json` | Registry |
| `/etc/freshvps/clients/<name>/link.txt` | `vless://` |
| `/etc/freshvps/clients/<name>/qr.png` | QR |
| `/etc/freshvps/sessions/<token>` | API session expiry unix |
| `/etc/freshvps/READY.txt` | Post-install operator summary |
| `/usr/local/bin/freshvps-vpn` | CLI |
| `/opt/freshvps-api/server.py` | API |

## CLI

```bash
freshvps-vpn add alice "phone"
freshvps-vpn list
freshvps-vpn link alice
freshvps-vpn disable alice
freshvps-vpn enable alice
freshvps-vpn revoke alice
freshvps-vpn session 72
```

Install seeds user **`operator`** automatically.

## Telegram (operator chat only)

`/vpn_add` `/vpn_list` `/vpn_link` `/vpn_disable` `/vpn_enable` `/vpn_revoke` `/session` `/status` `/ready`

## API + Shortcuts

- Default bind: `127.0.0.1:8787`
- Auth: `Authorization: Bearer <session>`
- Session: `freshvps-vpn session` or `/session`
- How to build Shortcut: [SHORTCUT-IOS.md](SHORTCUT-IOS.md) (no server-side `.shortcut` packaging)

## DNS

`vpn_apply_config` points sing-box DNS at Blocky `127.0.0.1` when validation succeeds.

## Non-goals

OpenWrt onboarding, public registration, full web UI, native Keychain app, VPS hosting of Apple Shortcut binaries.
