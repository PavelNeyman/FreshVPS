# Roadmap

## Done
- Go-only control plane (install, VPN, API, TUI, TG, edge)
- Production auth: session hash-at-rest, device tokens, enroll rate-limit, TG admin allowlist
- Node registry with **stable UUID** + separate hostname
- `buildAPIMux` + httptest API tests; agent enroll/heartbeat tests
- `netductor update` (stop api → replace binary → restart)
- Audit log: nodes.rename, session.revoke, edge.approve/deny/revoke, vpn.add
- APT install hardened

## Next
- OpenWrt e2e on real hardware
- MikroTik scheduler agent production tick
- Backup offsite
- Typed edge.Device

## Tests
- `go test ./...` on every change
- VPS clean install smoke
