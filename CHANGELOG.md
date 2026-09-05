# Changelog

## 0.1.1 — 2026-09-05

### Added

- `uninstall.sh` with per-module uninstall hooks
- OpenSOHO: Docker-first (`ghcr.io/opensoho/opensoho`) + binary fallback
- Beszel: hub compose + `enable-agent.sh` after Hub UI KEY/TOKEN
- Telegram: long-poll bot `/status` `/vpn`, `vpn-export.sh`
- docs/SMOKE.md already present; ROADMAP updated for VPS test wait

### Changed

- Module uninstall leaves data/config on disk by default; `--purge-secrets` optional

## 0.1.0 — 2026-09-05

### Added

- AGENTS.md project rules
- Modular installer: `install.sh`, `lib/common.sh`
- Modules: hardening, sing-box, Blocky, OpenSOHO, Uptime Kuma, Beszel, Telegram, restic
- Example non-interactive config
- docs: ROADMAP, INSTALL, SECURITY, SMOKE
