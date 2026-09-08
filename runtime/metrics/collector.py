#!/usr/bin/env python3
"""Append one metrics sample to history.jsonl (lightweight, no Prometheus)."""
import json
import os
import time

# import collect from server if path set
import importlib.util
import sys

ROOT = os.environ.get("FRESHVPS_API_ROOT", "/opt/freshvps/runtime/api")
sys.path.insert(0, ROOT)
spec = importlib.util.spec_from_file_location("server", os.path.join(ROOT, "server.py"))
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)

METRICS_DIR = os.environ.get("FRESHVPS_METRICS_DIR", "/var/lib/freshvps/metrics")
os.makedirs(METRICS_DIR, exist_ok=True)
m = mod.collect_metrics()
mem = m.get("mem") or {}
total = mem.get("total") or 1
used = mem.get("used") or 0
m["mem_pct"] = round(100.0 * used / total, 1)
path = os.path.join(METRICS_DIR, "history.jsonl")
with open(path, "a", encoding="utf-8") as f:
    f.write(json.dumps(m, separators=(",", ":")) + "\n")
# keep last ~7 days at 1/min ≈ 10080 lines; trim to 10000
try:
    with open(path, encoding="utf-8") as f:
        lines = f.readlines()
    if len(lines) > 10000:
        with open(path, "w", encoding="utf-8") as f:
            f.writelines(lines[-10000:])
except OSError:
    pass
