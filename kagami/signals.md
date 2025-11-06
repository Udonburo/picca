# KAGAMI Signals (WHAT Only)

## Run ID
- Format: `<script_sha8>·<model_ckpt>`.
- Derived from env vars `PICCA_SCRIPT_SHA` and `MODEL_CKPT_SHA` with defaults (`dev`, `na`).
- Surfaces: `/healthz`, `/livez`, `/readyz`, API responses, stdout OTS logs, Fig.1 caption.

## Observation Trace Sheet (OTS)
- Emission: JSONL via `services/api-go`/`services/ml_py` stdout, one line per HTTP request.
- Keys (fixed order): `ts`, `run_id`, `path`, `status`, `latency_ms`, `req_id`, `input_hash`, `output_hash`.
- Timestamp: RFC3339Nano (UTC); hashes are blank when body exceeds 1 MiB or is unavailable.
- Storage/relay: stdout -> Cloud Logging (no additional sinks by default).

## SLO Declaration
- Public WHAT lives in `ops/lachesis/slo.yaml` (targets + windows).
- Metrics tracked: `p95_ms`, `success`, `cost_per_1k_yen` (cost formula documented separately).
- Observation windows: latency/success = rolling 15 min, cost = daily notebook rollup.

## Figure 1 Legend Norms
- Required text elements: `p95(ms)`, `Success(%)`, `Samples`, `Run ID`, `Window (hits · concurrency)`.
- Legend/footer must include window, concurrency, and run_id exactly as produced by tooling.
- Panel is text-only (no axes), white background, primary metrics at >=18 pt font.

## Signal Catalog v0.3.9
- **S1 Run ID** - propagate `<script_sha8>·<model_ckpt>` across surfaces.
- **S2 OTS** - stdout JSONL adhering to fixed schema for replay/audit.
- **S3 Ops SLO** - publish targets/windows (HOW for tuning remains private).
- **S4 Fig.1 Ops Slice** - generated via `tools/plot_ops_fig1.py` following legend norms.
- **S5 Runbook** - README + `docs/runbook.md` describe the minimal bench-to-figure path.
