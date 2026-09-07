# Установка FreshVPS (русский)

**EN:** [../INSTALL.md](../INSTALL.md)

Готовая система: multi-user VPN (**VLESS+Reality** + **Hysteria2**), Blocky, мониторинг, CLI / Telegram (Go) / API.

## Требования

- Debian 12/13, root, исходящий HTTPS

## Роли

| Роль | Команда |
|------|---------|
| VPS | `sudo bash install.sh` |
| Edge SBC | `sudo bash install.sh --role edge-client` |
| OpenWrt | `openwrt/` на роутере, `sh install-openwrt.sh` |

## После установки

`/etc/freshvps/READY.txt` · пользователи: `freshvps-vpn add …` · выдача: **`subscription.txt`** (VLESS+HY2).

Порты: **443** VLESS, **8443** HY2. Панели только localhost + SSH-туннель.

Lampac — только `ENABLE_LAMPAC=1`.
