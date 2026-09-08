# FreshVPS Admin UI

Built-in operator panel (desktop + mobile) served by the admin API.

## URL

`http://127.0.0.1:8787/admin/`

Only reachable when you can reach the API (localhost, SSH tunnel, or VPN bind).

## Login

1. `sudo freshvps-vpn session 72` or Telegram **API session**
2. Paste token into the login form (stored in `sessionStorage`)

## Features (v0.5)

- Overview: CPU / RAM / disk / load, systemd services, docker containers
- VPN users: list, add, enable/disable/revoke, subscription + QR
- Metrics history: local JSONL collector (1/min), no Prometheus required

## Metrics design

- **Default:** self-collection via `freshvps-metrics.timer` → `/var/lib/freshvps/metrics/history.jsonl`
- **Not bundled:** full Prometheus + Grafana (optional later if you want long retention / PromQL)
- API: `GET /api/status`, `GET /api/metrics`, `GET /api/metrics/history`

## vs Kuma / Beszel

Default install **no longer** enables Uptime Kuma or Beszel. Use this admin for host/VPN health; add Kuma only if you still want external HTTP probes UI.

OpenSOHO stays a separate optional module for OpenWrt.
