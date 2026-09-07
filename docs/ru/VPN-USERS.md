# Пользователи VPN (русский)

Только оператор выдаёт доступ. Саморегистрации нет.

## Протоколы на пользователя

| Файл | Содержимое |
|------|------------|
| `link-vless.txt` / `qr.png` | VLESS + Reality |
| `link-hy2.txt` | Hysteria2 (свой пароль) |
| `subscription.txt` | **Обе** строки (удобная выдача) |
| `subscription.b64` | Base64 подписки |
| `link.txt` | = VLESS (совместимость) |

Одного стандартного URI «VLESS+HY2 сразу» не существует. Для приложений с подпиской/мультистрокой — `subscription.txt`.

## CLI

```bash
freshvps-vpn add alice "телефон"
cat /etc/freshvps/clients/alice/subscription.txt
```

См. также английскую версию: [../VPN-USERS.md](../VPN-USERS.md)
