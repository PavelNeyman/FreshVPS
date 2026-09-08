# FreshVPS Admin UI (админка)

Встроенная панель оператора (desktop + mobile) на том же API.

## URL

`http://127.0.0.1:8787/admin/`

Доступна только если вы дотягиваетесь до API (localhost, SSH-туннель или bind на VPN).

## Вход

1. `sudo freshvps-vpn session 72` или Telegram **API session**
2. Вставить токен в форму (хранится в `sessionStorage`)

## Возможности (v0.5)

- Overview: CPU / RAM / диск / load, systemd, docker
- VPN users: список, add/enable/disable/revoke, subscription + QR
- Metrics: локальный JSONL (раз в минуту), **без** Prometheus

## Метрики

- По умолчанию свой collector (`freshvps-metrics.timer`)
- Prometheus/Grafana не ставятся — при необходимости позже как optional
- API: `/api/status`, `/api/metrics`, `/api/metrics/history`

## Kuma / Beszel

По умолчанию **выключены**. OpenSOHO — по-прежнему отдельный optional модуль.
