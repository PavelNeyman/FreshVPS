# Node naming & registry

## Name format

```
nd-<role>-<marker>
```

| Part | Meaning | Examples |
|------|---------|----------|
| `nd` | fixed prefix (netductor) | always |
| `role` | function of the node | `core` main VPS, `edge` extra exit, `lab` test |
| `marker` | unique short id | `nl01`, `11870` (from IP), `home` |

**Examples**

- `nd-core-nl01` — primary control plane in Netherlands  
- `nd-core-11870` — auto from public IP `x.x.118.70`  
- `nd-edge-de02` — second VPN exit in Germany  
- `nd-lab-01` — scratch VPS  

**OpenWrt / MikroTik** use the same idea, different `kind`:

- `mt-home-01` or `nd-mt-home01` (MikroTik)  
- `ow-cudy-tr1` (OpenWrt site)

Rules: lowercase, digits, hyphen only; max ~32 chars; unique in the registry.

## How rename works (bidirectional)

```
Operator (admin/TG)                Device
       |                              |
       |  desired_hostname=nd-core-nl01
       |----------------------------->|  (next poll / install)
       |                              |  applies hostname
       |  heartbeat hostname=...      |
       |<-----------------------------|
       |  registry.hostname updated   |
       |  desired cleared             |
```

- **Device → registry:** install/heartbeat reports current hostname + IP → `UpsertFromDevice`.  
- **Registry → device:** operator sets `desired_hostname` → device applies on next cycle → reports back.

## Storage

`/var/lib/netductor/nodes/registry.json`

## API (operator session)

- `GET /api/nodes` — list  
- `POST /api/nodes/self` — local VPS register  
- `POST /api/nodes/{id}/hostname` — `{"hostname":"nd-core-nl01"}` set desired  
