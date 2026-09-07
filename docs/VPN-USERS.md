# VPN users (English)

**RU:** [ru/VPN-USERS.md](ru/VPN-USERS.md)

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

HY2 client links use **`insecure=1`** (self-signed cert). Prefer VLESS+Reality when possible. See [SECURITY.md](SECURITY.md).

## CLI

```bash
freshvps-vpn add alice "phone"
freshvps-vpn list
freshvps-vpn link alice           # subscription (VLESS + HY2)
freshvps-vpn link alice vless     # VLESS only
freshvps-vpn link alice hy2       # Hysteria2 only
freshvps-vpn disable|enable|revoke alice
freshvps-vpn session 72
```

## Telegram / API

Same operations via Go bot or localhost API + **session** token (`freshvps-vpn session`). Prefer VPN or SSH tunnel for API. API `/vpn/users/.../link` returns `subscription` (and separate `vless` / `hy2` fields).
