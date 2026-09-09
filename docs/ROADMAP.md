# Netductor Roadmap

Former name: FreshVPS.

## Done (FreshVPS 0.5.x–0.6.x)

- Admin UI + metrics/probes + TG alerts (no Prometheus; Kuma/Beszel off by default)
- Multi-user VPN (VLESS+HY2), Telegram bot, API sessions
- Edge hub + OpenWrt shell agent (MVP), `api-bind`, optional OpenSOHO
- Release: `freshvps-tg` linux amd64/arm64

## Phase G — Netductor + Go orchestrator (active)

**Name fixed: Netductor.** Parallel Go migration; bash remains until replaced.

| Step | Deliverable |
|------|-------------|
| G0 | AGENTS.md 2.0, ARCHITECTURE.md (EN+RU), this roadmap |
| G1 | `cmd/netductor` scaffold: `version`, `doctor` (bridge), `vpn list` via existing CLI |
| G2 | Compat: `netductor` binary in Releases; wrappers/`freshvps-*` still work |
| G3 | `netductor serve` — port API from Python to Go (Admin embed) |
| G4 | `netductor install/upgrade` — planes host→vpn→dns→core→edge |
| G5 | `cmd/netductor-agent` replace shell agent |
| G6 | Rename TG bot package/strings; optional GitHub repo rename |
| G7 | Cut over paths `/opt/netductor`; deprecate bash modules |

## Later

- Mini App (needs HTTPS/domain or tunnel)
- Per-user traffic stats
- Multi-site OpenWrt hardening of agent allowlist
- Cloudflare Tunnel module as documented extra

## Out of scope (for now)

- Prometheus/Grafana
- Full OpenWISP/OpenSOHO feature parity inside Admin
- Multi-tenant SaaS
