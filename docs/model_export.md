# Model Export

This document explains how to export an ONNX model and generate a SHA-256 hash.

```bash
# 1. ONNX 生成
python scripts/export_onnx.py --ckpt checkpoints/model.pt --out model.onnx

# 2. ハッシュ生成
bash scripts/hash_model.sh model.onnx

# 3. GCS へアップロード
gsutil cp model.onnx* gs://picca-models/models/dcv/v0.1.0/
```

To preview infrastructure changes before deploying:

```bash
terraform plan -var="project=<your-gcp-project>" -out=tfplan
```

To trigger the workflow via CLI:

```bash
gh workflow run ML\ Export --field ckpt_path=checkpoints/model.pt --field tag=v0.1.0
```

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
