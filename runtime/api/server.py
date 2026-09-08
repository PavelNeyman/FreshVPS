#!/usr/bin/env python3
"""FreshVPS admin API + SPA — Bearer session (freshvps-vpn session)."""
from __future__ import annotations

import json
import os
import re
import subprocess
import time
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path
from typing import Any, Optional
from urllib.parse import urlparse

ETC = "/etc/freshvps"
SESS = os.path.join(ETC, "sessions")
CLIENTS = os.path.join(ETC, "clients")
VPN_BIN = "/usr/local/bin/freshvps-vpn"
METRICS_DIR = os.environ.get("FRESHVPS_METRICS_DIR", "/var/lib/freshvps/metrics")
ADMIN_ROOT = os.environ.get(
    "FRESHVPS_ADMIN_ROOT", "/opt/freshvps/runtime/api/admin"
)
BIND = os.environ.get("VPN_API_BIND", "127.0.0.1")
PORT = int(os.environ.get("VPN_API_PORT", "8787"))
NAME_RE = re.compile(r"^[a-zA-Z0-9_][a-zA-Z0-9_-]{0,63}$")

SERVICES = [
    "sing-box",
    "blocky",
    "freshvps-api",
    "freshvps-telegram-bot",
]


def valid_name(name: str) -> bool:
    return bool(name and NAME_RE.match(name))


def valid_session(token: str) -> bool:
    if not token or "/" in token or ".." in token or len(token) > 128:
        return False
    path = os.path.join(SESS, token)
    if not os.path.isfile(path):
        return False
    try:
        with open(path, encoding="utf-8") as f:
            exp = int(f.read().strip())
    except (OSError, ValueError):
        return False
    if exp < int(time.time()):
        try:
            os.remove(path)
        except OSError:
            pass
        return False
    return True


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
    """Rough CPU busy% over ~0.15s."""
    def snap():
        with open("/proc/stat", encoding="utf-8") as f:
            p = f.readline().split()
        vals = list(map(int, p[1:]))
        idle = vals[3] + (vals[4] if len(vals) > 4 else 0)
        return idle, sum(vals)

    try:
        i1, t1 = snap()
        time.sleep(0.15)
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


def _service_active(name: str) -> str:
    try:
        r = subprocess.run(
            ["systemctl", "is-active", name],
            capture_output=True,
            text=True,
            timeout=3,
        )
        return (r.stdout or r.stderr or "unknown").strip()
    except (OSError, subprocess.TimeoutExpired):
        return "unknown"


def collect_metrics() -> dict[str, Any]:
    mem = _read_proc_mem()
    disk = _disk_root()
    services = {s: _service_active(s) for s in SERVICES}
    # optional containers
    try:
        r = subprocess.run(
            ["docker", "ps", "--format", "{{.Names}}\t{{.Status}}"],
            capture_output=True,
            text=True,
            timeout=5,
        )
        containers = []
        if r.returncode == 0:
            for line in r.stdout.strip().splitlines():
                if "\t" in line:
                    n, st = line.split("\t", 1)
                    containers.append({"name": n, "status": st})
        services["_containers"] = containers  # type: ignore
    except (OSError, subprocess.TimeoutExpired):
        pass

    return {
        "ts": int(time.time()),
        "cpu_pct": _cpu_pct(),
        "loadavg": _loadavg(),
        "mem": mem,
        "disk": disk,
        "services": {k: v for k, v in services.items() if k != "_containers"},
        "containers": services.get("_containers", []),
        "hostname": os.uname().nodename,
    }


def metrics_history(limit: int = 120) -> list[dict]:
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


class Handler(BaseHTTPRequestHandler):
    def log_message(self, fmt, *args):
        return

    def _token_from_request(self) -> str:
        auth = self.headers.get("Authorization", "")
        if auth.startswith("Bearer "):
            return auth[7:].strip()
        # cookie for SPA
        cookie = self.headers.get("Cookie", "")
        for part in cookie.split(";"):
            part = part.strip()
            if part.startswith("fv_session="):
                return part.split("=", 1)[1].strip()
        return ""

    def _auth(self) -> bool:
        return valid_session(self._token_from_request())

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
        root = Path(ADMIN_ROOT).resolve()
        if path in ("/admin", "/admin/"):
            rel = "index.html"
        else:
            rel = path[len("/admin/") :]
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
        elif rel.endswith(".svg"):
            ctype = "image/svg+xml"
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

        if path == "/api/metrics" or path == "/metrics":
            return self._json(200, collect_metrics())

        if path == "/api/metrics/history":
            return self._json(200, {"points": metrics_history(180)})

        if path in ("/api/probes", "/api/latest"):
            latest = os.path.join(METRICS_DIR, "latest.json")
            if os.path.isfile(latest):
                try:
                    with open(latest, encoding="utf-8") as f:
                        data = json.load(f)
                except (OSError, json.JSONDecodeError):
                    data = {}
            else:
                data = collect_metrics()
            if path == "/api/probes":
                return self._json(200, {"probes": data.get("probes") or [], "ts": data.get("ts")})
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

        if path == "/api/status" or path == "/status":
            m = collect_metrics()
            return self._json(
                200,
                {
                    "hostname": m["hostname"],
                    "ts": m["ts"],
                    "cpu_pct": m["cpu_pct"],
                    "loadavg": m["loadavg"],
                    "mem": m["mem"],
                    "disk": m["disk"],
                    "services": m["services"],
                    "containers": m["containers"],
                },
            )

        if path == "/vpn/users":
            try:
                out = subprocess.check_output(
                    [VPN_BIN, "list"], text=True, stderr=subprocess.STDOUT
                )
            except subprocess.CalledProcessError as e:
                return self._json(500, {"error": e.output or "list failed"})
            users = []
            for line in out.strip().splitlines():
                p = line.split("\t")
                if len(p) >= 3:
                    users.append(
                        {"name": p[0], "enabled": p[1] == "on", "uuid": p[2]}
                    )
            return self._json(200, {"users": users})

        if path.startswith("/vpn/users/") and path.endswith("/qr"):
            name = path[len("/vpn/users/") : -len("/qr")]
            if not valid_name(name):
                return self._json(400, {"error": "bad name"})
            qr = os.path.join(CLIENTS, name, "qr.png")
            if not os.path.isfile(qr):
                return self._json(404, {"error": "no qr"})
            data = Path(qr).read_bytes()
            return self._bytes(200, data, "image/png")

        if path.startswith("/vpn/users/") and path.endswith("/link"):
            name = path[len("/vpn/users/") : -len("/link")]
            if not valid_name(name):
                return self._json(400, {"error": "bad name"})
            sub = read_client_text(name, "subscription.txt", "link.txt")
            if sub is None:
                return self._json(404, {"error": "not found"})
            vless = read_client_text(name, "vless.txt")
            hy2 = read_client_text(name, "hy2.txt")
            return self._json(
                200,
                {
                    "name": name,
                    "subscription": sub,
                    "vless": vless,
                    "hy2": hy2,
                },
            )

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

        return self._json(404, {"error": "not found"})


def main():
    os.makedirs(METRICS_DIR, exist_ok=True)
    httpd = HTTPServer((BIND, PORT), Handler)
    httpd.serve_forever()


if __name__ == "__main__":
    main()
