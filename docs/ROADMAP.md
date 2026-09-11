# Roadmap

## Done
- Go-only control plane (install, VPN, API, TUI, TG, edge)
- Production auth: session hash-at-rest, device tokens, enroll rate-limit, TG admin allowlist
- Node registry with **stable UUID** + separate hostname (bidirectional rename)
- APT install hardened (skip installed pkgs, optional fail2ban)
- Admin UI + Nodes tab

## Next
- OpenWrt e2e on real hardware (enroll → template → VPN client)
- MikroTik scheduler agent production tick
- Split `cmd/netductor` / `cmd/netductor-tg` into smaller packages
- Backup offsite

## Tests
- VPS smoke (clean + upgrade path)
- Human/router when hardware available
