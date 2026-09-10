# Backup

```bash
netductor backup
```

Local encrypted files under `/var/lib/netductor/backups/`.
Key: `/etc/netductor/secrets/backup_key`.

## Offsite

`/etc/netductor/backup.offsite`:

```
method=scp
target=user@host:/var/backups/netductor/
scp_opts=-i /root/.ssh/id_ed25519
```

Methods: `scp`, `rsync`, `http` (curl PUT).
