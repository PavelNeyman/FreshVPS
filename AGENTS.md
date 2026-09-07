# AGENTS.md — FreshVPS

> Document Version: 1.1  
> Status: Approved

**Single source of truth for project rules and architecture.**  
**Conversation history must never replace this document.**

Progress: [docs/ROADMAP.md](docs/ROADMAP.md) · [docs/ru/ROADMAP.md](docs/ru/ROADMAP.md)

---

# Product category (mandatory for agents)

**Production (self-hosted, single-operator).** Not a lab prototype.

- Security, docs, tests, and releases are production-grade for automated Debian VPS bootstrap (VPN, DNS, OpenWrt/edge, monitoring, backups).
- Scope: **one VPS** + optional **OpenWrt sites** / **edge SBC** — not multi-tenant SaaS.
- Lab shortcuts are defects.

---

# Quick Start for AI Agents

1. Read this AGENTS.md entirely.
2. Read Project State + ROADMAP **Next**.
3. Inspect the **current repository**.
4. Confirm task does not violate Forbidden / Frozen Architecture.
5. If ambiguous → **STOP** and ask the owner.
6. Implement only the requested task.
7. Update docs when behaviour changes (EN + RU when docs change).
8. Prefer idempotent shell; run `scripts/check.sh` and unit tests when practical.
9. End with **one logical Git commit** per completed task.

---

# Editing AGENTS.md

**Do not modify AGENTS.md without explicit owner consent in the current task.**

---

# 1. Product

**FreshVPS** — modular bootstrap for a fresh Debian VPS plus optional edge roles.

| Component | Role |
|-----------|------|
| Hardening | SSH, fail2ban, firewall, BBR |
| **sing-box** | Multi-user **VLESS+Reality** + **Hysteria2** (per-user); subscription bundle |
| **Blocky** | DNS filtering |
| **OpenSOHO** | OpenWrt central mgmt |
| **Uptime Kuma / Beszel** | Monitoring (localhost + SSH tunnel) |
| Telegram | Operator bot (**Go** preferred; bash fallback) |
| Backups | restic scaffolding |
| **edge-client** | Home SBC VPN client |
| **openwrt/** | Site LAN/Wi-Fi + VPN client + OpenSOHO agent |
| Lampac | **Optional only** (`ENABLE_LAMPAC`) |

---

# 2–3. Philosophy / Principles

Simplicity, stability, readability, operator-owned security, modularity. Git is source of truth. Idempotent installs. No secrets in repo.

---

# 4. Frozen Architecture

| Layer | Choice |
|-------|--------|
| OS | Debian (VPS/edge); OpenWrt on routers |
| Installer | Bash modules |
| VPN | **sing-box** only |
| OpenWrt control | **OpenSOHO** |
| DNS | **Blocky** |
| Monitoring | Kuma + Beszel |
| Operator UX | CLI + Telegram (+ optional localhost API) |
| Docs | English + Russian |

**Forbidden without approval:** replace sing-box/OpenSOHO/Blocky defaults; commit secrets; multi-tenant SaaS; non-idempotent installers.

---

# 5–6. Technology / AI rules

Allowed: Debian, Bash, Go (Telegram bot), sing-box, Blocky, OpenSOHO, Kuma, Beszel, restic. New major languages/services need owner approval. No unrequested work; no drive-by refactors.

---

# 7. Code review

Report first; implement after approval or explicit order.

---

# 8. Layout

```
install.sh, install-edge.sh, install-openwrt.sh
modules/  lib/  runtime/  cmd/freshvps-tg/  openwrt/
configs/  docs/  docs/ru/  tests/  scripts/
```

---

# 9. Git and secrets

Default branch **main**. Never commit tokens, private keys, real Wi-Fi passwords, or live client UUID lists.

---

# 10. Definition of Done

Installs on clean Debian; documented roles; failures visible; unit tests for pure logic where present; EN+RU docs for user-facing changes.

---

# 12. Project State (mutable)

| Area | Status |
|------|--------|
| Category | Production self-hosted |
| Version | **0.3.x** line |
| VPN | Multi-user VLESS+HY2 + subscription artifacts |
| Telegram | Go bot in `cmd/freshvps-tg` |
| OpenWrt | Site installer + env detect + VPN client |
| Edge | `install-edge.sh` + MikroTik helper |
| Tests | `scripts/check.sh`, `tests/unit/*`, GitHub Actions CI |
| Lampac | Optional |
| Next | Owner VPS/OpenWrt smoke; tighten client templates |

---

Final: production quality, modular design, one commit per task, chat never overrides AGENTS.
