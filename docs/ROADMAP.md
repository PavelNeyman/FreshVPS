# Netductor Roadmap

Former name: FreshVPS.

## Done (FreshVPS 0.5.x–0.6.x)

- Admin UI + metrics/probes + TG alerts
- Multi-user VPN, Telegram, edge hub
- Release: `freshvps-tg` (manual REST once)

## Phase G — Netductor + Go (active)

| Step | Status | Deliverable |
|------|--------|-------------|
| G0 | **done** | AGENTS 2.0, ARCHITECTURE, roadmap |
| G1 | **done** | `cmd/netductor` CLI scaffold |
| G2 | **CI ready** | Workflow `release-netductor.yml` on `v*` tags; docs/RELEASES.md. **Owner:** `git tag v0.7.0-dev && git push origin v0.7.0-dev` then rename repo when assets appear |
| G3 | **started** | `netductor serve` → `:8790/health` only; Python still on `:8787` |
| G4–G7 | | install, agent, TG, `/opt/netductor` |

**Repo rename:** after first successful tag release (G2 assets live).
