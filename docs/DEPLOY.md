# Deploy & test netductor (hand guide)

## 1. Requirements

- Debian 12/13 VPS (root SSH)
- Public IPv4
- Optional: Telegram bot token + your numeric user id
- Optional: domain later (TLS); not required for loopback API

## 2. Bootstrap (one-liner)

```bash
# as root on the VPS
wget -qO /usr/local/bin/netductor \
  https://github.com/PavelNeyman/netductor/releases/download/v0.7.0-dev/netductor-linux-amd64
chmod 755 /usr/local/bin/netductor
netductor version
```

Telegram (optional, before install):

```bash
mkdir -p /etc/netductor/secrets
echo 'BOT_TOKEN' > /etc/netductor/secrets/telegram_bot_token
echo 'YOUR_TG_USER_ID' > /etc/netductor/secrets/telegram_admin_id
chmod 600 /etc/netductor/secrets/*
```

Full install:

```bash
netductor install
# optional media addon (localhost only):
# NETDUCTOR_LAMPAC=1 netductor install lampac
# or: netductor install lampac
```

## 3. Verify

```bash
netductor doctor
systemctl is-active sing-box blocky netductor-api netductor-telegram-bot
netductor vpn list
```

Expect doctor **ok**, services **active**, at least user `operator`.


Default Reality SNI is **ya.ru** (RU whitelist). Mobile often needs a **RU relay VPS** — [VPN-USERS.md](VPN-USERS.md).

```bash
netductor vpn set-sni ya.ru
```

## 4. VPN client

```bash
netductor vpn link operator          # subscription (VLESS + HY2)
netductor vpn link operator vless
netductor vpn link operator hy2
# QR PNG: /etc/netductor/clients/operator/qr.png
```

Import into v2rayN / Streisand / Hiddify / sing-box.

## 5. Telegram bot

1. Open your bot → `/start` (must be the admin id from secrets).
2. **VPN → list** → tap **🟢 operator** (or **Link / QR** and type `operator`).
3. You get VLESS + Hysteria2 as mono text + **QR photo**.
4. Long-press the mono link to copy (Telegram cannot put long URLs into “Copy” buttons >256 chars).

**Addons → Lampac**: status only; UI is `http://127.0.0.1:9118` via SSH tunnel / VPN path to loopback.

## 6. Admin SPA

```bash
# session token
netductor vpn session 72
# SSH tunnel
ssh -L 8787:127.0.0.1:8787 root@VPS
```

Browser: `http://127.0.0.1:8787/admin/` — paste token.

## 7. Update

```bash
systemctl stop netductor-api netductor-telegram-bot
wget -qO /usr/local/bin/netductor https://github.com/PavelNeyman/netductor/releases/download/v0.7.0-dev/netductor-linux-amd64
wget -qO /opt/netductor/bin/netductor-tg https://github.com/PavelNeyman/netductor/releases/download/v0.7.0-dev/netductor-tg-linux-amd64
chmod 755 /usr/local/bin/netductor /opt/netductor/bin/netductor-tg
systemctl start netductor-api netductor-telegram-bot
netductor doctor
```

Optional integrity: `NETDUCTOR_UPDATE_SHA256=<hex> netductor update …`

## 8. Uninstall / purge

```bash
netductor uninstall   # units/binaries; keep configs
netductor purge       # also configs/data (destructive)
```

## 9. Security notes

- API listens on `127.0.0.1:8787` by default
- Lampac on `127.0.0.1:9118` only
- Do not expose API without TLS + `NETDUCTOR_API_PUBLIC=1`
