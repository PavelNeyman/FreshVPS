# FreshVPS Roadmap

## Done (0.5.x)

- Built-in Admin UI (no Kuma/Beszel by default; no Prometheus)
- Metrics JSONL + probes (TCP/UDP/HTTP) + TG alerts/recovery
- Multi-user VPN, Telegram bot (menus, cards), API session
- OpenSOHO optional; OpenWrt role separate

## Next

1. Optional: Mini App shell pointing at same Admin UI (needs HTTPS + public/VPN URL)
2. Optional: per-user traffic (sing-box stats / access logs)
3. OpenWrt physical devices + multi-site
4. Release binaries for `freshvps-tg` in GitHub Releases
5. Human + AI test plans on clean VPS after each minor

## Out of scope (for now)

- Full Prometheus/Grafana stack
- Embedding OpenSOHO into Admin UI
