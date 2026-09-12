# RU relay VPS

Chain for **mobile whitelist**:

```text
Phone  --VLESS Reality (SNI ya.ru)-->  RU relay  --VLESS Reality-->  foreign core  --> Internet
Home   --VLESS Reality-------------->  foreign core  --> Internet
```

## Hardware / network (RU hop)

| Item | Minimum | Recommended |
|------|---------|-------------|
| CPU | 1 vCPU | 1–2 vCPU |
| RAM | **512 MB** | **1 GB** |
| Disk | 10 GB | 15–20 GB |
| OS | Debian 12/13 | same |
| IP | Public **IPv4 in Russia** | Prefer **Yandex Cloud, VK Cloud, Timeweb, Selectel** (better whitelist chance) |
| Ports | **443/tcp** inbound open | Outbound **443** to foreign core must work |
| Bandwidth | ~100 Mbps | Scale with users |

Relay runs **only sing-box** (no Blocky/API/Telegram/Lampac required).

## Automated setup

### On foreign **core**

```bash
netductor vpn list
netductor relay export -o bundle.json --sni ya.ru
scp bundle.json root@RU_VPS:/root/
```

Creates core user `relay-uplink` (not for humans) and the join file.

### On new **RU** VPS

```bash
wget -qO /usr/local/bin/netductor \
  https://github.com/PavelNeyman/netductor/releases/download/v0.7.0-dev/netductor-linux-amd64
chmod 755 /usr/local/bin/netductor
netductor relay join /root/bundle.json
netductor relay links
systemctl is-active sing-box
```

### Clients

| Network | Where to get the link |
|---------|------------------------|
| Home | Core: `netductor vpn link NAME` |
| Mobile | RU: `netductor relay links` (also `/var/lib/netductor/relay/clients/`) |

Mobile URI uses **RU IP** and **relay Reality pbk/sid**; UUID matches the core user name.

## Re-sync after new VPN users

```bash
# core
netductor relay export -o bundle.json
# RU
netductor relay join /root/bundle.json
```

## Security

- Never distribute `relay-uplink`.
- Treat `bundle.json` as secret (`scp` only).
- Firewall on RU: 443/tcp public; SSH restricted.

## Managed agent (no SSH after join)

After `relay join`, RU runs `netductor-relay-agent`:

- Heartbeat → core `:8788` (`NETDUCTOR_RELAY_API`)
- Auto-pulls user list when core VPN users change
- Core shows status: `netductor relay status` / Admin Relay / TG Relay
- Mobile links: `GET /api/relay/links` or TG (uses last reported IP+pbk)

Open on **core** firewall: **8788/tcp** from the RU IP (or world if needed).
