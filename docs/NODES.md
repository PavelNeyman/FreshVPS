# Node naming & registry / Именование нод и реестр

## Name format / Формат имени

```
nd-<role>-<marker>
```

| Part / Часть | Meaning / Смысл | Examples / Примеры |
|--------------|-----------------|--------------------|
| `nd` | project prefix / префикс проекта | always / всегда |
| `role` | what the node does / роль | `core` main VPS, `edge` extra exit, `lab` test |
| `marker` | which one / какой именно | `nl01`, `de02`, `11870` (from IP), `home` |

**Examples / Примеры**

| Name | Meaning |
|------|---------|
| `nd-core-nl01` | primary control plane, Netherlands #1 |
| `nd-core-11870` | auto from public IP `x.x.118.70` |
| `nd-edge-de02` | second VPN exit in Germany |
| `nd-lab-01` | scratch / test VPS |
| `mt-home-01` / `nd-mt-home01` | MikroTik site |
| `ow-cudy-tr1` | OpenWrt site |

**Rules / Правила**

- lowercase letters, digits, hyphen only (`a-z`, `0-9`, `-`)
- max ~32 characters
- unique in the fleet registry
- not a public DNS FQDN (optional DNS CNAME can point later)

**How to set / Как задать**

1. Env at install: `NETDUCTOR_HOSTNAME=nd-core-nl01 netductor install`
2. TUI → **Set hostname** (optional manual form)
3. Existing `/etc/netductor/node_id`
4. Auto: `nd-core-<suffix from public IP>`

## Bidirectional registry / Двусторонний реестр

```
Operator (Admin / TG / CLI)          Device (VPS / OpenWrt / MikroTik)
         |                                      |
         |  desired_hostname = nd-core-nl01     |
         |------------------------------------->|  next poll / TUI apply
         |                                      |  applies hostname
         |  heartbeat / self-register           |
         |<-------------------------------------|
         |  registry.hostname updated           |
         |  desired cleared                     |
```

- **Device → registry:** install or heartbeat reports current hostname + IP.
- **Registry → device:** operator sets desired name; device applies and confirms.

Storage: `/var/lib/netductor/nodes/registry.json`

## API & CLI

- `GET /api/nodes` — list (session)
- `POST /api/nodes/self` — register this VPS
- `POST /api/nodes/hostname` — `{"id":"…","hostname":"nd-core-nl01"}`
- `netductor nodes list`
- `netductor nodes rename <id> <hostname>`
