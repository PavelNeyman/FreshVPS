# G7 path migration (FreshVPS → Netductor)

## Strategy

**Dual layout with symlinks** — no forced data move on existing hosts.

| Legacy | Netductor alias |
|--------|-----------------|
| `/opt/freshvps` | `/opt/netductor` → symlink |
| `/etc/freshvps` | `/etc/netductor` → symlink |
| `/var/lib/freshvps` | `/var/lib/netductor` → symlink |
| `freshvps-*` CLI | still work; `netductor` is primary |
| `freshvps-telegram-bot` | `netductor-telegram-bot` |

Go code resolves via `internal/paths`: env override → netductor path → freshvps path.

## Env overrides

```
NETDUCTOR_ETC  NETDUCTOR_STATE  NETDUCTOR_ROOT
NETDUCTOR_SESSIONS  NETDUCTOR_CLIENTS  NETDUCTOR_EDGE_DIR
NETDUCTOR_METRICS_DIR  NETDUCTOR_ADMIN_ROOT  NETDUCTOR_VPN_BIN
```

## Full cutover (optional, later)

1. Copy trees instead of symlinks.
2. Point systemd units at `/opt/netductor`.
3. Retire `freshvps-*` unit names.
