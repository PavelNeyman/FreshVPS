# Backup & restore

```bash
netductor backup
netductor restore /var/lib/netductor/backups/netductor-YYYYMMDD-HHMMSS.tar.gz.ndenc
```

- Encryption: AES-256-GCM (Go), key `/etc/netductor/secrets/backup_key`
- Extension `.ndenc` (not openssl)
- Offsite: `/etc/netductor/backup.offsite` (`scp`|`rsync`|`http`)

## TLS for API

```bash
netductor serve --bind 0.0.0.0 --port 8443   --tls-cert /etc/netductor/tls/cert.pem   --tls-key /etc/netductor/tls/key.pem
```
