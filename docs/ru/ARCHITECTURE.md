# Архитектура

Один Go-бинарь **netductor** на VPS:

| Плоскость | Роль |
|-----------|------|
| install | пакеты + systemd |
| serve | Admin API + UI :8787 |
| vpn | VLESS Reality + HY2 |
| collect / probes | метрики и алерты в Telegram |
| doctor / status | здоровье |
| tui | Bubble Tea + Huh |

**netductor-agent** на OpenWrt: outbound heartbeat.  
**netductor-tg**: бот оператора.

Без Python control plane. Без shell-модулей установки.
