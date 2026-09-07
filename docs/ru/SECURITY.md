# Замечания по безопасности (русский)

**EN:** [../SECURITY.md](../SECURITY.md)

## SSH

Парольный вход отключается **только** при наличии `authorized_keys` (или `FRESHVPS_ALLOW_PASSWORD_SSH=1`).

## VPN

- **VLESS + Reality** — предпочтительный путь клиента.
- **Hysteria2** использует **self-signed** сертификат; в клиентских ссылках `insecure=1` для удобства. Аутентичность слабее, чем у Reality. По возможности выбирайте VLESS.

## Admin API

- По умолчанию bind `127.0.0.1`.
- Auth: только **Bearer session** из `freshvps-vpn session [hours]` (отдельного master token нет).
- Предпочтительно SSH-туннель или доступ только через ваш VPN.

## Тесты VPS

`freshvps-tests` может выполнять сторонние скрипты по сети; используйте на тестовой VPS или принимайте модель доверия.
