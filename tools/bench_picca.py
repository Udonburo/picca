#!/usr/bin/env python3
import argparse, hashlib, json, os, subprocess, time
from datetime import datetime
from urllib.request import Request, urlopen

def git_rev_short():
    try:
        return subprocess.check_output(["git","rev-parse","--short","HEAD"], timeout=3).decode().strip()
    except Exception:
        return "nogit"

def script_hash():
    p = __file__
    try:
        with open(p,'rb') as f: h = hashlib.sha256(f.read()).hexdigest()[:8]
        return h
    except Exception:
        return "nohash"

def fetch(url, timeout=3.0):
    t0 = time.perf_counter()
    code = 0
    try:
        r = urlopen(Request(url, headers={"Cache-Control":"no-store"}), timeout=timeout)
        code = getattr(r, "status", 200)
    except Exception:
        code = 0
    dt = (time.perf_counter() - t0) * 1000.0
    return code, dt

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("-b","--base", default="http://localhost:8080")
    ap.add_argument("-n","--num", type=int, default=50)
    ap.add_argument("-o","--out", required=True, help="output JSONL path")
    args = ap.parse_args()

    run_id = f"{script_hash()}·picca@{git_rev_short()}"
    livez = args.base.rstrip("/") + "/livez"
    os.makedirs(os.path.dirname(args.out), exist_ok=True)

    ok = 0
    lats = []
    with open(args.out, "a", encoding="utf-8") as f:
        for i in range(args.num):
            code, ms = fetch(livez)
            row = {
                "ts": datetime.utcnow().isoformat()+"Z",
                "status": code,
                "latency_ms": ms,
                "endpoint": "/livez",
                "run_id": run_id,
                "idx": i
            }
            f.write(json.dumps(row, ensure_ascii=False) + "\n")
            lats.append(ms)
            if code == 200: ok += 1
            time.sleep(0.05)

    p95 = sorted(lats)[max(0, int(len(lats)*0.95)-1)] if lats else None
    succ = ok/len(lats) if lats else 0.0
    print(f"run_id={run_id}\ncount={len(lats)} ok={ok} success={succ:.3f} p95_ms={p95:.1f}" )

if __name__ == "__main__":
    main()
