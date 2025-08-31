FROM python:3.11-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PORT=8080

WORKDIR /app

# 依存
COPY services/ml_py/requirements.txt /app/requirements.txt
RUN if [ -f /app/requirements.txt ]; then \
      pip install --no-cache-dir -r /app/requirements.txt ; \
    else \
      pip install --no-cache-dir fastapi "uvicorn[standard]" onnxruntime numpy gcsfs pydantic ; \
    fi

# パッケージルート "services" を丸ごとコピー（←これが肝）
COPY services /app/services

EXPOSE 8080
CMD ["sh","-c","uvicorn services.ml_py.main:app --host 0.0.0.0 --port ${PORT:-8080}"]
