# Edge agent (OpenWrt)

## Install

```sh
# pick arch: arm64 | arm | mipsle | amd64
curl -fsSL -o /usr/sbin/netductor-agent \
  https://github.com/PavelNeyman/netductor/releases/download/v0.7.0-dev/netductor-agent-linux-arm64
chmod 755 /usr/sbin/netductor-agent
mkdir -p /etc/netductor-agent
cat > /etc/netductor-agent/config <<CFG
SERVER=http://YOUR_VPS:8787
TOKEN=edge-token-from-vps
DEVICE_ID=site1
INTERVAL=60
CFG
```

Procd: `edge/openwrt/netductor-agent.init`

## Commands from VPS

```bash
netductor edge cmd site1 ping
netductor edge cmd site1 metrics
netductor edge cmd site1 config_backup
netductor edge cmd site1 uci_get network.lan.ipaddr
netductor edge cmd site1 uci_batch $'network.lan.ipaddr=192.168.1.1\ncommit\nnetwork_reload'
netductor edge cmd site1 agent_update 'https://…/netductor-agent-linux-arm64|sha256hex'
netductor edge cmd site1 sysupgrade 'https://…/firmware.bin|sha256hex|confirm=yes'
```

| Action | Notes |
|--------|--------|
| `config_backup` | `/etc/config` → VPS `/var/lib/netductor/edge/<id>/backups/` |
| `metrics` / heartbeat | history: `…/edge/<id>/metrics.jsonl` |
| `uci_batch` | lines `path=value`, `commit`, `network_reload`, `wifi_reload` |
| `agent_update` | `URL` or `URL\|sha256` |
| `sysupgrade` | requires `confirm=yes`; verifies sha256 |

Token: `/etc/netductor/secrets/edge_token` on VPS.
