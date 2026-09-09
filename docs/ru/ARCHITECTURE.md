# Архитектура Netductor

RU · [EN](../ARCHITECTURE.md)

## Имя

**Netductor** = network + conductor. Прежнее имя: FreshVPS (совместимость на время миграции).

## Роль

Личный **control plane** одного оператора:

- Хост (Debian VPS): VPN, DNS, API, Admin, Telegram, реестр edge
- Клиенты: VLESS/HY2
- Точки: OpenWrt с **исходящим** агентом (за NAT)

## Плоскости (planes)

host · vpn · dns · core · edge · operator · extras (optional)

## Бинарники (цель)

`netductor` · `netductor-agent` · `netductor-tg` — поставка из **GitHub Releases**.

## Миграция

Bash FreshVPS работает → Go-обёртка → перенос API/агента → отключение bash по плоскостям → `/opt/netductor`.

Подробности: [ROADMAP.md](../ROADMAP.md), фаза G.
