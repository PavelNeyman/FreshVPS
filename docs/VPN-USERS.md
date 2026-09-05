# Multi-user VPN & admin control (design)

> Status: **Agreed** — implementation pending (after core install smoke-test).
> Scope: people/devices connecting **to the VPS** (not OpenWrt).

## Goals

- Only the operator issues access (no end-user self-service).
- Users never talk to Telegram; they only receive a link/QR via any channel the operator chooses.
- Maximum automation: registry, sing-box config, DNS via Blocky inside the tunnel, service reload, client artifacts.
- Admin surfaces: **CLI** (source of truth), **Telegram** (operator-only panel), **API + Apple Shortcuts** (when on VPN).

## User model

| Role | Actions |
|------|---------|
| Operator | add / list / disable / enable / revoke / export link+QR |
| End user | import QR or `vless://` into a client; no bot, no API |

- **VLESS + Reality**: one UUID per user (primary).
- **Hysteria2**: optional shared or later per-user; not required for multi-user v1.
- Registry: `/etc/freshvps/vpn-users.json`
- Artifacts: `/etc/freshvps/clients/<name>/` (`link.txt`, `qr.png`)

## DNS (Blocky)

VPN clients must resolve through Blocky on the VPS (server-side DNS / hijack in sing-box), so ad/tracker filtering applies without manual DNS on the phone.

Prefer Blocky on localhost (or VPN-only), not a public open resolver.

## Admin channels

| Channel | When |
|---------|------|
| `freshvps-vpn` CLI | SSH; automation; source of truth |
| Telegram bot | Operator chat id only; same operations as CLI |
| HTTP API | Only reachable **from the operator VPN**; used by Shortcuts |

If VPN is down: API/Shortcuts unavailable by design → use Telegram or CLI.

## API + Shortcuts security (agreed)

1. **VPN-only**: API bind on VPN interface / not exposed on public WAN.
2. **Session token**: short-lived session issued via CLI or Telegram; **master token never** embedded in a Shortcut.
3. **Face ID gate**: Shortcut starts with system Authenticate (Face ID / passcode) before any network call.
4. **Token file**: session (or current token) stored in a **local** file on iPhone (prefer On My iPhone, not iCloud); Shortcut reads file after Face ID.

Optional later: native Keychain helper app — **not** required for v1.

### Suggested API shape (implementation)

- `Authorization: Bearer <session>`
- `POST /vpn/users` — create
- `GET /vpn/users` — list
- `POST /vpn/users/{name}/disable|enable`
- `DELETE /vpn/users/{name}`
- `GET /vpn/users/{name}/qr` — `image/png`
- `POST /auth/session` — only from already-authenticated operator path (or CLI-minted)

All mutating ops call the same core as CLI (generate UUID → write registry → render sing-box → validate → restart → write QR).

## Explicit non-goals (this doc)

- OpenWrt / OpenSOHO user onboarding
- Public registration or invite codes for strangers
- Full web admin UI (optional later; same VPN-only rules)
