# Обновление и повторный запуск (русский)

**EN:** [../UPGRADE.md](../UPGRADE.md)

Повторный `install` / `bootstrap` на уже настроенной VPS **не должен** сбрасывать UFW, VPN-пользователей и секреты. Здоровые модули **пропускаются**.

```bash
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh | sudo bash -s -- --upgrade
```

Принудительно: `--force` или `--force-module sing-box`.

OpenWrt: UCI и так сравнивает значения — меняет только отличия.
