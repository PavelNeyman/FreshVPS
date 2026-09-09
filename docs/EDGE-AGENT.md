# FreshVPS Edge Agent (OpenWrt)

## Why not OpenSOHO as-is

OpenSOHO reuses **openwisp-config / openwisp-monitoring** and focuses on:

| OpenSOHO | We need |
|----------|---------|
| Wi‑Fi SSID / radio / AP assignment | optional later |
| Device health check-in | **yes** |
| Wi‑Fi clients list | **yes** (basic count first) |
| VLAN / DSA ports | later |
| Templates / multi-tenant | no |
| Firmware OTA | later (manual ASU) |
| Works behind NAT without inbound | **must** (outbound agent) |

OpenSOHO assumes openwisp daemons + controller UI. For city APs behind NAT we want a **thin outbound agent** + our Admin API.

## MVP (implemented)

**Agent → VPS (HTTPS/HTTP):**

- Heartbeat: hostname, uptime, load, mem, OpenWrt version, board, WAN IP, wifi client count  
- Command poll (allowlist): `ping`, `uci_get`, `uci_show`, `wifi_reload`, `network_reload`, `logread`, `reboot`  
- Shared `edge_token` (Bearer)

**Not in MVP:** full LuCI remote, arbitrary shell, firmware flash, SSID wizard.

## VPS

```bash
# install module (or enable in install checklist later)
FRESHVPS_ROOT=/opt/freshvps bash -c 'source modules/edge-hub.sh; module_edge_hub_install'
cat /etc/freshvps/secrets/edge_token
```

API:

- `POST /api/edge/heartbeat` (edge token)
- `GET /api/edge/commands?device_id=...` (edge token)
- `POST /api/edge/cmd_result` (edge token)
- `GET /api/edge/devices` (operator session)
- `POST /api/edge/cmd` `{"device_id","action","arg"}` (operator session)

## Router

See `edge/openwrt/INSTALL.md`.

## Cloudflare Tunnel (HTTPS without opening 443)

1. Cloudflare Zero Trust → Tunnels → Create  
2. Install token on VPS: `CLOUDFLARE_TUNNEL_TOKEN=...` + module `cloudflared`  
3. Public hostname → service `http://127.0.0.1:8787`  
4. Agent `SERVER=https://your-hostname`  

Admin and edge share the same tunnel origin; protect with tokens/sessions.

## Reachability without Cloudflare

```bash
freshvps-vpn api-bind detect --ufw
```

Prefer Cloudflare Tunnel when you have a domain.
