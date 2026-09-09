# Netductor Roadmap

Repo: https://github.com/PavelNeyman/netductor

## Phase G status

| Step | Status |
|------|--------|
| G0–G2 | done |
| G3 serve | **mostly done** — edge, session, metrics, live probes, VPN users, admin, status |
| G4 install | **started** — `ENABLE_NETDUCTOR_API`, module, `netductor install` bridge, `netductor probe` |
| G5 agent | pending Go agent |
| G6 TG rename | pending |
| G7 /opt/netductor | pending |

## API modes

- **parallel** (default): Go `:8790` + Python `:8787`
- **cutover**: `NETDUCTOR_API_MODE=cutover` → Go `:8787 --no-proxy`, stop `freshvps-api`

## CLI

```
netductor version|doctor|status|vpn|edge|serve|install|probe
```
