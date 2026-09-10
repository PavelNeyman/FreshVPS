# Edge agent

Binary: GitHub Releases `netductor-agent-linux-{arm64,arm,mipsle,amd64}`.

Config `/etc/netductor-agent/config`:

```
SERVER=http://127.0.0.1:8787
TOKEN=...
DEVICE_ID=site1
INTERVAL=60
```

## Commands (allowlist)

| action | arg |
|--------|-----|
| ping | |
| status | |
| reboot | |
| wifi_reload | |
| network_reload | |
| uci_get | `network.lan.ipaddr` |
| uci_show | `wireless` |
| uci_set | `path=value` |
| uci_commit | |
| logread | |

```bash
netductor edge cmd site1 ping
netductor edge cmd site1 uci_get network.lan.ipaddr
```
