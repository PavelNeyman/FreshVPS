# Обновление и повторный запуск (русский)

**EN:** [../UPGRADE.md](../UPGRADE.md)

Проект рассчитан на **чистые** установки (VPS при необходимости проще пересоздать). Отдельного миграционного тулкита нет.

## Повтор / upgrade

```bash
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh -o /tmp/fv.sh
sudo bash /tmp/fv.sh --upgrade
```

- Здоровые модули **пропускаются** (`lib/idempotent.sh`)
- Выбор модулей берётся из `/etc/freshvps/install.conf`, если файл есть
- Секреты, `vpn-users.json`, volume панелей **сохраняются**
- UFW **никогда** не сбрасывается (`ufw --force reset` не вызывается)

### Принудительно

```bash
sudo bash install.sh --force --non-interactive
sudo bash install.sh --force-module blocky --upgrade
```

### SSH и lockout

PasswordAuthentication отключается **только** если есть `authorized_keys` (или `FRESHVPS_ALLOW_PASSWORD_SSH=1`).

### После установки

```bash
sudo freshvps-doctor
sudo freshvps-smoke
```
