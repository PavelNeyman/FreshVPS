# AGENTS.md — FreshVPS

> Document Version: 1.0  
> Status: Approved

**Single source of truth for project rules and architecture.**  
**Conversation history must never replace this document.**

Progress: [docs/ROADMAP.md](docs/ROADMAP.md) (when present)  
Also: README · SECURITY · INSTALL · CHANGELOG · VERSION (when present)

---

# Product category (mandatory for agents)

**Production (self-hosted, single-operator).** This is **not** a lab prototype.

- Treat behaviour, security, docs, tests, and releases as **production-grade** for an automated Debian VPS bootstrap (VPN, DNS, OpenWrt management, monitoring, backups).
- Scope remains **one VPS + optional managed OpenWrt routers**, not multi-tenant SaaS.
- Lab-quality shortcuts are **defects**, not acceptable trade-offs.
- Assume real credentials, VPN keys, SSH access, and router configs are at risk.

---

# Quick Start for AI Agents

1. Read **this** AGENTS.md entirely.
2. Read **Product category**, **Project State**, and **ROADMAP** (when present) → **Next**.
3. Inspect the **current repository** (do not rely on chat history alone).
4. Confirm the task does not violate **Forbidden** or **Frozen Architecture**.
5. If scope or design is ambiguous → **STOP** and ask the repository owner.
6. Implement **only** the requested task (or the single **Next** item when continuing the plan).
7. Update documentation when behaviour or modules change.
8. Update ROADMAP / Project State when progress changes.
9. Prefer idempotent shell; test modules in isolation when practical.
10. End with **one logical Git commit** per completed task.

Skipping these steps is a rule violation.

---

# Editing AGENTS.md (mandatory)

**Agents must NOT modify `AGENTS.md` unless the repository owner has given explicit consent in the current task.**

- Implied consent from unrelated tasks is **not** enough.
- When consent is given, bump the document version if rules change.

Violating this rule is a process **Critical** failure.

---

# 1. Product

**FreshVPS** is a **modular bootstrap** for a fresh Debian VPS: answer a few questions (TUI or non-interactive), get a hardened system with VPN, DNS filtering, OpenWrt central management, monitoring, Telegram control, and backups.

Ships as:

- modular **Bash** installer (`install.sh` + `modules/`);
- optional **whiptail** TUI and non-interactive / config-driven mode;
- declarative enable/disable of components.

**Primary stack (v1):**

| Component | Role |
|-----------|------|
| Hardening | SSH keys, fail2ban, firewall, unattended-upgrades, BBR |
| **sing-box** | VLESS + Reality + Hysteria2 |
| **OpenSOHO** | Central management of OpenWrt routers (outbound agent) |
| **Blocky** | Central DNS + ad/tracker blocking (hybrid with routers) |
| **Uptime Kuma** | Availability monitoring + alerts |
| **Beszel** | Resource monitoring (CPU/RAM/disk) |
| Telegram bot | VPN keys, status, notifications |
| Backups | VPS state + OpenWrt configs (restic/borg) |

**Deferred / optional (not v1 default):** port knocking, AdGuard Home, Blocky UI, parental profiles, reverse proxy, mesh.

---

# 2. Project Philosophy

- **Simplicity over cleverness**
- **Stability over novelty**
- **Readability over abstraction**
- **Local-first / operator-owned security**
- **Production quality within that scope**
- **Modularity** — easy to add, remove, or reconfigure components

Prefer standard Debian packages, official upstream binaries, and minimal custom code.

---

# 3. Project Principles

Immutable without owner approval:

- **Git is the source of truth**
- **Documentation is part of the product**
- **One repository — one truth**
- **Operator-owned** (single admin, not multi-tenant)
- **Idempotent installs** where practical

---

# 4. Frozen Architecture

| Layer | Choice |
|-------|--------|
| Target OS | Debian (current stable) |
| Installer language | Bash (modules); Python only if owner approves a specific need |
| TUI | whiptail (or agreed equivalent) |
| VPN engine | **sing-box** only (VLESS+Reality, Hysteria2) |
| OpenWrt control | **OpenSOHO** (agent + server model for NAT) |
| DNS | **Blocky** + HaGeZi (and similar) lists; hybrid VPS + router fallback |
| Monitoring | **Uptime Kuma** + **Beszel** on the same VPS |
| Control plane UX | Telegram bot + CLI; no mandatory Web UI for DNS/VPN in v1 |
| Backups | restic or borg (owner preference when implementing) |
| Layout | `install.sh`, `modules/`, `configs/`, `docs/` |

## Forbidden (unless the owner explicitly approves)

- Replacing sing-box with Xray (or other) as the default VPN engine.
- Replacing OpenSOHO with OpenWISP or heavy panels without approval.
- Replacing Blocky with AdGuard Home as the default DNS module.
- Adding parental control or Blocky UI as **default** required components.
- Committing **secrets**, private keys, or real VPN/client configs.
- Non-idempotent “run once and hope” installers without clear state checks.
- Scope creep into unrelated self-host stacks (media, full mail, etc.) without ROADMAP.
- Multi-tenant / SaaS redesign without approval.

---

# 5. Technology Policy

| Area | Allowed |
|------|---------|
| OS | Debian |
| Shell | Bash (POSIX-friendly where easy) |
| TUI | whiptail |
| VPN | sing-box upstream releases |
| DNS | Blocky upstream; list URLs (HaGeZi, OISD, etc.) |
| OpenWrt mgmt | OpenSOHO |
| Monitoring | Uptime Kuma, Beszel |
| Notifications | Telegram Bot API |
| Backups | restic / borg |

New major components or languages require owner approval.

---

# 6. AI Agent Rules

1. **Read before writing** — AGENTS, ROADMAP, repo.
2. **Never guess** — ask the owner.
3. **AGENTS.md overrides** personal preference and chat history.
4. **Do not edit AGENTS** without explicit owner consent.
5. **Do not change architecture** without approval.
6. **Do not add dependencies/services** without approval.
7. **Do not do unrequested work**.
8. **Do not optimize/refactor** unless ordered.
9. **Ask** when designs diverge and this doc does not decide.
10. Prefer **small, reviewable commits** and modular changes.

---

# 7. Code Review and Refactoring Policy

Report first (issue, impact, severity, fix). Implement only after owner approval **or** an explicit implement order.

---

# 8. Repository structure (target)

```
FreshVPS/
├── AGENTS.md
├── README.md
├── install.sh              # entrypoint
├── modules/                # one concern per module
│   ├── hardening.sh
│   ├── sing-box.sh
│   ├── blocky.sh
│   ├── openwisp_or_opensoho.sh
│   ├── kuma.sh
│   ├── beszel.sh
│   ├── telegram.sh
│   └── backup.sh
├── configs/                # templates, examples (no secrets)
├── docs/
└── VERSION                 # when versioning starts
```

Names may be refined; modularity must remain.

---

# 9. Git and secrets

- Default branch: **main**
- One logical task → one commit (or small focused series if owner agrees)
- Never commit secrets, `.env` with tokens, private keys, or client UUID lists
- Example configs only under `configs/` with placeholders

---

# 10. Definition of Done (module)

- Installs on clean Debian without manual steps beyond documented prompts
- Idempotent or clearly documented re-run behaviour
- Failures are visible; no silent partial success
- Documented enable/disable and key paths
- No secrets in the repo

---

# 11. Decision priority

**AGENTS.md wins** over chat history and agent preference.

---

# 12. Project State (mutable)

| Area | Status |
|------|--------|
| Category | **Production** goal (self-hosted, single-operator) |
| Architecture | **Agreed** for v1 stack; installer code not yet complete |
| VPN | sing-box (VLESS+Reality, Hysteria2) |
| DNS | Blocky + HaGeZi; hybrid with routers |
| OpenWrt | OpenSOHO |
| Monitoring | Uptime Kuma + Beszel |
| Parental / DNS UI | **Not** in v1 |
| Next | Repository layout + modular `install.sh` skeleton |

---

# 13–14

Mutable: Project State, ROADMAP, CHANGELOG.  
Immutable: core rules; AGENTS edits only with owner consent.

Final: production quality, no unrequested work, modular design, one commit per task, chat never overrides AGENTS.

---

# End of Document
