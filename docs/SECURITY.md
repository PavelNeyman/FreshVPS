# Security notes

## SSH

Password authentication is disabled **only** when `authorized_keys` is present (or `FRESHVPS_ALLOW_PASSWORD_SSH=1`).

## VPN

- **VLESS + Reality** — preferred client path.
- **Hysteria2** uses a **self-signed** certificate; client links set `insecure=1` for convenience. Treat as weaker authenticity than Reality. Prefer VLESS when possible.

## Admin API

- Default bind `127.0.0.1`.
- Auth: **Bearer session** from `freshvps-vpn session [hours]` only (no master token).
- Prefer SSH tunnel or access only over your VPN.

## VPS tests

`freshvps-tests` may pipe third-party scripts; use on evaluation VPS or accept the trust model.
