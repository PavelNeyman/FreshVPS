#!/usr/bin/env python3
"""FreshVPS admin API + SPA — Bearer session auth."""
from __future__ import annotations

import json
import os
import re
import subprocess
import time
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path
from typing import Any, Optional
from urllib.parse import parse_qs, urlparse

ETC = "/etc/freshvps"
SESS = os.path.join(ETC, "sessions")
CLIENTS = os.path.join(ETC, "clients")
VPN_BIN = "/usr/local/bin/freshvps-vpn"
METRICS_DIR = os.environ.get("FRESHVPS_METRICS_DIR", "/var/lib/freshvps/metrics")
ADMIN_ROOT = os.environ.get("FRESHVPS_ADMIN_ROOT", "/opt/freshvps/runtime/api/admin")
BIND = os.environ.get("VPN_API_BIND", "127.0.0.1")
PORT = int(os.environ.get("VPN_API_PORT", "8787"))
NAME_RE = re.compile(r"^[a-zA-Z0-9_][a-zA-Z0-9_-]{0,63}$")
SERVICES = ["sing-box", "blocky", "freshvps-api", "freshvps-telegram-bot"]


def valid_name(name: str) -> bool:
    return bool(name and NAME_RE.match(name))


def session_expiry(token: str) -> Optional[int]:
    if not token or "/" in token or ".." in token or len(token) > 128:
        return None
    path = os.path.join(SESS, token)
    if not os.path.isfile(path):
        return None
    try:
        with open(path, encoding="utf-8") as f:
            exp = int(f.read().strip())
    except (OSError, ValueError):
        return None
    if exp < int(time.time()):
        try:
            os.remove(path)
        except OSError:
            pass
        return None
    return exp


def valid_session(token: str) -> bool:
    return session_expiry(token) is not None


def read_client_text(name: str, *candidates: str) -> Optional[str]:
    for c in candidates:
        path = os.path.join(CLIENTS, name, c)
        if os.path.isfile(path):
            with open(path, encoding="utf-8") as f:
                return f.read().strip()
    return None


def _read_proc_mem() -> dict[str, int]:
    out = {"total": 0, "available": 0, "used": 0}
    try:
        with open("/proc/meminfo", encoding="utf-8") as f:
            kv = {}
            for line in f:
                p = line.split()
                if len(p) >= 2:
                    kv[p[0].rstrip(":")] = int(p[1]) * 1024
        out["total"] = kv.get("MemTotal", 0)
        out["available"] = kv.get("MemAvailable", kv.get("MemFree", 0))
        out["used"] = max(0, out["total"] - out["available"])
    except OSError:
        pass
    return out


def _cpu_pct() -> float:
    def snap():
        with open("/proc/stat", encoding="utf-8") as f:
            p = f.readline().split()
        vals = list(map(int, p[1:]))
        idle = vals[3] + (vals[4] if len(vals) > 4 else 0)
        return idle, sum(vals)

    try:
        i1, t1 = snap()
        time.sleep(0.12)
        i2, t2 = snap()
        dt, di = t2 - t1, i2 - i1
        if dt <= 0:
            return 0.0
        return round(100.0 * (1.0 - di / dt), 1)
    except OSError:
        return 0.0


def _loadavg() -> list[float]:
    try:
        with open("/proc/loadavg", encoding="utf-8") as f:
            p = f.read().split()
        return [float(p[0]), float(p[1]), float(p[2])]
    except (OSError, ValueError, IndexError):
        return [0.0, 0.0, 0.0]


def _disk_root() -> dict[str, int]:
    try:
        st = os.statvfs("/")
        total = st.f_frsize * st.f_blocks
        free = st.f_frsize * st.f_bavail
        return {"total": total, "free": free, "used": total - free}
    except OSError:
        return {"total": 0, "free": 0, "used": 0}


def _net_counters() -> dict[str, int]:
    rx = tx = 0
    try:
        with open("/proc/net/dev", encoding="utf-8") as f:
            for line in f.readlines()[2:]:
                if ":" not in line:
                    continue
                name, rest = line.split(":", 1)
                name = name.strip()
                if name in ("lo",) or name.startswith(("docker", "br-", "veth", "virbr")):
                    continue
                parts = rest.split()
                if len(parts) >= 9:
                    rx += int(parts[0])
                    tx += int(parts[8])
    except (OSError, ValueError):
        pass
    return {"rx_bytes": rx, "tx_bytes": tx}


def _service_active(name: str) -> str:
    try:
        r = subprocess.run(
            ["systemctl", "is-active", name],
            capture_output=True, text=True, timeout=3,
        )
        return (r.stdout or "unknown").strip()
    except (OSError, subprocess.TimeoutExpired):
        return "unknown"


def collect_metrics() -> dict[str, Any]:
    mem = _read_proc_mem()
    disk = _disk_root()
    services = {s: _service_active(s) for s in SERVICES}
    containers = []
    try:
        r = subprocess.run(
            ["docker", "ps", "--format", "{{.Names}}\t{{.Status}}"],
            capture_output=True, text=True, timeout=5,
        )
        if r.returncode == 0:
            for line in r.stdout.strip().splitlines():
                if "\t" in line:
                    n, st = line.split("\t", 1)
                    containers.append({"name": n, "status": st})
    except (OSError, subprocess.TimeoutExpired):
        pass
    return {
        "ts": int(time.time()),
        "cpu_pct": _cpu_pct(),
        "loadavg": _loadavg(),
        "mem": mem,
        "disk": disk,
        "net": _net_counters(),
        "services": services,
        "containers": containers,
        "hostname": os.uname().nodename,
    }


def metrics_history(limit: int = 180) -> list[dict]:
    path = os.path.join(METRICS_DIR, "history.jsonl")
    if not os.path.isfile(path):
        return []
    rows: list[dict] = []
    try:
        with open(path, encoding="utf-8") as f:
            lines = f.readlines()[-limit:]
        for line in lines:
            line = line.strip()
            if not line:
                continue
            try:
                rows.append(json.loads(line))
            except json.JSONDecodeError:
                continue
    except OSError:
        return []
    return rows


def probe_uptime(limit: int = 1440) -> dict[str, Any]:
    """Uptime % per probe name over last `limit` samples (~1 day if 1/min)."""
    rows = metrics_history(limit)
    stats: dict[str, dict[str, int]] = {}
    for row in rows:
        for p in row.get("probes") or []:
            name = p.get("name") or "?"
            st = stats.setdefault(name, {"ok": 0, "total": 0})
            st["total"] += 1
            if p.get("ok"):
                st["ok"] += 1
    out = {}
    for name, st in stats.items():
        total = st["total"] or 1
        out[name] = {
            "ok": st["ok"],
            "total": st["total"],
            "uptime_pct": round(100.0 * st["ok"] / total, 1),
        }
    return out


class Handler(BaseHTTPRequestHandler):
    def log_message(self, fmt, *args):
        return

    def _token(self) -> str:
        auth = self.headers.get("Authorization", "")
        if auth.startswith("Bearer "):
            return auth[7:].strip()
        for part in self.headers.get("Cookie", "").split(";"):
            part = part.strip()
            if part.startswith("fv_session="):
                return part.split("=", 1)[1].strip()
        return ""

    def _auth(self) -> bool:
        return valid_session(self._token())

    def _json(self, code: int, obj):
        body = json.dumps(obj).encode()
        self.send_response(code)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Cache-Control", "no-store")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def _bytes(self, code: int, data: bytes, ctype: str):
        self.send_response(code)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def _read(self):
        n = int(self.headers.get("Content-Length", 0))
        if n <= 0 or n > 65536:
            return {}
        try:
            return json.loads(self.rfile.read(n).decode() or "{}")
        except json.JSONDecodeError:
            return {}

    def _serve_admin(self, path: str):
        if path == "/admin":
            self.send_response(302)
            self.send_header("Location", "/admin/")
            self.end_headers()
            return
        root = Path(ADMIN_ROOT).resolve()
        rel = "index.html" if path in ("/admin/",) else path[len("/admin/"):]
        if ".." in rel or rel.startswith("/"):
            return self._json(400, {"error": "bad path"})
        fp = (root / rel).resolve()
        if not str(fp).startswith(str(root)) or not fp.is_file():
            return self._json(404, {"error": "not found"})
        data = fp.read_bytes()
        ctype = "text/plain"
        if rel.endswith(".html"):
            ctype = "text/html; charset=utf-8"
        elif rel.endswith(".js"):
            ctype = "application/javascript; charset=utf-8"
        elif rel.endswith(".css"):
            ctype = "text/css; charset=utf-8"
        return self._bytes(200, data, ctype)

    def do_GET(self):
        u = urlparse(self.path)
        path = u.path

        if path == "/health":
            return self._json(200, {"ok": True, "bind": BIND, "port": PORT})
        if path.startswith("/admin"):
            return self._serve_admin(path)
        if not self._auth():
            return self._json(401, {"error": "unauthorized"})

        tok = self._token()
        if path == "/api/session":
            exp = session_expiry(tok)
            now = int(time.time())
            return self._json(200, {
                "expires_unix": exp,
                "expires_in_sec": (exp - now) if exp else 0,
                "now": now,
            })

        if path in ("/api/metrics", "/metrics"):
            return self._json(200, collect_metrics())

        if path == "/api/metrics/history":
            qs = parse_qs(u.query)
            limit = int((qs.get("limit") or ["180"])[0])
            limit = max(10, min(limit, 2000))
            return self._json(200, {"points": metrics_history(limit)})

        if path == "/api/probes/uptime":
            return self._json(200, {"probes": probe_uptime(1440)})

        if path in ("/api/probes", "/api/latest"):
            latest = os.path.join(METRICS_DIR, "latest.json")
            data: dict = {}
            if os.path.isfile(latest):
                try:
                    with open(latest, encoding="utf-8") as f:
                        data = json.load(f)
                except (OSError, json.JSONDecodeError):
                    data = {}
            if not data:
                data = collect_metrics()
            if path == "/api/probes":
                return self._json(200, {
                    "probes": data.get("probes") or [],
                    "ts": data.get("ts"),
                    "uptime": probe_uptime(1440),
                })
            return self._json(200, data)

        if path == "/api/probes/config":
            cfg_path = os.path.join(ETC, "probes.json")
            if os.path.isfile(cfg_path):
                try:
                    with open(cfg_path, encoding="utf-8") as f:
                        return self._json(200, json.load(f))
                except (OSError, json.JSONDecodeError):
                    pass
            return self._json(200, {"probes": [], "alerts": {}})

        if path in ("/api/status", "/status"):
            m = collect_metrics()
            latest = os.path.join(METRICS_DIR, "latest.json")
            probes = []
            if os.path.isfile(latest):
                try:
                    with open(latest, encoding="utf-8") as f:
                        probes = json.load(f).get("probes") or []
                except (OSError, json.JSONDecodeError):
                    pass
            # stale collector?
            hist = metrics_history(2)
            last_ts = hist[-1]["ts"] if hist else 0
            stale = (int(time.time()) - last_ts) > 180 if last_ts else True
            return self._json(200, {
                "hostname": m["hostname"],
                "ts": m["ts"],
                "cpu_pct": m["cpu_pct"],
                "loadavg": m["loadavg"],
                "mem": m["mem"],
                "disk": m["disk"],
                "net": m["net"],
                "services": m["services"],
                "containers": m["containers"],
                "probes": probes,
                "collector_stale": stale,
                "collector_last_ts": last_ts,
            })

        if path == "/vpn/users":
            try:
                out = subprocess.check_output([VPN_BIN, "list"], text=True, stderr=subprocess.STDOUT)
            except subprocess.CalledProcessError as e:
                return self._json(500, {"error": e.output or "list failed"})
            users = []
            for line in out.strip().splitlines():
                p = line.split("\t")
                if len(p) >= 3:
                    users.append({
                        "name": p[0],
                        "enabled": p[1] == "on",
                        "uuid": p[2],
                        "note": p[3] if len(p) > 3 else "",
                        "created": p[4] if len(p) > 4 else "",
                    })
            return self._json(200, {"users": users})

        if path.startswith("/vpn/users/") and path.endswith("/qr"):
            name = path[len("/vpn/users/"):-len("/qr")]
            if not valid_name(name):
                return self._json(400, {"error": "bad name"})
            qr = os.path.join(CLIENTS, name, "qr.png")
            if not os.path.isfile(qr):
                return self._json(404, {"error": "no qr"})
            return self._bytes(200, Path(qr).read_bytes(), "image/png")

        if path.startswith("/vpn/users/") and path.endswith("/link"):
            name = path[len("/vpn/users/"):-len("/link")]
            if not valid_name(name):
                return self._json(400, {"error": "bad name"})
            sub = read_client_text(name, "subscription.txt", "link.txt")
            if sub is None:
                return self._json(404, {"error": "not found"})
            return self._json(200, {
                "name": name,
                "subscription": sub,
                "vless": read_client_text(name, "link-vless.txt", "link.txt"),
                "hy2": read_client_text(name, "link-hy2.txt"),
            })

        return self._json(404, {"error": "not found"})

    def do_POST(self):
        u = urlparse(self.path)
        path = u.path
        if not self._auth():
            return self._json(401, {"error": "unauthorized"})
        body = self._read()

        if path == "/vpn/users":
            name = str(body.get("name", "")).strip()
            note = str(body.get("note", "")).strip()
            if not valid_name(name):
                return self._json(400, {"error": "bad name"})
            args = [VPN_BIN, "add", name]
            if note:
                args.append(note)
            try:
                out = subprocess.check_output(args, text=True, stderr=subprocess.STDOUT)
            except subprocess.CalledProcessError as e:
                return self._json(500, {"error": e.output or "add failed"})
            return self._json(200, {"ok": True, "output": out.strip(), "name": name})

        if path.startswith("/vpn/users/") and path.endswith("/note"):
            name = path[len("/vpn/users/"):-len("/note")]
            if not valid_name(name):
                return self._json(400, {"error": "bad name"})
            note = str(body.get("note", ""))
            try:
                out = subprocess.check_output(
                    [VPN_BIN, "note", name, note], text=True, stderr=subprocess.STDOUT
                )
            except subprocess.CalledProcessError as e:
                return self._json(500, {"error": e.output or "note failed"})
            return self._json(200, {"ok": True, "output": out.strip()})

        m = re.match(r"^/vpn/users/([^/]+)/(disable|enable|revoke)$", path)
        if m:
            name, action = m.group(1), m.group(2)
            if not valid_name(name):
                return self._json(400, {"error": "bad name"})
            try:
                out = subprocess.check_output(
                    [VPN_BIN, action, name], text=True, stderr=subprocess.STDOUT
                )
            except subprocess.CalledProcessError as e:
                return self._json(500, {"error": e.output or "failed"})
            return self._json(200, {"ok": True, "output": out.strip()})

        if path == "/api/probes/config":
            # write full config
            cfg_path = os.path.join(ETC, "probes.json")
            try:
                with open(cfg_path, "w", encoding="utf-8") as f:
                    json.dump(body, f, indent=2)
                    f.write("\n")
            except OSError as e:
                return self._json(500, {"error": str(e)})
            return self._json(200, {"ok": True})

        return self._json(404, {"error": "not found"})


def main():
    os.makedirs(METRICS_DIR, exist_ok=True)
    HTTPServer((BIND, PORT), Handler).serve_forever()


if __name__ == "__main__":
    main()
