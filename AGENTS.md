# AGENTS.md — Netductor

> Document Version: **2.0**  
> Status: **Approved** (owner: rename FreshVPS → Netductor + Go orchestrator)  
> Former name: **FreshVPS** (compat until migration complete)

**Single source of truth for project rules and architecture.**  
**Conversation history must never replace this document.**

Progress: [docs/ROADMAP.md](docs/ROADMAP.md) · [docs/ru/ROADMAP.md](docs/ru/ROADMAP.md) · [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)

---

# Product category (mandatory for agents)

**Production (self-hosted, single-operator).** Not a lab prototype.

- Security, docs, tests, and releases are production-grade.
- Scope: **one control-plane host (VPS)** + optional **OpenWrt sites** / edge clients — **not** multi-tenant SaaS.
- Lab shortcuts are defects.

---

# Quick Start for AI Agents

1. Read this AGENTS.md entirely.
2. Read Project State + ROADMAP **Next** + ARCHITECTURE.
3. Inspect the **current repository** (still may contain FreshVPS paths during migration).
4. Confirm task does not violate Forbidden / Frozen Architecture.
5. If ambiguous → **STOP** and ask the owner.
6. Implement only the requested task.
7. Update docs when behaviour changes (**EN + RU** when docs change).
8. Prefer **Go** for new orchestration/API/agent code; bash only as bridge or OpenWrt ash where needed.
9. End with **one logical Git commit** per completed task.

---

# Editing AGENTS.md

**Do not modify AGENTS.md without explicit owner consent in the current task.**

---

# 1. Product

**Netductor** (net + conductor) — personal **network control plane**: orchestrate VPN, DNS, edge OpenWrt agents, operator UX (CLI / Admin / Telegram).

| Plane | Role |
|-------|------|
| **Host** | Debian bootstrap, hardening, swap, packages |
| **VPN** | **sing-box** multi-user **VLESS+Reality** + **Hysteria2**; subscriptions / QR |
| **DNS** | **Blocky** |
| **Core / API** | Sessions, metrics/probes, edge registry, Admin SPA |
| **Edge** | Outbound OpenWrt agent (heartbeat + allowlisted commands) |
| **Operator** | CLI `netductor`, Telegram bot, Admin UI |
| **Extras (optional)** | OpenSOHO, Cloudflare Tunnel, Lampac, legacy Kuma/Beszel — **default off** |

**Not the product:** multi-tenant SaaS, full OpenWISP clone, Prometheus/Grafana stack.

---

# 2. Philosophy

- Simplicity, stability, readability, operator-owned security, modularity.
- Git is source of truth; **releases ship binaries** (primary install path long-term).
- Idempotent installs; no secrets in repo.
- **One contract:** operator actions go through API/CLI; TG and Admin are clients.
- Evolution over big-bang rewrites; compat aliases during rename.

---

# 3. Target architecture (Frozen direction)

```
netductor (Go CLI + serve)     netductor-agent (OpenWrt)
        │                              │
        ▼                              │ outbound only
   Core API + Admin embed  ◄───────────┘
        │
   planes: host | vpn | dns | edge | extras
```

| Layer | Choice |
|-------|--------|
| Product name | **Netductor** |
| CLI binary | **`netductor`** (optional short alias `nd` later) |
| Edge binary | **`netductor-agent`** |
| Telegram | **`netductor-tg`** or subcommand `netductor telegram` (migrate from `freshvps-tg`) |
| Orchestrator language | **Go** (modules as Go packages; bash bridge until ported) |
| OS | Debian (control plane); OpenWrt on routers |
| VPN | **sing-box** only |
| DNS | **Blocky** |
| Edge control | **Own agent** (primary); OpenSOHO **optional/legacy** |
| Monitoring | Built-in metrics/probes + TG alerts (not Kuma/Beszel by default) |
| Operator UX | CLI + Admin UI + Telegram |
| Docs | English + Russian, same meaning |
| Deploy | GitHub **Releases** single binary; repo clone secondary |

### Migration rules (FreshVPS → Netductor)

1. **Parallel path:** new Go code under `cmd/netductor`, `cmd/netductor-agent`; do not delete working bash until plane is covered.
2. Paths `/opt/freshvps`, `freshvps-*` remain valid until explicit cutover; add `netductor` symlinks/wrappers.
3. Repo may stay `FreshVPS` until owner renames GitHub repo to `netductor` (or new repo + archive).
4. User-facing strings prefer **Netductor**; logs may mention legacy paths during transition.
5. New features → **Go first**. Bash only if Go path not ready and owner accepts bridge.

### Forbidden without approval

- Replace sing-box or Blocky defaults.
- Make OpenSOHO/Kuma/Beszel/Lampac default-on again.
- Commit secrets, multi-tenant SaaS, non-idempotent installers.
- Big-bang delete of bash modules without Go replacement + doctor/smoke.
- Rename GitHub repository without owner confirmation of exact new repo name/visibility.

---

# 4. Technology

**Allowed:** Go, Debian, Bash/ash (bridge/OpenWrt), sing-box, Blocky, optional Docker for extras, restic scaffolding, cloudflared (optional).

**Deprecated direction:** Python Admin API (`runtime/api/server.py`) → port to Go `serve`; shell edge agent → `netductor-agent`.

**Not without approval:** new major languages, k8s requirement, Prometheus stack.

---

# 5. Layout (target + current)

**Target**

```
cmd/netductor/          # CLI: install, upgrade, doctor, vpn, edge, serve, menu
cmd/netductor-agent/    # OpenWrt outbound agent
cmd/netductor-tg/       # Telegram (or merge into netductor)
internal/               # api, planes, host, vpn, dns, edge, ui embed
docs/  docs/ru/
```

**Current (legacy, still valid)**

```
install.sh  modules/  lib/  runtime/  cmd/freshvps-tg/  edge/openwrt/
bin/freshvps-*  openwrt/  bootstrap.sh
```

---

# 6. Git and secrets

Default branch **main**. Never commit tokens, private keys, real Wi-Fi passwords, or live client UUID lists.

---

# 7. Definition of Done

- Behaviour works on clean Debian for touched planes.
- `doctor` / smoke still meaningful.
- Docs EN+RU for user-facing changes.
- One logical commit per task.
- During migration: legacy entrypoints still work or documented breakage is explicit.

---

# 8. Project State (mutable)

| Area | Status |
|------|--------|
| Name | **Netductor** (fixed); code/docs migrating from FreshVPS |
| Category | Production self-hosted control plane |
| Version line | **0.6.x** (bash+hybrid); next major Go orchestrator track **0.7+** |
| VPN | Multi-user VLESS+HY2 + subscription |
| Admin UI | Built-in SPA + metrics/probes (no Prometheus) |
| Telegram | Go bot `cmd/freshvps-tg` → rename to netductor |
| Edge | Shell agent + hub API; Go agent planned |
| OpenSOHO / Kuma / Beszel | Optional, default off |
| Go orchestrator | **In progress** — see ROADMAP Phase G |
| Next | Scaffold `cmd/netductor`, ARCHITECTURE, compat wrappers; port serve/vpn/edge |

---

# 9. Code review

Report first when asked; implement when owner says to proceed. No drive-by refactors outside the task.

---

Final: production quality, planes architecture, Go orchestrator direction, Netductor name, chat never overrides AGENTS.
