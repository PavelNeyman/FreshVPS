# Changelog

## 0.2.0 — 2026-09-06

### Added

- Multi-user VPN: `lib/vpn-core.sh`, `freshvps-vpn` CLI, auto **operator** profile + QR
- Admin API + session tokens (`vpn-api`)
- Telegram: `/vpn_add` `/vpn_list` `/vpn_link` `/vpn_disable` `/vpn_enable` `/vpn_revoke` `/session` `/ready` `/shortcut`
- `docs/SHORTCUT-IOS.md`
- Post-install **`/etc/freshvps/READY.txt`** (ready-to-use summary)

### Changed

- Install targets a usable operator system, not empty defaults
- sing-box: secrets/binary; users via vpn-core after Blocky

## 0.1.1 — 2026-09-05

Uninstall, OpenSOHO Docker, Beszel agent helper, TG scaffold

## 0.1.0 — 2026-09-05

Initial modular installer
