FROM python:3.11-slim

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1

WORKDIR /app
COPY . .

RUN pip install --no-cache-dir fastapi uvicorn[standard] onnxruntime numpy gcsfs pydantic

EXPOSE 8080
CMD ["uvicorn", "services.ml_py.main:app", "--host", "0.0.0.0", "--port", "8080"]
