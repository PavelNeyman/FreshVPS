# Bootstrap (русский)

**EN:** [../BOOTSTRAP.md](../BOOTSTRAP.md)

GitHub **на лету** отдаёт `tar.gz` репозитория (Actions не нужны).

> **Приватный репозиторий:** `raw.githubusercontent.com` и анонимный `codeload` отвечают **404**. Нужен public repo, либо PAT / `git clone` с доступом, либо копирование дерева на VPS (`scp` tarball).

## VPS (Debian)

**Предпочтительно** (без сюрпризов curl|bash):

```bash
curl -fsSL https://raw.githubusercontent.com/PavelNeyman/FreshVPS/main/bootstrap.sh -o /tmp/fv.sh
sudo bash /tmp/fv.sh
```

Также: `curl … | sudo bash` (нужен root/sudo с самого начала).

Обновление: `sudo bash /tmp/fv.sh --upgrade`

После установки: `freshvps-doctor` · `freshvps-smoke` · `freshvps-vpn link operator`

## macOS (plan)

```bash
curl -fsSL …/bootstrap.sh | bash -s -- --mode plan --keep
```

VPN на Mac **не** устанавливается.

## Режимы

| ОС | Режим по умолчанию |
|----|---------------------|
| Darwin | plan |
| Debian Linux | vps |
| OpenWrt | openwrt |

Явно: `--mode vps|edge-client|openwrt|plan`
