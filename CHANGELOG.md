# Changelog

## 0.5.4 — 2026-09-08

- Hardening: blacklist virtio_gpu (headless KVM soft lockup mitigation)
- Hardening: 2G /swapfile by default (FRESHVPS_SWAP_MB, FRESHVPS_SKIP_SWAP=1)
- Kuma/Beszel optional off by default

# Changelog

## 0.5.2 — 2026-09-08

- Admin UI: auto-refresh, probe strip, session expiry, toasts, note edit, search, copy VLESS/HY2
- Metrics: net counters, probe 24h uptime %, Settings for alerts/probes JSON
- Alerts: recovery messages; UDP hy2; flock + cooldown before notify
- API: `/api/session`, `/vpn/users/:name/note`, `/api/probes/uptime`, POST probes config
- CLI: `freshvps-vpn note`
- Doctor: admin UI + metrics timer checks

## 0.5.1 — probes + charts + UDP hy2 fix
## 0.5.0 — Admin UI + self metrics (no Prometheus); Kuma/Beszel off by default
## 0.4.x — install/prepare, telegram bot, uninstall tiers
