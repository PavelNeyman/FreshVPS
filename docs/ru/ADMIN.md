# FreshVPS Admin UI (админка)

Встроенная панель оператора (desktop + mobile) на том же API.

## URL

`http://127.0.0.1:8787/admin/`

Доступна, если вы дотягиваетесь до API (localhost, SSH-туннель или bind на VPN).

## Вход

1. `sudo freshvps-vpn session 72` или Telegram **API session**
2. Вставить токен в форму (`sessionStorage`)

## Возможности

- Overview: CPU / RAM / диск / load, systemd, docker
- VPN users: список, add/enable/disable/revoke, subscription + QR
- Метрики: локальный JSONL раз в минуту

## Метрики (без Prometheus)

- Collector: `freshvps-metrics.timer` → `/var/lib/freshvps/metrics/history.jsonl`
- Live: `GET /api/status`, `GET /api/metrics`
- История: `GET /api/metrics/history`
- В стеке **нет** Prometheus, Grafana, node_exporter

## Kuma / Beszel

По умолчанию **выключены**. OpenSOHO — отдельный optional модуль.
