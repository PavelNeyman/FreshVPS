# Установка (Netductor)

Debian-like VPS, root.

```bash
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/netductor/main/bootstrap.sh | bash
netductor install
# или TUI
netductor tui --mode vps
```

Компоненты по умолчанию:

`dirs hardening singbox blocky vpn-users api metrics telegram backup`

Пути: `/etc/netductor`, `/var/lib/netductor`, `/opt/netductor`.
