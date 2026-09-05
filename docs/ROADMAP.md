# FreshVPS Roadmap

> Living plan. Agents follow **Next** and AGENTS.md.

## Product bar

Install must leave a **ready-to-use** system: services running, operator VPN profile + QR, DNS via Blocky for clients, admin CLI/TG/API wired — not empty default configs.

## Phases 0–3

Installer skeleton, modules, uninstall, docs — largely done. Smoke-test still needs owner VPS.

## Phase 4 — Multi-user VPN + operator automation

Design: [VPN-USERS.md](VPN-USERS.md) · Shortcut: [SHORTCUT-IOS.md](SHORTCUT-IOS.md)

| ID | Task | Status |
|----|------|--------|
| P4.1 | `lib/vpn-core.sh` + `freshvps-vpn` CLI | Done |
| P4.2 | Multi-UUID render + restart | Done |
| P4.3 | DNS via Blocky in sing-box route | Done (best-effort validate) |
| P4.4 | QR/link under `/etc/freshvps/clients/` | Done |
| P4.5 | Telegram operator commands | Done |
| P4.6 | VPN/localhost API + sessions | Done |
| P4.7 | SHORTCUT-IOS.md | Done |
| P4.8 | Shortcut file host dir + `/shortcut` | Done (template upload by operator) |
| P4.9 | `READY.txt` end-of-install summary | Done |
| P4.10 | Smoke-test + fix apply/DNS edge cases | Pending VPS |
| P4.11 | API bind on VPN interface helper | Pending |

## Next

1. Owner VPS smoke-test
2. Tighten sing-box DNS route for installed sing-box version
3. Optional: serve `.shortcut` over VPN bind

## Deferred

Port knocking, public web admin, native iOS Keychain app, OpenWrt user flows
