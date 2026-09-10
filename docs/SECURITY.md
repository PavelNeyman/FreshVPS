# Security

- API binds 127.0.0.1 by default; sessions under `/etc/netductor/sessions`.
- Secrets: `/etc/netductor/secrets` (700/600).
- Backups: AES-256-CBC (`openssl` + `backup_key`).
- Edge agent: token auth; allowlisted actions only.
- VPN: Reality + HY2; operator-only provisioning.
