# Security

- API binds 127.0.0.1 by default; sessions under `/etc/netductor/sessions`.
- Secrets: `/etc/netductor/secrets` (700/600).
- Backups: AES-256-CBC (`openssl` + `backup_key`).
- Edge agent: token auth; allowlisted actions only.
- VPN: Reality + HY2; operator-only provisioning.

## Authentication model (production)

### Operator session (Admin API / Shortcuts)

- Token: **256-bit** random, shown **once**
- On disk: only **SHA-256 hash** of token + expiry metadata (`/etc/netductor/sessions/*.json`)
- Default TTL **8h**, hard max **72h**
- `POST /api/session/revoke` — revoke current or `{ "all": true }`
- Headers: `nosniff`, `DENY` frame, `no-store`

### Edge devices

- **Bootstrap token**: enroll only (`/api/edge/enroll`)
- **Device token**: issued on approve, **256-bit**, constant-time compare
- Global `edge_token` **disabled** unless `NETDUCTOR_EDGE_LEGACY_TOKEN=1`
- Enroll **rate-limited** per IP (10 / 15 min)
- Pending devices require human approve (Admin/TG)

### Telegram bot

- Operator allowlist: `/etc/netductor/secrets/telegram_admin_id` or `NETDUCTOR_TG_ADMIN`
- First-writer claim **disabled** unless `NETDUCTOR_TG_CLAIM_FIRST=1`

### Network

- API listens on `127.0.0.1` by default — expose only via SSH tunnel, VPN, or reverse proxy with TLS
