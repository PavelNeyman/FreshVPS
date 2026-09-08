# FreshVPS Admin UI

Built-in operator panel (desktop + mobile) on the admin API. No Prometheus.

## URL

`http://127.0.0.1:8787/admin/` (`/admin` redirects here)

Reachable via localhost, SSH tunnel, or VPN bind (`freshvps-vpn api-bind`).

## Login

1. `sudo freshvps-vpn session 72` or Telegram **API session** / `/session`
2. Paste token (kept in `sessionStorage`)
3. Language: **RU/EN** toggle in header

## Tabs

| Tab | Content |
|-----|---------|
| Overview | CPU/RAM/disk/load, net RX/TX, services, containers, probe strip, collector stale warning |
| VPN | users search, add, note, enable/disable/revoke, subscription + VLESS + HY2 copy, QR |
| Metrics | CPU & RAM history chart |
| Probes | live status, latency, 24h uptime % |
| Settings | alert thresholds + probes JSON → `/etc/freshvps/probes.json` |

## Metrics & alerts

- Timer `freshvps-metrics` every minute → JSONL + Telegram alerts / recovery
- Config: `/etc/freshvps/probes.json`
- HY2 probe must be **UDP** (not TCP)

## Telegram

Bot: **Admin UI** button or `/admin` — tunnel + URL hints.

## Security

API default bind `127.0.0.1`. Prefer tunnel or VPN interface; avoid public `0.0.0.0` without extra auth.
