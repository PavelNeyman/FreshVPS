# Prebuilt binaries

Place cross-compiled operator tools here so VPS install does **not** need `golang-go`.

## Telegram bot

| File | Arch |
|------|------|
| `netductor-tg-linux-amd64` | x86_64 VPS |
| `netductor-tg-linux-arm64` | aarch64 VPS |

Build on Mac or Linux:

```bash
./scripts/build-tg-bot.sh          # both
./scripts/build-tg-bot.sh amd64    # one
```

Copy to a running VPS:

```bash
# detect: ssh root@VPS uname -m   # x86_64 → amd64
scp dist/netductor-tg-linux-amd64 root@VPS:/opt/netductor/bin/netductor-tg
ssh root@VPS 'chmod 755 /opt/netductor/bin/netductor-tg; systemctl restart netductor-telegram-bot'
```

Or before first install:

```bash
scp dist/netductor-tg-linux-amd64 root@VPS:/root/netductor-tg
# on VPS:
export NETDUCTOR_TG_BIN=/root/netductor-tg
sudo bash install.sh …
```

Binaries are optional in git (large). CI/release assets named `netductor-tg-linux-$arch` are also tried automatically.
