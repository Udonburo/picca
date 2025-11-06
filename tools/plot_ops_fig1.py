import argparse
import json
import math
import os

import matplotlib.pyplot as plt


def load_rows(path):
    with open(path, "r", encoding="utf-8") as fh:
        return [json.loads(line) for line in fh if line.strip()]


def percentile(values, q):
    if not values:
        return None
    xs = sorted(values)
    idx = max(0, min(len(xs) - 1, int(math.ceil(q * len(xs)) - 1)))
    return xs[idx]


def main():
    parser = argparse.ArgumentParser(description="Render Ops Fig.1 summary.")
    parser.add_argument("--infile", required=True, help="Input JSONL from bench_picca.")
    parser.add_argument("--out", required=True, help="Output PNG path.")
    parser.add_argument("--window", default="N=50 @ ~3s", help="Window annotation.")
    parser.add_argument(
        "--concurrency", type=int, default=1, help="Concurrency annotation."
    )
    args = parser.parse_args()

    rows = load_rows(args.infile)
    successes = [r for r in rows if r.get("status") == 200]
    success_pct = (len(successes) / len(rows) * 100.0) if rows else 0.0
    p95_val = percentile([float(r.get("ms", 0.0)) for r in successes], 0.95)
    p95_display = f"{p95_val:.1f} ms" if p95_val is not None else "--"
    run_id = rows[0].get("run_id", "") if rows else ""

    fig = plt.figure(figsize=(6, 3.5))
    ax = fig.add_subplot(111)
    ax.axis("off")
    messages = [
        f"p95 Latency: {p95_display}",
        f"Success Rate: {success_pct:.1f}%",
        f"Samples: {len(rows)}",
        f"Run ID: {run_id}",
        f"Window: {args.window} · c={args.concurrency}",
    ]
    for i, text in enumerate(messages):
        ax.text(
            0.05,
            0.85 - i * 0.18,
            text,
            fontsize=16 if i < 3 else 13,
            ha="left",
            va="top",
        )
    ax.set_title("Fig.1 - Ops Slice", fontsize=16, loc="left")

    os.makedirs(os.path.dirname(args.out) or ".", exist_ok=True)
    fig.tight_layout()
    fig.savefig(args.out, dpi=160)
    plt.close(fig)


if __name__ == "__main__":
    main()
