# Bootstrap, compose, SSH (русский)

**EN:** [../BOOTSTRAP.md](../BOOTSTRAP.md)

## Без `git clone`

```bash
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh | sudo bash
```

С конфигом: `--non-interactive --config /root/freshvps.conf`. Тянется **tar.gz с GitHub**, git не обязателен.

Режимы: `vps` | `edge-client` | `openwrt` | `plan` (macOS — только распаковка и подготовка, VPN на Mac не ставится).

## Выбор компонентов

Как и раньше: **не всё сразу**. `ENABLE_*` / TUI / конфиг. Docker-панели — [`compose/panels.yml`](../../compose/panels.yml) с **profiles**. VPN/DNS на хосте.

## TUI

`lib/tui.sh`: whiptail → dialog → osascript (macOS) → обычный `read`. На OpenWrt без лишних пакетов.

## SSH-ключи

Лучше ключи, не пароли. Один operator-ключ — ок для «всё моё»; отдельные ключи VPS/роутеры — меньше радиус поражения. На Mac: **passphrase + ssh-agent**; без passphrase — только если понимаешь риск кражи ноутбука.
