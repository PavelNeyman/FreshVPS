# Netductor Roadmap

Former name: FreshVPS.

## Done (FreshVPS 0.5.x–0.6.x)

- Admin UI + metrics/probes + TG alerts (no Prometheus; Kuma/Beszel off by default)
- Multi-user VPN (VLESS+HY2), Telegram bot, API sessions
- Edge hub + OpenWrt shell agent (MVP), `api-bind`, optional OpenSOHO
- Release: `freshvps-tg` linux amd64/arm64

## Phase G — Netductor + Go orchestrator (active)

**Name fixed: Netductor.** Parallel Go migration; bash remains until replaced.

| Step | Status | Deliverable |
|------|--------|-------------|
| G0 | **done** | AGENTS.md 2.0, ARCHITECTURE.md (EN+RU), roadmap |
| G1 | **done** | `cmd/netductor` scaffold: version, doctor, vpn, edge, status |
| G2 | next | `netductor` in GitHub Releases; `freshvps-*` still work |
| G3 | | `netductor serve` — port API from Python to Go |
| G4 | | `install/upgrade` by planes |
| G5 | | `netductor-agent` |
| G6 | | TG rename; **GitHub repo rename → netductor** (owner) |
| G7 | | `/opt/netductor`; deprecate bash |

**Repo rename:** after **G2** (first `netductor` release assets published), not at G7.

## Later / out of scope

Mini App, per-user traffic, Prometheus, full OpenSOHO parity — unchanged.
