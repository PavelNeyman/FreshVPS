# Установка FreshVPS (русский)

**EN:** [../INSTALL.md](../INSTALL.md)

На выходе — **готовая к использованию** система: multi-user VPN (VLESS+Reality + Hysteria2), Blocky, опциональные панели, CLI / Telegram / API оператора.

## Требования

- Debian 12/13 (лучше чистая VPS)
- Root; исходящий HTTPS (GitHub, Docker Hub / зеркала)
- **SSH-ключ** в `authorized_keys` до установки (парольный вход отключается, если ключ есть)

## Рекомендуемая установка (bootstrap)

GitHub отдаёт живой `tar.gz` репозитория (Actions не нужны).

```bash
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh -o /tmp/fv.sh
sudo bash /tmp/fv.sh
```

Также: `curl … | sudo bash` (предпочтительнее файл).

Обновление: `sudo bash /tmp/fv.sh --upgrade`

## Роли

| Роль | Как |
|------|-----|
| VPS (по умолчанию на Debian) | bootstrap / `install.sh --role vps` |
| Edge SBC | `install.sh --role edge-client` или bootstrap `--mode edge-client` |
| OpenWrt | `openwrt/install-openwrt.sh` **на роутере** |
| macOS control plane | bootstrap `--mode plan` (VPN на Mac не ставится) |

## Неинтерактивно

```bash
cp configs/freshvps.conf.example /root/freshvps.conf
# PUBLIC_IP, TELEGRAM_*, ENABLE_*
sudo bash /tmp/fv.sh --non-interactive --config /root/freshvps.conf
```

## После установки

Читайте **`/etc/freshvps/READY.txt`**.

| Путь / команда | Назначение |
|----------------|------------|
| `/etc/freshvps/clients/<name>/subscription.txt` | VLESS + HY2 (рекомендуемая выдача) |
| `freshvps-vpn link <name>` | Печатает subscription по умолчанию |
| `freshvps-doctor` / `freshvps-smoke` | Проверки |
| `freshvps-tests` | Опциональные сетевые пробы |

Порты: **443** TCP VLESS+Reality, **8443** UDP Hysteria2. DNS клиентов: лучше remote/server (Blocky на VPS).

### VPN-пользователи

```bash
freshvps-vpn add alice "phone"
freshvps-vpn link alice
freshvps-vpn list
freshvps-vpn session 72
```

См. [VPN-USERS.md](VPN-USERS.md).

### Telegram

Только оператор: `/status` `/vpn_add` `/vpn_list` `/vpn_link` `/vpn_disable` `/vpn_enable` `/vpn_revoke` `/session` `/ready`

### Admin API

По умолчанию `127.0.0.1:8787`. Auth: Bearer **session** из `freshvps-vpn session`. См. [SHORTCUT-IOS.md](SHORTCUT-IOS.md).

### OpenWrt (на роутере)

```bash
scp -r openwrt root@ROUTER:/root/freshvps-openwrt
# site.conf: WIFI_PASSWORD, LAN; WAN_DEVICE=auto или eth0.2 (Cudy TR1200)
ssh root@ROUTER 'cd /root/freshvps-openwrt && sh install-openwrt.sh'
```

### Панели (localhost)

```bash
ssh -L 8090:127.0.0.1:8090 -L 3001:127.0.0.1:3001 -L 8091:127.0.0.1:8091 root@VPS
```

Lampac **опционален** (`ENABLE_LAMPAC=1`).

## Удаление

`sudo bash uninstall.sh` · `sudo bash uninstall.sh --purge-secrets`
