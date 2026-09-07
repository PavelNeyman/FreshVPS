# Upgrade and re-run (English)

**RU:** [ru/UPGRADE.md](ru/UPGRADE.md)

Project assumes **greenfield** installs (recreate VPS if needed). No legacy migration tooling.

## Re-run / upgrade

```bash
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh | sudo bash -s -- --upgrade
```

- Healthy modules **skipped** (`lib/idempotent.sh`)
- Choices restored from `/etc/freshvps/install.conf` when present
- Secrets, `vpn-users.json`, panel volumes **kept**
- UFW is **never** reset

### Force

```bash
sudo bash install.sh --force --non-interactive
sudo bash install.sh --force-module blocky --upgrade
```

### SSH lockout

Password auth is disabled only if `authorized_keys` exists (or `FRESHVPS_ALLOW_PASSWORD_SSH=1`).

### After install

```bash
sudo freshvps-doctor
sudo bash scripts/smoke-host.sh   # if tree still on disk
```
