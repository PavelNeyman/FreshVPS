# G7 миграция путей (FreshVPS → Netductor)

**Стратегия:** dual layout + symlink’и, без обязательного переноса данных.

| Legacy | Алиас Netductor |
|--------|-----------------|
| `/opt/freshvps` | `/opt/netductor` |
| `/etc/freshvps` | `/etc/netductor` |
| `/var/lib/freshvps` | `/var/lib/netductor` |

Переопределение: `NETDUCTOR_ETC`, `NETDUCTOR_STATE`, `NETDUCTOR_ROOT`, …

Подробности: [EN](../MIGRATION-G7.md).
