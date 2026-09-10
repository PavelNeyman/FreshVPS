# Обновление

```bash
curl -fsSL -o /usr/local/bin/netductor \
  https://github.com/PavelNeyman/netductor/releases/download/v0.7.0-dev/netductor-linux-amd64
chmod 755 /usr/local/bin/netductor
netductor install
systemctl restart netductor-api
```
