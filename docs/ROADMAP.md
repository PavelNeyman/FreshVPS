# FreshVPS Roadmap

> Living plan. Agents follow **Next** and AGENTS.md.

## Goal (v1)

Modular Debian VPS bootstrap: hardening, sing-box (VLESS+Reality+Hysteria2), OpenSOHO, Blocky, Uptime Kuma, Beszel, Telegram bot, backups.

## Phases

### Phase 0 — Foundation

| ID | Task | Status |
|----|------|--------|
| P0.1 | AGENTS.md | Done |
| P0.2 | ROADMAP.md | Done |
| P0.3 | VERSION, .gitignore, README refresh | Done |
| P0.4 | `lib/` common helpers | Done |
| P0.5 | `install.sh` entrypoint (TUI + non-interactive) | Done |
| P0.6 | Directory layout `modules/`, `configs/` | Done |

### Phase 1 — Core modules

| ID | Task | Status |
|----|------|--------|
| P1.1 | `modules/hardening.sh` | Done |
| P1.2 | `modules/sing-box.sh` | Done |
| P1.3 | `modules/blocky.sh` + HaGeZi defaults | Done |
| P1.4 | `modules/opensoho.sh` | Done (binary download best-effort) |
| P1.5 | `modules/kuma.sh` (Docker) | Done |
| P1.6 | `modules/beszel.sh` (hub) | Done |
| P1.7 | `modules/telegram.sh` | Done (scaffold) |
| P1.8 | `modules/backup.sh` (restic) | Done |

### Phase 2 — Integration

| ID | Task | Status |
|----|------|--------|
| P2.1 | Wire modules in install.sh with enable flags | Done |
| P2.2 | Firewall rules coordination (ufw in modules) | Done (baseline) |
| P2.3 | State file / idempotency markers | Done |
| P2.4 | INSTALL.md operator guide | Done |
| P2.5 | SECURITY.md notes | Done |

### Phase 3 — Polish

| ID | Task | Status |
|----|------|--------|
| P3.1 | Non-interactive config file example | Done |
| P3.2 | Uninstall / disable hooks per module | Pending |
| P3.3 | CHANGELOG.md | Next |
| P3.4 | Smoke-test checklist on real Debian VPS | Pending |
| P3.5 | Full Telegram bot (key issue / revoke) | Pending |
| P3.6 | OpenSOHO binary asset detection hardening | Pending |
| P3.7 | Beszel agent auto-join on same host | Pending |

## Deferred (not v1)

- Port knocking
- AdGuard Home / Blocky UI
- Parental profiles
- Reverse proxy / TLS front
- Mesh (Headscale etc.)

## Next

1. CHANGELOG.md for 0.1.0
2. Smoke-test on a real VPS and fix module bugs
3. Uninstall hooks
4. Richer Telegram bot

## Notes

- Target: Debian 12/13, root or passwordless sudo.
- Secrets never in git; generated on host under `/etc/freshvps/`.
- Prefer official upstream binaries + systemd units.
