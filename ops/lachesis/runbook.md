# Picca  Ops Slice (LACHESIS)
## Run-ID
run_id = <script_hash8><model_ckpt@sha>  （例: 9ac1beefllama3@f0c1d2e）
- script_hash8: 計測スクリプト（tools/fig1_delta_avail.py）のSHA1先頭8桁
- model_ckpt@sha: 推論モデルの識別子

## 収集
- ルート: /predict or /score を固定条件で30+リクエスト/条件
- ログ: logs/lachesis/YYYYMMDD/*.jsonl （OTS準拠）

## 図1
- H_U (bits or norm) vs Δ_avail (JS base-2, ε=1e-12)
- Isotonic + Theil–Sen, Spearman ρ + block-perm p (B≥10,000*, early-stop可)
