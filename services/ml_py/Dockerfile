FROM python:3.11-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PORT=8080

WORKDIR /app

# まずパッケージルートを丸ごとコピー（←これが肝）
COPY services /app/services

# 依存インストール：requirements があればそれ、無ければ標準セット
RUN set -eux; \
    if [ -f /app/services/ml_py/requirements.txt ]; then \
      pip install --no-cache-dir -r /app/services/ml_py/requirements.txt; \
    else \
      pip install --no-cache-dir \
        fastapi "uvicorn[standard]" onnxruntime numpy gcsfs pydantic; \
    fi

EXPOSE 8080
CMD ["sh","-c","uvicorn services.ml_py.main:app --host 0.0.0.0 --port ${PORT:-8080}"]
