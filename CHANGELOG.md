# Changelog

## 0.4.2 — 2026-09-07

Fixes from live VPS smoke (Debian 13):

- **blocky:** wait until DNS answers on 127.0.0.1 before rewriting `resolv.conf` (avoids apt DNS race right after install)
- **doctor:** fix `set -o pipefail` breakage on container checks; detect API unit without `grep -q` race; check **lampac** when present
- **lampac:** `init.conf` always writable/readable by container UID 1000; force `chromium.enable=false` (root-only 600 caused silent Chromium ON ~1GB RAM)
- **docs/BOOTSTRAP:** note that private repos need PAT/`git clone` (raw/codeload 404)

## 0.4.1 — 2026-09-07

P0 from deep review:

- Bootstrap: no broken sudo re-exec from pipe; Debian-only auto vps; document file install
- Host tools module: doctor + smoke survive tree delete (`freshvps-smoke`)
- API: subscription-first JSON; name sanitize; no dead master token; rebind detect fix
- Doctor: FAIL on zero enabled VPN users
- WARN if sing-box without vpn-users; HY2 note in READY/SECURITY
- Test plans: `docs/TEST-PLAN-HUMAN.md`, `docs/TEST-PLAN-AI.md`

## 0.4.0 — greenfield hardening, install.conf, shared sing-box install
## 0.3.x — panels, tests menu, idempotent upgrade
