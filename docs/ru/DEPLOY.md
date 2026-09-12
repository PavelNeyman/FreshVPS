# Развёртывание и ручной тест netductor

## 1. Требования

- VPS Debian 12/13 (root по SSH)
- Публичный IPv4
- Опционально: токен Telegram-бота и ваш numeric user id

## 2. Bootstrap

```bash
wget -qO /usr/local/bin/netductor \
  https://github.com/PavelNeyman/netductor/releases/download/v0.7.0-dev/netductor-linux-amd64
chmod 755 /usr/local/bin/netductor
netductor version
```

Telegram (до install):

```bash
mkdir -p /etc/netductor/secrets
echo 'BOT_TOKEN' > /etc/netductor/secrets/telegram_bot_token
echo 'ВАШ_TG_ID' > /etc/netductor/secrets/telegram_admin_id
chmod 600 /etc/netductor/secrets/*
```

```bash
netductor install
# аддон Lampac (только localhost):
# netductor install lampac
```

## 3. Проверка

```bash
netductor doctor
systemctl is-active sing-box blocky netductor-api netductor-telegram-bot
netductor vpn list
```

## 4. VPN

```bash
netductor vpn link operator
netductor vpn link operator vless
netductor vpn link operator hy2
# QR: /etc/netductor/clients/operator/qr.png
```

## 5. Telegram

1. `/start` (только admin id из secrets).
2. **VPN → список** → нажать **🟢 operator** (или **Ссылка / QR** и ввести `operator`).
3. Придут VLESS + Hysteria2 (mono) и **фото QR**.
4. Длинное нажатие на mono-текст — копирование (кнопка Copy не вмещает длинные URI >256 символов).

**Аддоны → Lampac**: статус; UI `http://127.0.0.1:9118` через SSH-туннель.

## 6. Admin SPA

```bash
netductor vpn session 72
ssh -L 8787:127.0.0.1:8787 root@VPS
```

Открыть `http://127.0.0.1:8787/admin/`, вставить токен.

## 7. Обновление

```bash
systemctl stop netductor-api netductor-telegram-bot
wget -qO /usr/local/bin/netductor https://github.com/PavelNeyman/netductor/releases/download/v0.7.0-dev/netductor-linux-amd64
wget -qO /opt/netductor/bin/netductor-tg https://github.com/PavelNeyman/netductor/releases/download/v0.7.0-dev/netductor-tg-linux-amd64
chmod 755 /usr/local/bin/netductor /opt/netductor/bin/netductor-tg
systemctl start netductor-api netductor-telegram-bot
```

## 8. Удаление

```bash
netductor uninstall   # без конфигов
netductor purge       # полностью
```
