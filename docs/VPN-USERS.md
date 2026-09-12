# VPN users

```bash
netductor vpn add alice "laptop"
netductor vpn list
netductor vpn link alice
netductor vpn disable alice
netductor vpn revoke alice
```

Files: `/etc/netductor/clients/<name>/` (`subscription.txt`, `qr.png`).


## Reality SNI

Default is **www.microsoft.com** (not Cloudflare). Change:

```bash
netductor vpn set-sni www.microsoft.com
# or: SINGBOX_REALITY_SNI=... at install
# stored in /etc/netductor/secrets/singbox_reality_sni
```
