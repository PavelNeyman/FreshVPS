# FreshVPS Admin UI

Built-in operator panel (desktop + mobile) served by the admin API.

## URL

`http://127.0.0.1:8787/admin/`

Only reachable when you can reach the API (localhost, SSH tunnel, or VPN bind).

## Login

1. `sudo freshvps-vpn session 72` or Telegram **API session**
2. Paste token into the login form (stored in `sessionStorage`)

## Features

- Overview: CPU / RAM / disk / load, systemd services, docker containers
- VPN users: list, add, enable/disable/revoke, subscription + QR
- Metrics history: local JSONL collector (1/min)

## Metrics (no Prometheus)

- Collector: `freshvps-metrics.timer` → `/var/lib/freshvps/metrics/history.jsonl`
- Live: `GET /api/status`, `GET /api/metrics`
- History: `GET /api/metrics/history`
- No Prometheus, Grafana, or node_exporter in the stack

## vs Kuma / Beszel

Default install does **not** enable Uptime Kuma or Beszel. Use this admin for host/VPN health.

OpenSOHO remains a separate optional module for OpenWrt.
