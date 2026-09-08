# FreshVPS edge agent (OpenWrt)

Outbound-only: router connects to your VPS (NAT/CGNAT OK).

## Router

```sh
mkdir -p /etc/freshvps-agent
cat > /etc/freshvps-agent/config <<CFG
SERVER=https://your.domain
TOKEN=paste-edge-token-from-vps
DEVICE_ID=cudy-home
INTERVAL=60
CFG
chmod 600 /etc/freshvps-agent/config
cp freshvps-agent /usr/sbin/freshvps-agent && chmod 755 /usr/sbin/freshvps-agent
cp freshvps-agent.init /etc/init.d/freshvps-agent && chmod 755 /etc/init.d/freshvps-agent
/etc/init.d/freshvps-agent enable
/etc/init.d/freshvps-agent start
```

Needs `curl` or `wget`.
