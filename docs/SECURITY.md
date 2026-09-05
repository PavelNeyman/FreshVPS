# Security notes — FreshVPS

## Principles

- Single operator; no multi-tenant isolation claims
- Secrets only under `/etc/freshvps/secrets` (mode 600/700)
- Never commit real tokens, keys, or client UUIDs to git

## Default posture

- SSH: public key only; root `prohibit-password`
- fail2ban enabled
- UFW: deny incoming by default; allow SSH + module ports
- unattended-upgrades enabled
- BBR enabled

## VPN

- VLESS + Reality reduces plaintext fingerprinting; still rotate keys if leaked
- Hysteria2 uses a local self-signed cert; clients may need insecure/skip-verify or pin the cert
- Prefer restricting management UIs (Kuma, Beszel, OpenSOHO, Blocky API) to VPN or firewall allowlists

## DNS

- Blocky on 53; host `resolv.conf` points to 127.0.0.1 after install
- Exposing DNS publicly is optional and risky (amplification / abuse); prefer VPN or router forwarders only

## Recommendations

1. Change default public UI ports or put them behind reverse proxy + auth
2. Store restic password offline; test restore
3. Keep OpenSOHO shared secret long and only on trusted routers
4. Review UFW after install: `ufw status numbered`
