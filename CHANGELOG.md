# Changelog

## 0.1.0 — 2026-09-05

### Added

- AGENTS.md project rules
- Modular installer: `install.sh`, `lib/common.sh`
- Modules: hardening, sing-box (VLESS Reality + Hysteria2), Blocky (HaGeZi), OpenSOHO, Uptime Kuma, Beszel, Telegram helpers, restic backups
- Example non-interactive config
- docs: ROADMAP, INSTALL, SECURITY
- VERSION, .gitignore, README

### Notes

- First runnable skeleton; needs smoke-test on clean Debian
- OpenSOHO binary fetch is best-effort across release asset names
- Telegram module is notify/status scaffold, not a full key-management bot yet
