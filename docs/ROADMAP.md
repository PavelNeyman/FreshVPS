# Netductor Roadmap

Repo: https://github.com/PavelNeyman/netductor

## Phase G

| Step | Status |
|------|--------|
| G0–G2 | done (docs, CLI, release v0.7.0-dev) |
| G3 serve | **in progress** — health, edge, session, metrics, admin static; proxy rest to Python |
| G4 install | next |
| G5 agent | |
| G6 TG rename | |
| G7 /opt/netductor | |

### serve ownership

| Path | Handler |
|------|---------|
| `/health` | Go |
| `/api/edge/*` | Go |
| `/api/session` | Go |
| `/api/metrics*` | Go |
| `/admin/*` | Go (files from admin root) |
| other `/api/*` | proxy → Python :8787 |

