# Security

## Defaults
- API bind `127.0.0.1` only; public bind requires `NETDUCTOR_API_PUBLIC=1` **and TLS cert/key**
- Sessions: 256-bit, hash-at-rest, max 72h; cookie `HttpOnly` + `SameSite=Strict` (+ `Secure` with TLS)
- Edge: bootstrap token ≠ device token; enroll rate-limit; human approve
- Global `edge_token` disabled unless `NETDUCTOR_EDGE_LEGACY_TOKEN=1`
- Destructive agent cmds need `confirm=yes`: reboot, agent_update, sysupgrade
- Edge command allowlist on enqueue
- API errors truncated via `publicErr`
- Audit: vpn.add, edge.approve/deny/revoke, nodes.rename, session.revoke

## Optional Lampac
- Off by default. `NETDUCTOR_LAMPAC=1` or `netductor install lampac`
- Docker image `immisterio/lampac` on port 9118

## Operator
- Keep API behind SSH/VPN
- Rotate device tokens: `edge rotate` / API (when exposed)
- Verify updates: `NETDUCTOR_UPDATE_SHA256=<hex> netductor update`
