# Edge agent — enrollment

Outbound only (NAT). No management VPN required.

## Tokens

| Secret | Role |
|--------|------|
| `/etc/netductor/secrets/edge_bootstrap_token` | enroll only |
| per-device `device_token` (after approve) | heartbeat, commands, backup |
| `edge_token` | operator/global (legacy) |

## Flow

1. Agent starts with bootstrap token in config `TOKEN=…`
2. `POST /api/edge/enroll` → `pending`
3. Operator notified (Telegram) + `netductor edge pending`
4. `netductor edge approve site1` → device token
5. Agent stores `/etc/netductor-agent/device_token` and uses it
6. `deny` / `revoke` cuts access

```bash
netductor edge pending
netductor edge approve site1
netductor edge deny site1
netductor edge revoke site1
netductor edge list
netductor edge cmd site1 config_backup
```

Templates / provision (SSH install agent only) — next.
