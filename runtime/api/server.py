#!/usr/bin/env python3
"""FreshVPS admin API — session auth only."""
import json
import os
import subprocess
import time
from http.server import BaseHTTPRequestHandler, HTTPServer
from urllib.parse import urlparse

ETC = "/etc/freshvps"
SESS = os.path.join(ETC, "sessions")
CLIENTS = os.path.join(ETC, "clients")
VPN_BIN = "/usr/local/bin/freshvps-vpn"
BIND = os.environ.get("VPN_API_BIND", "127.0.0.1")
PORT = int(os.environ.get("VPN_API_PORT", "8787"))


def valid_session(token: str) -> bool:
    if not token or "/" in token or ".." in token:
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


class Handler(BaseHTTPRequestHandler):
    def log_message(self, fmt, *args):
        return

    def _auth(self) -> bool:
        auth = self.headers.get("Authorization", "")
        if auth.startswith("Bearer "):
            return valid_session(auth[7:].strip())
        return False

    def _json(self, code: int, obj):
        body = json.dumps(obj).encode()
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def _read(self):
        n = int(self.headers.get("Content-Length", 0))
        if n <= 0:
            return {}
        try:
            return json.loads(self.rfile.read(n).decode() or "{}")
        except json.JSONDecodeError:
            return {}

    def do_GET(self):
        u = urlparse(self.path)
        if u.path == "/health":
            return self._json(200, {"ok": True, "bind": BIND, "port": PORT})
        if not self._auth():
            return self._json(401, {"error": "unauthorized"})
        if u.path == "/vpn/users":
            try:
                out = subprocess.check_output([VPN_BIN, "list"], text=True, stderr=subprocess.STDOUT)
            except subprocess.CalledProcessError as e:
                return self._json(500, {"error": e.output or "list failed"})
            users = []
            for line in out.strip().splitlines():
                p = line.split("\t")
                if len(p) >= 3:
                    users.append({"name": p[0], "enabled": p[1] == "on", "uuid": p[2]})
            return self._json(200, {"users": users})
        if u.path.startswith("/vpn/users/") and u.path.endswith("/qr"):
            name = u.path.split("/")[3]
            path = os.path.join(CLIENTS, name, "qr.png")
            if not os.path.isfile(path):
                return self._json(404, {"error": "no qr"})
            with open(path, "rb") as f:
                data = f.read()
            self.send_response(200)
            self.send_header("Content-Type", "image/png")
            self.send_header("Content-Length", str(len(data)))
            self.end_headers()
            self.wfile.write(data)
            return
        if u.path.startswith("/vpn/users/") and u.path.endswith("/link"):
            name = u.path.split("/")[3]
            path = os.path.join(CLIENTS, name, "link.txt")
            if not os.path.isfile(path):
                return self._json(404, {"error": "no link"})
            with open(path, encoding="utf-8") as f:
                return self._json(200, {"link": f.read().strip()})
        if u.path == "/admin/client-config":
            return self._json(
                200,
                {
                    "api_base": f"http://{BIND}:{PORT}",
                    "note": "Bearer session; prefer 127.0.0.1 + SSH tunnel",
                },
            )
        return self._json(404, {"error": "not found"})

    def do_POST(self):
        u = urlparse(self.path)
        if not self._auth():
            return self._json(401, {"error": "unauthorized"})
        body = self._read()
        try:
            if u.path == "/vpn/users":
                name = (body.get("name") or "").strip()
                note = body.get("note") or ""
                if not name:
                    return self._json(400, {"error": "name required"})
                subprocess.check_call([VPN_BIN, "add", name, note])
                with open(os.path.join(CLIENTS, name, "link.txt"), encoding="utf-8") as f:
                    link = f.read().strip()
                return self._json(201, {"name": name, "link": link, "qr": f"/vpn/users/{name}/qr"})
            if u.path.startswith("/vpn/users/") and u.path.endswith("/disable"):
                name = u.path.split("/")[3]
                subprocess.check_call([VPN_BIN, "disable", name])
                return self._json(200, {"disabled": name})
            if u.path.startswith("/vpn/users/") and u.path.endswith("/enable"):
                name = u.path.split("/")[3]
                subprocess.check_call([VPN_BIN, "enable", name])
                return self._json(200, {"enabled": name})
        except subprocess.CalledProcessError as e:
            return self._json(500, {"error": str(e)})
        return self._json(404, {"error": "not found"})

    def do_DELETE(self):
        u = urlparse(self.path)
        if not self._auth():
            return self._json(401, {"error": "unauthorized"})
        if u.path.startswith("/vpn/users/"):
            name = u.path.rstrip("/").split("/")[-1]
            try:
                subprocess.check_call([VPN_BIN, "revoke", name])
            except subprocess.CalledProcessError as e:
                return self._json(500, {"error": str(e)})
            return self._json(200, {"revoked": name})
        return self._json(404, {"error": "not found"})


def main():
    httpd = HTTPServer((BIND, PORT), Handler)
    httpd.serve_forever()


if __name__ == "__main__":
    main()
