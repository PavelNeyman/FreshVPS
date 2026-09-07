# Upgrade (русский)

**EN:** [../UPGRADE.md](../UPGRADE.md)

Только **чистые** установки; миграций со старых версий нет — проще пересоздать VPS.

```bash
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh | sudo bash -s -- --upgrade
```

Флаги: `--force` / `--force-module`. Конфиг модулей: `/etc/freshvps/install.conf`. Пароли SSH не отключаются без ключа в `authorized_keys`.
