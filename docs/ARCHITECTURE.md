# Netductor architecture

EN · [RU](ru/ARCHITECTURE.md)

## Name

**Netductor** = network + conductor. Former name: FreshVPS (compat during migration).

## Role

Personal **control plane** for one operator:

- Control-plane host (Debian VPS): VPN, DNS, API, Admin, Telegram, edge registry
- Clients: phones/laptops via VLESS/HY2
- Sites: OpenWrt routers with **outbound** agent (NAT-friendly)

## Planes

| Plane | Responsibility |
|-------|----------------|
| host | OS packages, hardening, swap, timers |
| vpn | sing-box, users, links, subscriptions |
| dns | Blocky |
| core | HTTP API, sessions, metrics/probes, Admin UI |
| edge | Device registry, commands, agent protocol |
| operator | CLI, TUI, Telegram |
| extras | OpenSOHO, cloudflared, Lampac, … (optional) |

## Binaries (target)

| Binary | Where |
|--------|--------|
| `netductor` | VPS / operator laptop (install, serve, vpn, edge, doctor) |
| `netductor-agent` | OpenWrt |
| `netductor-tg` | VPS (or `netductor telegram`) |

Install path long-term: **GitHub Release** asset, not full git clone.

## Trust boundaries

- Admin API: session token (operator); bind may be localhost, VPN IP, or `0.0.0.0` with token auth
- Edge API: shared `edge_token`; agents only dial out
- No arbitrary remote shell in MVP agent (allowlist only)

## Migration

1. Keep FreshVPS bash working
2. Add Go CLI that shells out to existing tools
3. Port API → Go `serve`
4. Port agent → Go
5. Retire bash modules plane by plane
6. Rename paths `/opt/netductor` when cut over

See [ROADMAP.md](ROADMAP.md) Phase G.
