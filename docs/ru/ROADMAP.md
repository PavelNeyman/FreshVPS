# Roadmap Netductor

Прежнее имя: FreshVPS. **EN:** [../ROADMAP.md](../ROADMAP.md)

## Сделано (FreshVPS 0.5.x–0.6.x)

- Admin UI, метрики/пробы, алерты TG
- Multi-user VPN, бот, session API
- Edge hub + shell-агент OpenWrt
- Release `freshvps-tg`

## Фаза G — Netductor + Go (активна)

Имя **Netductor** зафиксировано. Bash живёт, пока плоскость не перенесена в Go.

| Шаг | Результат |
|-----|-----------|
| G0 | AGENTS 2.0, ARCHITECTURE, roadmap |
| G1 | Каркас `cmd/netductor` |
| G2 | Releases + compat `freshvps-*` |
| G3 | `netductor serve` (API на Go) |
| G4 | `install/upgrade` по planes |
| G5 | `netductor-agent` |
| G6 | TG / опционально rename репо |
| G7 | `/opt/netductor`, deprecate bash |

## Позже / вне скоупа

Mini App, трафик per-user, полный паритет OpenSOHO, Prometheus — см. EN roadmap.
