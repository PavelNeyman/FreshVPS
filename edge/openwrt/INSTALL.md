# Netductor edge agent (OpenWrt)

Outbound-only: router dials the VPS (NAT/CGNAT OK).

## Go agent (preferred)

```sh
# linux/arm64 example — pick arch for your board
ARCH=arm64   # or arm, amd64, mipsle
TAG=v0.7.0-dev
curl -fsSL -o /usr/sbin/netductor-agent \
  "https://github.com/PavelNeyman/netductor/releases/download/${TAG}/netductor-agent-linux-${ARCH}"
chmod 755 /usr/sbin/netductor-agent

mkdir -p /etc/netductor-agent
cat > /etc/netductor-agent/config <<CFG
SERVER=https://your.vps.or.ip:8790
TOKEN=paste-edge-token-from-vps
DEVICE_ID=cudy-home
INTERVAL=60
CFG
chmod 600 /etc/netductor-agent/config

# simple procd init (or use systemd on non-OpenWrt)
cat > /etc/init.d/netductor-agent <<'INIT'
#!/bin/sh /etc/rc.common
START=99
USE_PROCD=1
start_service() {
  procd_open_instance
  procd_set_param command /usr/sbin/netductor-agent
  procd_set_param respawn
  procd_close_instance
}
INIT
chmod 755 /etc/init.d/netductor-agent
/etc/init.d/netductor-agent enable
/etc/init.d/netductor-agent start
```

Legacy shell agent: `freshvps-agent` still works with `/etc/freshvps-agent/config`.

## Allowlisted commands

`ping` · `reboot` · `wifi_reload` · `network_reload` · `uci_get` · `uci_show` · `logread`
