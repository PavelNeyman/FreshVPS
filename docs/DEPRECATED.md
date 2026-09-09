# Deprecated / legacy

Still shipped for compatibility; do not use for new installs.

| Item | Replace with |
|------|----------------|
| `cmd/freshvps-tg` | **removed** → `cmd/netductor-tg` |
| `runtime/telegram/bot.sh` | `netductor-tg` binary |
| `modules/vpn-api.sh` (Python) | `modules/netductor-api.sh` (cutover) or parallel |
| `freshvps-telegram-bot.service` | `netductor-telegram-bot.service` |
| Binary name `freshvps-tg` | `netductor-tg` (symlink may remain) |
| Paths `/opt/freshvps` only | dual `/opt/netductor` (see MIGRATION-G7) |

## Keep (not deprecated)

- Bash modules for host/sing-box/blocky/hardening/openwrt site config
- `bin/freshvps-vpn`, `bin/freshvps-doctor` until native Go ports
- Python `runtime/api/server.py` while `NETDUCTOR_API_MODE=parallel`
- Shell `edge/openwrt/freshvps-agent` as fallback next to Go agent
