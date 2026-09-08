# Changelog

## 0.4.4 — 2026-09-08

- **Menu:** `Prepare VPS` action; install flow checklist for components
- **Prepare steps:** apt update/upgrade, base tools, dnsutils, Docker, golang, restic — pick one/many/all
- **Smart prep:** from selected components derive Docker (panels/Lampac), golang only if Telegram and no prebuilt bot, restic if backup
- **Install flow:** install now | prepare then install | prepare only
- **Telegram bot:** prefer `FRESHVPS_TG_BIN` → existing binary → `dist/freshvps-tg-linux-$arch` → GitHub release → `go build` → bash fallback
- **scripts/build-tg-bot.sh:** cross-compile linux/amd64|arm64 from Mac or Linux
- **bootstrap:** `fetch_tree` logs to stderr (fix broken `cd` path)

## 0.4.3 — 2026-09-08

- **uninstall:** two levels
  - default — remove runtime: units, binaries (`sing-box`/`blocky`), CLI, containers, UFW rules; **restore DNS** after blocky; **keep configs + data**
  - `--purge` — also wipe `/etc/freshvps`, `/etc/blocky`, panel data, certs, restic repo, docker images, state
- Modules: lampac in default module list; container `rm` instead of only stop

## 0.4.2 — 2026-09-07

Fixes from live VPS smoke (Debian 13):

- **blocky:** wait until DNS answers on 127.0.0.1 before rewriting `resolv.conf`
- **doctor / lampac / BOOTSTRAP** private-repo note

## 0.4.1 — 2026-09-07 — P0 review fixes
## 0.4.0 — greenfield hardening
## 0.3.x — panels, tests menu, idempotent upgrade
