# FreshVPS Roadmap

> Living plan. Agents follow **Next** and AGENTS.md.

## Product bar

Install must leave a **ready-to-use** system: services running, operator VPN profile + QR, DNS via Blocky for clients, admin CLI/TG/API wired — not empty default configs.

## Phases 0–3

Installer skeleton, modules, uninstall, docs — largely done. Smoke-test still needs owner VPS.

## Phase 4 — Multi-user VPN + operator automation

Design: [VPN-USERS.md](VPN-USERS.md) · Shortcut (manual): [SHORTCUT-IOS.md](SHORTCUT-IOS.md)

| ID | Task | Status |
|----|------|--------|
| P4.1 | `lib/vpn-core.sh` + `freshvps-vpn` CLI | Done |
| P4.2 | Multi-UUID render + restart | Done |
| P4.3 | DNS via Blocky in sing-box route | Done (best-effort) |
| P4.4 | QR/link under `/etc/freshvps/clients/` | Done |
| P4.5 | Telegram operator commands | Done |
| P4.6 | localhost API + sessions | Done |
| P4.7 | SHORTCUT-IOS.md (build on device only) | Done |
| P4.9 | `READY.txt` end-of-install summary | Done |
| P4.10 | Smoke-test + fix apply/DNS | Pending VPS |
| P4.11 | API bind on VPN interface helper | Pending |

**Not in scope:** generating or hosting `.shortcut` on the VPS (operator builds once on iPhone/Mac).

## Next

1. Owner VPS smoke-test
2. Tighten sing-box DNS route for installed version
3. Optional VPN-interface API bind

## Deferred

Port knocking, public web admin, native iOS Keychain app, OpenWrt user flows, VPS-side Shortcut distribution
