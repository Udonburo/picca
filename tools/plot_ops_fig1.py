#!/usr/bin/env python3
import argparse, csv, json, os
import matplotlib.pyplot as plt

def read_jsonl(path):
    rows=[]
    with open(path,encoding="utf-8") as f:
        for line in f:
            if line.strip():
                rows.append(json.loads(line))
    return rows

def percentile(values, p):
    if not values: return None
    xs = sorted(values)
    k = max(0, min(len(xs)-1, int(len(xs)*p) - 1))
    return xs[k]

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--infile", required=True)
    ap.add_argument("--out-png", required=True)
    ap.add_argument("--out-csv", default=None)
    ap.add_argument("--cost-per-1k", type=float, default=0.0, help="¥ per 1k requests (liveness=0)")
    args = ap.parse_args()

    rows = read_jsonl(args.infile)
    lats = [r["latency_ms"] for r in rows if "latency_ms" in r]
    ok = sum(1 for r in rows if r.get("status")==200)
    n = len(rows)
    succ = (ok/n*100.0) if n>0 else 0.0
    p95 = percentile(lats, 0.95) if lats else 0.0
    cost = args.cost_per_1k  # /livez は0でOK

    # CSV
    out_csv = args.out_csv or os.path.splitext(args.out_png)[0] + ".csv"
    os.makedirs(os.path.dirname(out_csv) or ".", exist_ok=True)
    with open(out_csv,"w",newline="",encoding="utf-8") as f:
        w=csv.writer(f); w.writerow(["metric","value"])
        w.writerow(["p95_ms", f"{p95:.1f}"])
        w.writerow(["success_percent", f"{succ:.1f}"])
        w.writerow(["cost_per_1k_yen", f"{cost:.2f}"])

    # PNG（3本の水平バー）
    fig = plt.figure(figsize=(6,3.2))
    labels = ["p95 (ms)","Success (%)","Cost/1k (¥)"]
    vals = [p95, succ, cost]
    plt.barh(range(len(vals)), vals)
    plt.yticks(range(len(vals)), labels)
    plt.xlabel("value")
    plt.tight_layout()
    os.makedirs(os.path.dirname(args.out_png) or ".", exist_ok=True)
    plt.savefig(args.out_png, dpi=160)
    print(f"Wrote {args.out_png} and {out_csv}")

if __name__ == "__main__":
    main()
