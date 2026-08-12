# Model Export

This document explains how to export an ONNX model and generate a SHA-256 hash.

```bash
# 1. ONNX 生成
python scripts/export_onnx.py --ckpt checkpoints/model.pt --out model.onnx

# 2. ハッシュ生成
python scripts/hash_model.py model.onnx
```

The former GCS upload and GitHub Actions export workflow were removed when the
hosted GCP environment was retired. The infrastructure files remain only as
historical implementation evidence; see [`infra/README.md`](../infra/README.md)
for read-only validation.

The shell wrapper `bash scripts/hash_model.sh model.onnx` remains available for
POSIX environments and delegates to the same Python implementation.

## Predict API

The ML Python service exposes a single `/predict` endpoint. Input keypoints are
uniformly resampled to **75** frames using `uniform_sample` before being
flattened and fed into the ONNX model.

Example request:

```bash
curl -X POST http://localhost:8080/predict \
  -H 'Content-Type: application/json' \
  -d '{"keypoints":[{"x":0.1,"y":0.2}], "fps":30}'
```

Example response:

```json
{
  "score": 50,
  "symmetry": 0.2,
  "power": 0.3,
  "consistency": 0.4
}
```
