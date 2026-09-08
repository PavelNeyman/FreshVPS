"""Edge device registry."""
from __future__ import annotations
import json, os, time, uuid
from pathlib import Path

EDGE_DIR = Path(os.environ.get("FRESHVPS_EDGE_DIR", "/var/lib/freshvps/edge"))
DEVICES = EDGE_DIR / "devices.json"
COMMANDS = EDGE_DIR / "commands.json"
TOKEN_FILE = Path("/etc/freshvps/secrets/edge_token")

def ensure():
    EDGE_DIR.mkdir(parents=True, exist_ok=True)
    if not DEVICES.is_file():
        DEVICES.write_text('{"devices":{}}' + chr(10))
    if not COMMANDS.is_file():
        COMMANDS.write_text('{"pending":[]}' + chr(10))

def edge_token() -> str:
    return TOKEN_FILE.read_text().strip() if TOKEN_FILE.is_file() else ""

def valid_edge_token(auth: str) -> bool:
    tok = edge_token()
    if not tok or not auth.startswith("Bearer "):
        return False
    return auth[7:].strip() == tok

def _load(path: Path) -> dict:
    ensure()
    try:
        return json.loads(path.read_text())
    except Exception:
        return {}

def _save(path: Path, data: dict) -> None:
    ensure()
    tmp = path.with_suffix(".tmp")
    tmp.write_text(json.dumps(data, indent=2) + chr(10))
    tmp.replace(path)

def heartbeat(payload: dict) -> dict:
    data = _load(DEVICES)
    devices = data.setdefault("devices", {})
    did = str(payload.get("device_id") or "unknown")
    devices[did] = {**payload, "last_seen": int(time.time())}
    _save(DEVICES, data)
    return {"ok": True}

def list_devices() -> list:
    data = _load(DEVICES)
    now = int(time.time())
    out = []
    for did, d in (data.get("devices") or {}).items():
        last = int(d.get("last_seen") or 0)
        row = dict(d)
        row["device_id"] = did
        row["healthy"] = (now - last) < 120
        out.append(row)
    return out

def enqueue_cmd(device_id: str, action: str, arg: str = "") -> str:
    data = _load(COMMANDS)
    pending = data.setdefault("pending", [])
    cid = uuid.uuid4().hex[:12]
    pending.append({"id": cid, "device_id": device_id, "action": action, "arg": arg, "created": int(time.time())})
    _save(COMMANDS, data)
    return cid

def poll_commands(device_id: str) -> list:
    data = _load(COMMANDS)
    pending = data.get("pending") or []
    mine = [c for c in pending if c.get("device_id") == device_id]
    data["pending"] = [c for c in pending if c.get("device_id") != device_id]
    _save(COMMANDS, data)
    return mine

def cmd_result(payload: dict) -> None:
    ensure()
    with (EDGE_DIR / "results.jsonl").open("a") as f:
        f.write(json.dumps(payload) + chr(10))
