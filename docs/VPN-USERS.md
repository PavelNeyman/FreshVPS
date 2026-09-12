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

Default is **ya.ru** (not Cloudflare). Change:

```bash
netductor vpn set-sni ya.ru
# or: SINGBOX_REALITY_SNI=... at install
# stored in /etc/netductor/secrets/singbox_reality_sni
```

## Reality SNI and Russian whitelist

Under whitelist/DPI, Reality **SNI/dest must be a RU-whitelisted domain**. Defaults like Cloudflare or Microsoft are wrong for this threat model.

| Access path | Approach |
|-------------|----------|
| **Home broadband** | Foreign core VPS + Reality, SNI from whitelist (default **`ya.ru`**; also `vk.com`, `mail.ru`, …) |
| **Mobile (strict whitelist)** | Phone → **RU intermediate VPS** (whitelist IP: Yandex/VK/Timeweb Cloud) → foreign netductor core. SNI on the RU entry also whitelist. |

```bash
netductor vpn set-sni ya.ru
# stored in /etc/netductor/secrets/singbox_reality_sni
```

RU relay is **not auto-provisioned** by netductor yet; document/manual hop for mobile until a dedicated plane exists.

