# FreshVPS Roadmap

> Living plan. Agents follow **Next** and AGENTS.md.

## Goal (v1)

Modular Debian VPS bootstrap: hardening, sing-box (VLESS+Reality+Hysteria2), OpenSOHO, Blocky, Uptime Kuma, Beszel, Telegram bot, backups.

## Phases

### Phase 0 — Foundation — Done

### Phase 1 — Core modules — Done

### Phase 2 — Integration — Done

### Phase 3 — Polish

| ID | Task | Status |
|----|------|--------|
| P3.1 | Non-interactive config file example | Done |
| P3.2 | Uninstall / disable hooks per module | Done (`uninstall.sh`) |
| P3.3 | CHANGELOG.md | Done |
| P3.4 | Smoke-test on real Debian VPS | **Waiting on owner VPS** |
| P3.5 | Telegram bot (`/status` `/vpn`) | Done (minimal long-poll) |
| P3.6 | OpenSOHO Docker-first + binary fallback | Done |
| P3.7 | Beszel agent enable script after Hub UI | Done |
| P3.8 | Multi-user VPN key issue/revoke in bot | Pending |
| P3.9 | Fix bugs found in smoke-test | Pending |

## Deferred (not v1)

- Port knocking
- AdGuard Home / Blocky UI
- Parental profiles
- Reverse proxy / TLS front
- Mesh (Headscale etc.)

## Next

1. **Smoke-test on owner VPS** (`docs/SMOKE.md`)
2. Fix install bugs from real run
3. Optional: multi-client UUID management in Telegram bot

## Notes

- Target: Debian 12/13, root or passwordless sudo.
- Secrets never in git; generated on host under `/etc/freshvps/`.
- Prefer official upstream binaries + systemd / Docker.
