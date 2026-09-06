# Changelog

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
