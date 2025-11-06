import argparse
import json
import os
import sys
import time
import urllib.request


def request_json(url: str, timeout: float):
    start = time.perf_counter()
    req = urllib.request.Request(url, headers={"Cache-Control": "no-store"})
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            ms = (time.perf_counter() - start) * 1000.0
            body = resp.read()
            payload = None
            if body:
                try:
                    payload = json.loads(body.decode("utf-8"))
                except json.JSONDecodeError:
                    payload = None
            return resp.getcode(), ms, payload
    except Exception as exc:
        ms = (time.perf_counter() - start) * 1000.0
        print(f"[bench] {exc}", file=sys.stderr)
        return 0, ms, None


def write_jsonl(path: str, rows):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w", encoding="utf-8") as fh:
        for row in rows:
            fh.write(json.dumps(row) + "\n")


def percentile(values, q):
    if not values:
        return None
    xs = sorted(values)
    idx = max(0, min(len(xs) - 1, int(round((len(xs) - 1) * q))))
    return xs[idx]


def main():
    parser = argparse.ArgumentParser(description="Probe /livez and record latency.")
    parser.add_argument("--base", required=True, help="Base URL for the API service.")
    parser.add_argument("--n", type=int, default=50, help="Number of requests.")
    parser.add_argument("--out", required=True, help="Output JSONL file.")
    parser.add_argument(
        "--summary",
        help="Optional summary JSON path (run_id, p95_ms, success, window, concurrency).",
    )
    parser.add_argument("--timeout", type=float, default=3.0, help="Request timeout.")
    parser.add_argument(
        "--concurrency", type=int, default=1, help="Documented concurrency for summary."
    )
    args = parser.parse_args()

    base = args.base.rstrip("/")
    status, _, payload = request_json(f"{base}/livez", args.timeout)
    run_id = (payload or {}).get("run_id", "") if status == 200 else ""

    rows = []
    for _ in range(max(args.n, 0)):
        status, ms, _ = request_json(f"{base}/livez", args.timeout)
        if status == 0:
            time.sleep(0.15)
            status, ms, _ = request_json(f"{base}/livez", args.timeout)
        rows.append(
            {
                "ts": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
                "run_id": run_id,
                "status": status,
                "ms": ms,
            }
        )
        time.sleep(0.05)

    write_jsonl(args.out, rows)

    if args.summary:
        successes = sum(1 for r in rows if r["status"] == 200)
        success_pct = (successes / len(rows) * 100.0) if rows else 0.0
        p95_val = percentile([r["ms"] for r in rows if r["status"] == 200], 0.95)
        summary = {
            "run_id": run_id,
            "samples": len(rows),
            "success_pct": round(success_pct, 2),
            "p95_ms": round(p95_val, 3) if p95_val is not None else None,
            "window": f"N={len(rows)} @ sequential",
            "concurrency": args.concurrency,
        }
        summary_dir = os.path.dirname(args.summary) or "."
        os.makedirs(summary_dir, exist_ok=True)
        with open(args.summary, "w", encoding="utf-8") as fh:
            json.dump(summary, fh, indent=2)


if __name__ == "__main__":
    main()
