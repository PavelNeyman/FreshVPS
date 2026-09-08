# FreshVPS Admin UI (админка)

Встроенная панель оператора (desktop + mobile). Без Prometheus.

## URL

`http://127.0.0.1:8787/admin/`

Localhost, SSH-туннель или bind на VPN (`freshvps-vpn api-bind`).

## Вход

1. `sudo freshvps-vpn session 72` или Telegram **API session**
2. Вставить токен
3. Язык: кнопка **RU/EN**

## Вкладки

Overview · VPN · Metrics · Probes · Settings — метрики, пользователи, пробы, пороги алертов.

## Алерты

`freshvps-metrics.timer` → Telegram при падении и **recovery**.  
hy2 — проба **UDP**.

## Telegram

Кнопка **Admin UI** или `/admin`.
