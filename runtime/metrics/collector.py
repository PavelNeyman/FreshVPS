#!/usr/bin/env python3
"""FreshVPS metrics tick: sample + probes + Telegram alerts (no Prometheus)."""
from __future__ import annotations

import fcntl
import importlib.util
import json
import os
import socket
import subprocess
import sys
import time
import urllib.error
import urllib.request

ROOT = os.environ.get("FRESHVPS_API_ROOT", "/opt/freshvps/runtime/api")
METRICS_DIR = os.environ.get("FRESHVPS_METRICS_DIR", "/var/lib/freshvps/metrics")
ETC = "/etc/freshvps"
PROBES_CFG = os.path.join(ETC, "probes.json")
NOTIFY = "/opt/freshvps/runtime/telegram/notify.sh"
STATE_PATH = os.path.join(METRICS_DIR, "alert_state.json")
LOCK_PATH = os.path.join(METRICS_DIR, "collector.lock")

sys.path.insert(0, ROOT)
spec = importlib.util.spec_from_file_location("server", os.path.join(ROOT, "server.py"))
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)

os.makedirs(METRICS_DIR, exist_ok=True)


def default_probes_cfg() -> dict:
    return {
        "probes": [
            {"name": "dns-blocky", "type": "tcp", "host": "127.0.0.1", "port": 53, "timeout": 2},
            {"name": "vless", "type": "tcp", "host": "127.0.0.1", "port": 443, "timeout": 2},
            # Hysteria2 is UDP — TCP probe always fails with connection refused
            {"name": "hy2", "type": "udp", "host": "127.0.0.1", "port": 8443, "timeout": 2},
            {"name": "api-health", "type": "http", "url": "http://127.0.0.1:8787/health", "timeout": 3},
        ],
        "alerts": {
            "cpu_pct": 90,
            "mem_pct": 92,
            "disk_pct": 90,
            "service_not_active": True,
            "probe_fail": True,
            "cooldown_sec": 1800,
        },
    }


def load_cfg() -> dict:
    if os.path.isfile(PROBES_CFG):
        try:
            with open(PROBES_CFG, encoding="utf-8") as f:
                cfg = json.load(f)
            # migrate legacy hy2 tcp → udp
            changed = False
            for p in cfg.get("probes") or []:
                if p.get("name") == "hy2" and p.get("type") == "tcp":
                    p["type"] = "udp"
                    changed = True
            if changed:
                with open(PROBES_CFG, "w", encoding="utf-8") as f:
                    json.dump(cfg, f, indent=2)
                    f.write("\n")
            return cfg
        except (OSError, json.JSONDecodeError):
            pass
    return default_probes_cfg()


def _ss_listening(port: int, udp: bool) -> bool:
    """True if local socket listens on port (IPv4 or IPv6)."""
    flag = "-ulnH" if udp else "-tlnH"
    try:
        r = subprocess.run(
            ["ss", flag, "sport", "=", f":{port}"],
            capture_output=True,
            text=True,
            timeout=3,
        )
        return bool((r.stdout or "").strip())
    except (OSError, subprocess.TimeoutExpired):
        return False


def probe_tcp(host: str, port: int, timeout: float) -> dict:
    t0 = time.time()
    try:
        with socket.create_connection((host, port), timeout=timeout):
            return {"ok": True, "ms": round((time.time() - t0) * 1000, 1)}
    except OSError as e:
        # fallback: process may listen on :: only
        if _ss_listening(port, udp=False):
            return {"ok": True, "ms": round((time.time() - t0) * 1000, 1), "via": "ss"}
        return {"ok": False, "ms": round((time.time() - t0) * 1000, 1), "error": str(e)}


def probe_udp(host: str, port: int, timeout: float) -> dict:
    """Hysteria2 etc.: check UDP listen (ss), optional sendto."""
    t0 = time.time()
    if _ss_listening(port, udp=True):
        return {"ok": True, "ms": round((time.time() - t0) * 1000, 1), "via": "ss"}
    # try bind-free send (won't prove server, but complements ss miss)
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.settimeout(timeout)
        s.sendto(b"\x00", (host, port))
        s.close()
    except OSError:
        pass
    return {
        "ok": False,
        "ms": round((time.time() - t0) * 1000, 1),
        "error": f"UDP :{port} not listening",
    }


def probe_http(url: str, timeout: float) -> dict:
    t0 = time.time()
    try:
        req = urllib.request.Request(url, method="GET")
        with urllib.request.urlopen(req, timeout=timeout) as r:
            code = r.getcode()
            return {"ok": 200 <= code < 400, "ms": round((time.time() - t0) * 1000, 1), "code": code}
    except (urllib.error.URLError, OSError) as e:
        return {"ok": False, "ms": round((time.time() - t0) * 1000, 1), "error": str(e)}


def run_probes(cfg: dict) -> list:
    out = []
    for p in cfg.get("probes") or []:
        name = p.get("name") or "probe"
        timeout = float(p.get("timeout") or 3)
        typ = (p.get("type") or "tcp").lower()
        if typ == "http":
            res = probe_http(str(p.get("url") or ""), timeout)
        elif typ == "udp":
            res = probe_udp(str(p.get("host") or "127.0.0.1"), int(p.get("port") or 0), timeout)
        else:
            res = probe_tcp(str(p.get("host") or "127.0.0.1"), int(p.get("port") or 0), timeout)
        res["name"] = name
        res["type"] = typ
        out.append(res)
    return out


def load_state() -> dict:
    if os.path.isfile(STATE_PATH):
        try:
            with open(STATE_PATH, encoding="utf-8") as f:
                return json.load(f)
        except (OSError, json.JSONDecodeError):
            pass
    return {"last": {}}


def save_state(st: dict) -> None:
    tmp = STATE_PATH + ".tmp"
    with open(tmp, "w", encoding="utf-8") as f:
        json.dump(st, f)
    os.replace(tmp, STATE_PATH)


def notify(msg: str) -> None:
    if not os.path.isfile(NOTIFY):
        return
    try:
        subprocess.run([NOTIFY, msg], timeout=15, check=False)
    except (OSError, subprocess.TimeoutExpired):
        pass


def maybe_alert(key: str, msg: str, cooldown: int, st: dict) -> None:
    now = int(time.time())
    last = int((st.get("last") or {}).get(key) or 0)
    if now - last < cooldown:
        return
    # record cooldown FIRST so concurrent runners skip
    st.setdefault("last", {})[key] = now
    save_state(st)
    notify(msg)


def evaluate_alerts(m: dict, probes: list, cfg: dict, st: dict) -> None:
    al = cfg.get("alerts") or {}
    cd = int(al.get("cooldown_sec") or 1800)
    cpu = float(m.get("cpu_pct") or 0)
    mem_pct = float(m.get("mem_pct") or 0)
    disk_pct = float(m.get("disk_pct") or 0)

    if cpu >= float(al.get("cpu_pct") or 90):
        maybe_alert("cpu", f"⚠️ FreshVPS CPU {cpu}%", cd, st)
    if mem_pct >= float(al.get("mem_pct") or 92):
        maybe_alert("mem", f"⚠️ FreshVPS RAM {mem_pct}%", cd, st)
    if disk_pct >= float(al.get("disk_pct") or 90):
        maybe_alert("disk", f"⚠️ FreshVPS disk {disk_pct}%", cd, st)

    if al.get("service_not_active", True):
        for name, status in (m.get("services") or {}).items():
            if status != "active":
                maybe_alert(f"svc:{name}", f"🔴 Service {name}: {status}", cd, st)

    if al.get("probe_fail", True):
        for p in probes:
            if not p.get("ok"):
                maybe_alert(
                    f"probe:{p.get('name')}",
                    f"🔴 Probe {p.get('name')} failed: {p.get('error') or p.get('code') or 'down'}",
                    cd,
                    st,
                )


def main() -> None:
    # single-flight lock
    lock_f = open(LOCK_PATH, "w")
    try:
        fcntl.flock(lock_f.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)
    except BlockingIOError:
        lock_f.close()
        return

    try:
        cfg = load_cfg()
        if not os.path.isfile(PROBES_CFG):
            os.makedirs(ETC, exist_ok=True)
            with open(PROBES_CFG, "w", encoding="utf-8") as f:
                json.dump(cfg, f, indent=2)
                f.write("\n")

        m = mod.collect_metrics()
        mem = m.get("mem") or {}
        total = mem.get("total") or 1
        used = mem.get("used") or 0
        m["mem_pct"] = round(100.0 * used / total, 1)
        disk = m.get("disk") or {}
        dt = disk.get("total") or 1
        m["disk_pct"] = round(100.0 * (disk.get("used") or 0) / dt, 1)

        probes = run_probes(cfg)
        m["probes"] = probes

        path = os.path.join(METRICS_DIR, "history.jsonl")
        with open(path, "a", encoding="utf-8") as f:
            f.write(json.dumps(m, separators=(",", ":")) + "\n")
        try:
            with open(path, encoding="utf-8") as f:
                lines = f.readlines()
            if len(lines) > 10000:
                with open(path, "w", encoding="utf-8") as f:
                    f.writelines(lines[-10000:])
        except OSError:
            pass

        with open(os.path.join(METRICS_DIR, "latest.json"), "w", encoding="utf-8") as f:
            json.dump(m, f)

        st = load_state()
        evaluate_alerts(m, probes, cfg, st)
        save_state(st)
    finally:
        fcntl.flock(lock_f.fileno(), fcntl.LOCK_UN)
        lock_f.close()


if __name__ == "__main__":
    main()
