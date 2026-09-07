# Upgrade and re-run (English)

**RU:** [ru/UPGRADE.md](ru/UPGRADE.md)

## Goals

- Re-running install on an **already configured** VPS must not lock you out or wipe VPN users.
- Upgrade = new tree + re-apply modules that are missing/unhealthy, or forced.

## Safe re-run (idempotent)

```bash
# same as first install — healthy modules are skipped
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh | sudo bash

# or from a tree
sudo bash install.sh --config /root/freshvps.conf --non-interactive
```

Skip logic (`lib/idempotent.sh`):

| Module | “Healthy” means |
|--------|------------------|
| hardening | fail2ban active, ufw present (**no UFW reset**) |
| sing-box | unit active + `sing-box check` OK |
| blocky | unit active |
| vpn-users | CLI + registry file exist (**users not recreated**) |
| panels | container running |

Secrets and `/etc/freshvps/vpn-users.json` are **not** deleted on re-run.

### Force

```bash
sudo FRESHVPS_FORCE=1 bash install.sh --non-interactive --config …
# or one module:
sudo FRESHVPS_FORCE_MODULES=sing-box bash install.sh …
sudo bash install.sh --force-module blocky …
sudo bash install.sh --force   # all modules
```

## Upgrade path

1. Bootstrap downloads **current** tree (or `--ref v0.3.3`).
2. Compares `VERSION` to `/var/lib/freshvps/installed_version`.
3. Runs modules with skip-if-healthy (same as re-run).
4. Writes new `installed_version`.

```bash
curl -fsSL .../bootstrap.sh | sudo bash -s -- --upgrade
# equivalent: install.sh --upgrade
```

`--upgrade` sets non-destructive defaults: does not prompt to wipe; uses existing `/etc/freshvps` and config if present.

## What will not happen on normal re-run

- `ufw --force reset` (fixed)
- New Reality keypair if secrets already exist
- New operator user if already in registry
- Deletion of panel data volumes

## OpenWrt

Site installer was already change-diff based (UCI). Re-run only commits deltas. VPN/OpenSOHO sections still gated by `ENABLE_*`.
