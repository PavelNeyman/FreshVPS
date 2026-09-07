# Security notes — FreshVPS (English)

**RU:** [ru/SECURITY.md](ru/SECURITY.md)

## Principles

- Single operator; no multi-tenant isolation claims
- Secrets only under `/etc/freshvps/secrets` (mode 600/700)
- Never commit tokens, keys, client UUIDs, or real Wi‑Fi passwords

## Default posture

- SSH: public key; root `prohibit-password`
- fail2ban; UFW deny incoming by default
- unattended-upgrades; BBR

## VPN

- Multi-user **VLESS+Reality** and **Hysteria2** (per-user HY2 password)
- Hand out `subscription.txt`, not a shared global HY2 password
- HY2 uses local self-signed TLS (`insecure=1` in generated links)
- Rotate Reality keys if leaked

## DNS / panels

- Blocky on 53; prefer not exposing DNS to the whole Internet
- OpenSOHO / Kuma / Beszel / API bind **127.0.0.1** — SSH tunnel

## Recommendations

1. Offline restic password; test restore
2. Long OpenSOHO shared secret on trusted routers only
3. `ufw status numbered` after install
