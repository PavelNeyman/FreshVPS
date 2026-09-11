# MikroTik edge (thin agent)

RouterOS **cannot** run the Go `netductor-agent` without container support.
This path keeps the **same control plane** (enroll → approve → commands) via:

1. `/system scheduler` every 1–5 minutes  
2. `/tool fetch` to VPS API  
3. Config apply via downloaded `.rsc` + `/import`

## Capabilities

| Feature | Support |
|---------|---------|
| Enroll / approve | Yes (bootstrap token) |
| Heartbeat + hostname | Yes |
| Apply network/Wi‑Fi/firewall | Yes — server renders `.rsc` |
| Metrics/status | Yes — script posts identity/resources |
| VLESS/sing-box client | No on ROS without container → WG/IPsec if needed |

## Install sketch

1. On VPS: create bootstrap, approve device in admin/TG  
2. On router: set `NdServer`, `NdBootstrap`, `NdDeviceId`  
3. `/import netductor-agent.rsc`  
4. After approve, put device token into `NdToken`

Full apply pipeline (`apply_rsc` command) is wired on the server side next.
