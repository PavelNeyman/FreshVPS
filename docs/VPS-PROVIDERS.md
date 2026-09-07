# VPS providers (RU payment / VPN exit) — snapshot 2026-09

**RU:** [ru/VPS-PROVIDERS.md](ru/VPS-PROVIDERS.md)

Summary from public rankings and tests. **Not a substitute** for your own `iperf3`/ping from your ISP.

## Channel: claimed vs measured

| Provider | Typical claimed port | Public measurements (ballpark) | Notes |
|----------|----------------------|--------------------------------|-------|
| **Aeza** | 1–25 Gbit (by location) | Speedtest snapshots up to **~700–800 Mbps** DL/UL on some EU nodes | User Speedtests (e.g. AMS); Habr stronger on CPU/disk |
| **Timeweb** | 1 Gbit | Reviews often **hundreds Mbps–~1 Gbit** DL; UL depends on plan | Habr NL/RU; balanced |
| **VDSina** | up to 1 Gbit | Mid-tier **comparable** to Timeweb on network | Habr comparisons |
| **RUVDS** | up to 1 Gbit | Cheap plans: **~900 Mbps DL / ~100 Mbps UL**; higher tiers closer to symmetric | Habr / VDS tests |
| **FirstVDS** | 100 M–1 G | NL reviews: strong CPU; net **~700 Mbps–Gbit** | Habr NL 2024–26 |
| **AdminVPS** | 1 Gbit | Mid class; few public iperf lists | Aggregators + Habr |
| **Hetzner** | 1–20 Gbit | Sustained **~900+ Mbps** in independent EU tests | **RU mobile often poor** (blocklists) |
| **Fornex / ishosting / 4VPS** | 1 Gbit (often) | High variance; toaster plans drop at peak | Measure yourself |

**Takeaway:** almost everyone claims “1 Gbit”. For VPN, **peak-hour stability** and ping from **your** ISP matter. Symmetric UL helps multi-client Reality/HY2.

## Payment / location / role

| Provider | Pay from RU | EU exit | Stability* | Support* | Entry price |
|----------|-------------|---------|------------|----------|-------------|
| Timeweb | excellent | yes | high | high | mid |
| VDSina | excellent | yes | good | good | low–mid |
| AdminVPS | excellent | yes | good | good | low |
| Aeza | cards + crypto | yes | top iron / 2026 reputation mixed | fast | mid |
| Fornex | often crypto | yes | calm | RU ok | mid |
| ishosting | crypto+cards | yes, many DCs | medium | ok | mid |
| 4VPS / FirstByte | RU/crypto | yes | toaster lottery | mixed | low |
| Selectel/Beget/FirstVDS | excellent | weaker as exit | high (infra) | high | varies |

\*Qualitative from reviews, not an SLA.

## For FreshVPS

- At least **2 GB RAM**, KVM, Debian 12/13, public IPv4.
- Prefer **EU** (FI/NL/DE).
- After order: 2–3 days of ping + `iperf3` + phone tunnel; keep a **second** host on another ASN.
