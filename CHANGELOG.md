# Changelog

## 0.4.0 — 2026-09-07

Greenfield-focused (no legacy migrations):

- SSH key preflight before disabling passwords
- Shared `lib/install-singbox.sh` (VPS + edge)
- Empty sing-box users until operator seed (no throwaway UUID)
- `/etc/freshvps/install.conf` for upgrades
- `freshvps-vpn link` → subscription; doctor variant/panels
- `scripts/smoke-host.sh`, `plan-deploy-targets.sh`, targets.example
- Docs: upgrade assumes recreate VPS

## 0.3.5 — unified panels compose, plan SSH
## 0.3.4 — freshvps-tests
## 0.3.3 — idempotent upgrade
