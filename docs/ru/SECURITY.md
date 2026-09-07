# Безопасность — FreshVPS (русский)

**EN:** [../SECURITY.md](../SECURITY.md)

- Один оператор; секреты только в `/etc/freshvps/secrets` (600/700)
- Не коммитить токены, ключи, UUID клиентов, реальные пароли Wi‑Fi
- SSH по ключу; fail2ban; UFW deny by default
- VLESS+Reality; HY2 — локальный self-signed (в ссылке `insecure=1` для удобства)
- Панели OpenSOHO/Kuma/Beszel на **127.0.0.1** — доступ через SSH-туннель
- DNS 53 наружу не открывать без нужды
- Пароль restic хранить офлайн; тестировать restore
