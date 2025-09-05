FROM python:3.11-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PIP_NO_CACHE_DIR=1 \
    PYTHONPATH=/app

WORKDIR /app

# 依存を先に入れてレイヤー最適化
COPY services/ml_py/requirements.txt /app/requirements.txt
RUN pip install -r /app/requirements.txt

# サービス一式を確実にコピー（.gcloudignoreの影響を受けないよう絶対パス）
COPY services /app/services

# ビルド時に最低限の存在検証（早期失敗）
RUN python -c "import services, services.ml_py, onnxruntime"

EXPOSE 8080
# 非rootで実行（ビルドはrootで完了させてから切替）
RUN useradd -m app && chown -R app:app /app
USER app
# Shell 形式で ${PORT} を展開（exec 形式だと '$PORT' 文字列のままになり失敗する）
CMD uvicorn services.ml_py.main:app --host 0.0.0.0 --port ${PORT:-8080}
