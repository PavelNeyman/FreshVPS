# Edge agent (OpenWrt)

## Install

```sh
curl -fsSL -o /usr/sbin/netductor-agent \
  https://github.com/PavelNeyman/netductor/releases/download/v0.7.0-dev/netductor-agent-linux-arm64
chmod 755 /usr/sbin/netductor-agent
mkdir -p /etc/netductor-agent
# SERVER TOKEN DEVICE_ID INTERVAL in config
```

## Commands

```bash
netductor edge cmd site1 config_backup
netductor edge cmd site1 config_restore 20260310-120000.tar.gz
netductor edge cmd site1 metrics
netductor edge cmd site1 uci_batch $'network.lan.ipaddr=192.168.1.1\ncommit\nnetwork_reload'
netductor edge cmd site1 agent_update 'https://…/agent|sha256'
netductor edge cmd site1 sysupgrade 'https://…/fw.bin|sha256|confirm=yes'
```

Backups on VPS: `/var/lib/netductor/edge/<id>/backups/`  
API: `GET /api/edge/backups?device_id=site1`  
Download: `GET /api/edge/backups?device_id=site1&name=….tar.gz` (session or edge token)
