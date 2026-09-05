# FreshVPS Roadmap

> Living plan. Agents follow **Next** and AGENTS.md.

## Goal (v1)

Modular Debian VPS bootstrap: hardening, sing-box (VLESS+Reality+Hysteria2), OpenSOHO, Blocky, Uptime Kuma, Beszel, Telegram bot, backups.

## Phases

### Phase 0 — Foundation ✅ / in progress

| ID | Task | Status |
|----|------|--------|
| P0.1 | AGENTS.md | Done |
| P0.2 | ROADMAP.md | Done |
| P0.3 | VERSION, .gitignore, README refresh | Next |
| P0.4 | `lib/` common helpers | Pending |
| P0.5 | `install.sh` entrypoint (TUI + non-interactive) | Pending |
| P0.6 | Directory layout `modules/`, `configs/` | Pending |

### Phase 1 — Core modules

| ID | Task | Status |
|----|------|--------|
| P1.1 | `modules/hardening.sh` | Pending |
| P1.2 | `modules/sing-box.sh` + configs | Pending |
| P1.3 | `modules/blocky.sh` + HaGeZi defaults | Pending |
| P1.4 | `modules/opensoho.sh` | Pending |
| P1.5 | `modules/kuma.sh` (Docker) | Pending |
| P1.6 | `modules/beszel.sh` (hub + local agent) | Pending |
| P1.7 | `modules/telegram.sh` | Pending |
| P1.8 | `modules/backup.sh` (restic) | Pending |

### Phase 2 — Integration

| ID | Task | Status |
|----|------|--------|
| P2.1 | Wire modules in install.sh with enable flags | Pending |
| P2.2 | Firewall rules coordination (nftables/ufw) | Pending |
| P2.3 | State file / idempotency markers | Pending |
| P2.4 | INSTALL.md operator guide | Pending |
| P2.5 | SECURITY.md notes | Pending |

### Phase 3 — Polish

| ID | Task | Status |
|----|------|--------|
| P3.1 | Non-interactive config file example | Pending |
| P3.2 | Uninstall / disable hooks per module | Pending |
| P3.3 | CHANGELOG.md | Pending |
| P3.4 | Smoke-test checklist | Pending |

## Deferred (not v1)

- Port knocking
- AdGuard Home / Blocky UI
- Parental profiles
- Reverse proxy / TLS front
- Mesh (Headscale etc.)

## Next

**P0.3 → P0.6 → Phase 1 modules in order P1.1 … P1.8 → Phase 2.**

## Notes

- Target: Debian 12/13, root or passwordless sudo.
- Secrets never in git; generated on host under `/etc/freshvps/`.
- Prefer official upstream binaries + systemd units.
