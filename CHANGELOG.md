# Changelog

## 0.2.6 — 2026-09-07

### Added
- **edge-client**: `install-edge.sh` for home SBC (Armbian), no VPS stack
- MikroTik helper `--with-mikrotik` (policy route via address-list)
- `configs/edge-client.conf.example`

## 0.2.5 — 2026-09-06

### Added
- Optional Lampac NextGen module (ENABLE_LAMPAC): light profile, localhost, TorrServer/JacRed


## 0.2.4 — 2026-09-06

### Added
- Config vars OPENSOHO_ADMIN_EMAIL/PASSWORD and BESZEL_ADMIN_* (empty = auto-generate)
- OpenSOHO: auto-create admin via `superuser upsert`; credentials in secrets + READY
- Beszel: auto-create admin the same way; agent still needs KEY/TOKEN from Hub UI once
- Kuma: document first-browser admin + SEED-MONITORS.md (no stable CLI for setup)

### Changed
- OpenSOHO Docker: `--dir /data` so PocketBase data persists on host volume
- READY.txt: panel URLs as localhost + SSH tunnel; print admin emails from secrets


## 0.2.3 — 2026-09-06

### Fixed
- sing-box **1.14** config generation (new DNS server format, route rule actions, default_domain_resolver)

### Changed
- OpenSOHO / Uptime Kuma / Beszel bind to **127.0.0.1** (SSH tunnel for admin access)
- Blocky API port no longer opened in UFW by default


## 0.2.2 — 2026-09-06

### Added

- `runtime/` sources for API + Telegram (no giant heredocs)
- `freshvps-doctor`, `scripts/check.sh`
- `docs/VPS-PROVIDERS.md`
- sing-box config **variant** marker `/etc/freshvps/singbox-config-variant`

### Changed

- Telegram bot: safer arg parsing
- API rebind: UFW only with explicit `--ufw`

## 0.2.1 — 2026-09-06

api-bind, vpn_apply A/B/C, docs polish

## 0.2.0 — 2026-09-06

Multi-user VPN, API, READY.txt

## 0.1.x

Initial modular installer
