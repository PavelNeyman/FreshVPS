#!/usr/bin/env bash
# Module: VPN-only admin API (127.0.0.1 + optional VPN bind)
# shellcheck disable=SC2154

module_vpn_api_install() {
  pkg_install python3

  local port="${VPN_API_PORT:-8787}"
  local bind="${VPN_API_BIND:-127.0.0.1}"

  mkdir -p /opt/freshvps-api /etc/freshvps/sessions
  chmod 700 /etc/freshvps/sessions

  # Master token for minting is separate; sessions via CLI
  if [[ ! -f "${FRESHVPS_ETC}/secrets/api_master_token" ]]; then
    write_secret api_master_token "$(openssl rand -hex 32)"
  fi

  cat >/opt/freshvps-api/server.py <<'PY'
#!/usr/bin/env python3
"""FreshVPS admin API — session auth, localhost/VPN bind only."""
import json, os, subprocess, time
from http.server import BaseHTTPRequestHandler, HTTPServer
from urllib.parse import urlparse, parse_qs

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
        exp = int(open(path).read().strip())
    except Exception:
        return False
    if exp < int(time.time()):
        try:
            os.remove(path)
        except OSError:
            pass
        return False
    return True

class H(BaseHTTPRequestHandler):
    def log_message(self, fmt, *args):
        pass

    def _auth(self):
        auth = self.headers.get("Authorization", "")
        if auth.startswith("Bearer "):
            return valid_session(auth[7:].strip())
        return False

    def _json(self, code, obj):
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
        return json.loads(self.rfile.read(n).decode() or "{}")

    def do_GET(self):
        u = urlparse(self.path)
        if u.path == "/health":
            return self._json(200, {"ok": True})
        if not self._auth():
            return self._json(401, {"error": "unauthorized"})
        if u.path == "/vpn/users":
            out = subprocess.check_output([VPN_BIN, "list"], text=True)
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
            data = open(path, "rb").read()
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
            return self._json(200, {"link": open(path).read().strip()})
        if u.path == "/admin/shortcut":
            # serve template if present
            p = "/opt/freshvps/shortcuts/FreshVPS-Admin.shortcut"
            if not os.path.isfile(p):
                return self._json(404, {"error": "template missing — see docs/SHORTCUT-IOS.md"})
            data = open(p, "rb").read()
            self.send_response(200)
            self.send_header("Content-Type", "application/octet-stream")
            self.send_header("Content-Disposition", "attachment; filename=FreshVPS-Admin.shortcut")
            self.send_header("Content-Length", str(len(data)))
            self.end_headers()
            self.wfile.write(data)
            return
        if u.path == "/admin/client-config":
            return self._json(200, {
                "api_base": f"http://{BIND}:{PORT}",
                "note": "Use only over operator VPN; store session token in local file"
            })
        self._json(404, {"error": "not found"})

    def do_POST(self):
        u = urlparse(self.path)
        if not self._auth():
            return self._json(401, {"error": "unauthorized"})
        body = self._read()
        if u.path == "/vpn/users":
            name = body.get("name", "").strip()
            note = body.get("note", "")
            if not name:
                return self._json(400, {"error": "name required"})
            subprocess.check_call([VPN_BIN, "add", name, note])
            link = open(os.path.join(CLIENTS, name, "link.txt")).read().strip()
            return self._json(201, {"name": name, "link": link, "qr": f"/vpn/users/{name}/qr"})
        if u.path.startswith("/vpn/users/") and u.path.endswith("/disable"):
            name = u.path.split("/")[3]
            subprocess.check_call([VPN_BIN, "disable", name])
            return self._json(200, {"disabled": name})
        if u.path.startswith("/vpn/users/") and u.path.endswith("/enable"):
            name = u.path.split("/")[3]
            subprocess.check_call([VPN_BIN, "enable", name])
            return self._json(200, {"enabled": name})
        self._json(404, {"error": "not found"})

    def do_DELETE(self):
        u = urlparse(self.path)
        if not self._auth():
            return self._json(401, {"error": "unauthorized"})
        if u.path.startswith("/vpn/users/"):
            name = u.path.rstrip("/").split("/")[-1]
            subprocess.check_call([VPN_BIN, "revoke", name])
            return self._json(200, {"revoked": name})
        self._json(404, {"error": "not found"})

def main():
    httpd = HTTPServer((BIND, PORT), H)
    httpd.serve_forever()

if __name__ == "__main__":
    main()
PY
  chmod 755 /opt/freshvps-api/server.py

  cat >/etc/systemd/system/freshvps-api.service <<EOF
[Unit]
Description=FreshVPS admin API (VPN/localhost only)
After=network-online.target

[Service]
Type=simple
Environment=VPN_API_BIND=${bind}
Environment=VPN_API_PORT=${port}
ExecStart=/usr/bin/python3 /opt/freshvps-api/server.py
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

  systemd_enable_start freshvps-api
  # Do not open WAN; bind defaults to 127.0.0.1 — access via SSH tunnel or VPN IP if bind changed
  info "API on ${bind}:${port} — mint session: freshvps-vpn session 72"
  info "Shortcuts: Face ID → read token file → Authorization: Bearer <session>"
}

module_vpn_api_uninstall() {
  systemctl disable --now freshvps-api 2>/dev/null || true
  rm -f /etc/systemd/system/freshvps-api.service
  systemctl daemon-reload 2>/dev/null || true
}
