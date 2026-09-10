# Install (Netductor)

Debian-like VPS, root.

```bash
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/netductor/main/bootstrap.sh | bash
netductor install
# or interactive
netductor tui --mode vps
```

Components (default all core):

`dirs hardening singbox blocky vpn-users api metrics telegram backup`

```bash
netductor install singbox blocky api
netductor doctor
netductor vpn list
```

Paths: `/etc/netductor`, `/var/lib/netductor`, `/opt/netductor`.
