# VPN-пользователи (русский)

**EN:** [../VPN-USERS.md](../VPN-USERS.md)

Управление только оператором. Пользователи сами не регистрируются.

## Протоколы на пользователя

| Артефакт | Содержимое |
|----------|------------|
| `link-vless.txt` / `qr.png` | VLESS + Reality |
| `link-hy2.txt` | Hysteria2 (свой пароль) |
| `subscription.txt` | **Обе** строки (рекомендуемая выдача) |
| `subscription.b64` | Base64 subscription |
| `link.txt` | То же, что VLESS (совместимость) |

Одного стандартного URI сразу с VLESS и HY2 **нет**. Приложения с multi-line / subscription — через `subscription.txt`.

В ссылках HY2 стоит **`insecure=1`** (self-signed). По возможности предпочитайте VLESS+Reality. См. [SECURITY.md](SECURITY.md).

## CLI

```bash
freshvps-vpn add alice "phone"
freshvps-vpn list
freshvps-vpn link alice           # subscription (VLESS + HY2)
freshvps-vpn link alice vless
freshvps-vpn link alice hy2
freshvps-vpn disable|enable|revoke alice
freshvps-vpn session 72
```

## Telegram / API

Те же операции через Go-бота или localhost API + **session** (`freshvps-vpn session`). API лучше через VPN или SSH-туннель. Ответ `/vpn/users/.../link` содержит `subscription`, а также `vless` / `hy2`.
