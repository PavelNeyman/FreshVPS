# Prebuilt binaries

Place cross-compiled operator tools here so VPS install does **not** need `golang-go`.

## Telegram bot

| File | Arch |
|------|------|
| `freshvps-tg-linux-amd64` | x86_64 VPS |
| `freshvps-tg-linux-arm64` | aarch64 VPS |

Build on Mac or Linux:

```bash
./scripts/build-tg-bot.sh          # both
./scripts/build-tg-bot.sh amd64    # one
```

Copy to a running VPS:

```bash
# detect: ssh root@VPS uname -m   # x86_64 → amd64
scp dist/freshvps-tg-linux-amd64 root@VPS:/opt/freshvps/bin/freshvps-tg
ssh root@VPS 'chmod 755 /opt/freshvps/bin/freshvps-tg; systemctl restart freshvps-telegram-bot'
```

Or before first install:

```bash
scp dist/freshvps-tg-linux-amd64 root@VPS:/root/freshvps-tg
# on VPS:
export FRESHVPS_TG_BIN=/root/freshvps-tg
sudo bash install.sh …
```

Binaries are optional in git (large). CI/release assets named `freshvps-tg-linux-$arch` are also tried automatically.
