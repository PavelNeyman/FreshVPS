# VPN users (English)

Operator-only control plane. Users never self-register.

## Protocols per user

| Artifact | Content |
|----------|---------|
| `link-vless.txt` / `qr.png` | VLESS + Reality |
| `link-hy2.txt` | Hysteria2 (own password) |
| `subscription.txt` | **Both** lines (recommended handoff) |
| `subscription.b64` | Base64 of subscription |
| `link.txt` | Same as VLESS (compat) |

There is **no** single standard URI that carries VLESS and HY2 together. Apps that support multi-line / subscription import should use `subscription.txt`.

## CLI

```bash
freshvps-vpn add alice "phone"
freshvps-vpn list
freshvps-vpn link alice          # prints VLESS by default
cat /etc/freshvps/clients/alice/subscription.txt
freshvps-vpn disable|enable|revoke alice
freshvps-vpn session 72
```

## Telegram / API

Same operations via Go bot or localhost API + session token. Prefer VPN or SSH tunnel for API.

See also: [ru/VPN-USERS.md](ru/VPN-USERS.md)
