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
