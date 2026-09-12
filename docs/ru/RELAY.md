# Промежуточный VPS в РФ (relay)

```text
Телефон --VLESS Reality (SNI ya.ru)--> RU relay --VLESS Reality--> зарубежный core --> Интернет
Дом     --VLESS Reality--------------> зарубежный core --> Интернет
```

## Железо и сеть

| | Минимум | Рекомендуется |
|--|---------|----------------|
| CPU | 1 vCPU | 1–2 |
| RAM | **512 МБ** | **1 ГБ** |
| Диск | 10 ГБ | 15–20 ГБ |
| ОС | Debian 12/13 | |
| IP | Публичный **IPv4 в РФ** | **Yandex / VK Cloud, Timeweb, Selectel** |
| Порты | **443/tcp** | Исходящий 443 на core |
| Канал | ~100 Мбит/с | по пользователям |

На relay нужен только **sing-box**.

## Автонастройка

**Core:**

```bash
netductor relay export -o bundle.json --sni ya.ru
scp bundle.json root@RU_VPS:/root/
```

**RU VPS:**

```bash
wget -qO /usr/local/bin/netductor \
  https://github.com/PavelNeyman/netductor/releases/download/v0.7.0-dev/netductor-linux-amd64
chmod 755 /usr/local/bin/netductor
netductor relay join /root/bundle.json
netductor relay links
```

Домашние клиенты — ссылки **core**; мобильные — **relay links**.

## Managed agent (no SSH after join)

After `relay join`, RU runs `netductor-relay-agent`:

- Heartbeat → core `:8788` (`NETDUCTOR_RELAY_API`)
- Auto-pulls user list when core VPN users change
- Core shows status: `netductor relay status` / Admin Relay / TG Relay
- Mobile links: `GET /api/relay/links` or TG (uses last reported IP+pbk)

Open on **core** firewall: **8788/tcp** from the RU IP (or world if needed).

## Provision с core

`netductor relay provision --host IP --user root --password …`

Core ставит свой SSH-ключ, отключает пароль, join + agent.
