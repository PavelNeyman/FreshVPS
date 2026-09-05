# FreshVPS Roadmap

> Living plan. Agents follow **Next** and AGENTS.md.

## Goal (v1)

Modular Debian VPS bootstrap: hardening, sing-box (VLESS+Reality+Hysteria2), OpenSOHO, Blocky, Uptime Kuma, Beszel, Telegram bot, backups.

## Phases

### Phase 0–2 — Foundation, modules, integration — Done

### Phase 3 — Polish

| ID | Task | Status |
|----|------|--------|
| P3.1–P3.3, P3.5–P3.7 | uninstall, docs, OpenSOHO Docker, Beszel, TG scaffold | Done |
| P3.4 | Smoke-test on real Debian VPS | **Waiting on owner VPS** |
| P3.9 | Fix bugs from smoke-test | Pending |

### Phase 4 — Multi-user VPN (operator-only)

Design: [docs/VPN-USERS.md](VPN-USERS.md)

| ID | Task | Status |
|----|------|--------|
| P4.1 | Registry + `freshvps-vpn` CLI (add/list/disable/revoke/link) | Pending |
| P4.2 | sing-box multi-UUID render + validate + restart | Pending |
| P4.3 | DNS via Blocky for VPN clients | Pending |
| P4.4 | QR/link artifacts under `/etc/freshvps/clients/` | Pending |
| P4.5 | Telegram admin commands over CLI | Pending |
| P4.6 | VPN-only API + session tokens | Pending |
| P4.7 | Shortcuts doc: Face ID gate + local token file + QR | Pending |

**Agreed security for API/Shortcuts:** VPN-only · session token · Face ID gate on Shortcut · token in local file (not master in Shortcut body).

## Deferred

- Port knocking, AdGuard/Blocky UI, parental, reverse proxy, mesh
- Native iOS Keychain helper app
- Full web admin UI

## Next

1. Smoke-test on owner VPS
2. Phase 4 implementation (can start after or in parallel with P3.4 if owner prefers)

## Notes

- Secrets never in git; `/etc/freshvps/` on host only.
