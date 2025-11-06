import hashlib
import json
import logging
import os
import time
from fastapi import FastAPI, HTTPException, Request, Response

from .model import get_session, predict as _predict
from .schemas import KeypointsInput, ScoreOutput

logging.basicConfig(level=logging.INFO)

RUN_ID = f"{os.getenv('PICCA_SCRIPT_SHA', 'dev')}·{os.getenv('MODEL_CKPT_SHA', 'na')}"
MAX_CAPTURE_BYTES = 1 << 20  # 1 MiB

app = FastAPI()


def sha16(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()[:16]


def hash_or_blank(data: bytes) -> str:
    if not data or len(data) > MAX_CAPTURE_BYTES:
        return ""
    return sha16(data)


@app.middleware("http")
async def ots_middleware(request: Request, call_next):
    t0 = time.perf_counter()
    body = await request.body()
    in_hash = hash_or_blank(body)

    async def receive():
        return {"type": "http.request", "body": body, "more_body": False}

    request_with_body = Request(request.scope, receive=receive)
    response: Response = await call_next(request_with_body)
    latency_ms = round((time.perf_counter() - t0) * 1000, 3)

    req_id = response.headers.get("x-request-id") or request.headers.get("x-request-id") or ""
    out_hash = ""
    try:
        if getattr(response, "body_iterator", None) is None:
            payload = getattr(response, "body", b"") or b""
            if isinstance(payload, (bytes, bytearray)):
                out_hash = hash_or_blank(bytes(payload))
    except Exception:
        out_hash = ""

    print(
        json.dumps(
            {
                "ts": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
                "run_id": RUN_ID,
                "path": request.url.path,
                "status": getattr(response, "status_code", 0),
                "latency_ms": latency_ms,
                "req_id": req_id,
                "input_hash": in_hash,
                "output_hash": out_hash,
            }
        )
    )
    return response


@app.on_event("startup")
def preload_model():
    _ = get_session()


@app.post("/predict", response_model=ScoreOutput)
async def predict_endpoint(payload: KeypointsInput, request: Request) -> ScoreOutput:
    start = time.perf_counter()
    pts = [[pt.x, pt.y] for pt in payload.keypoints]
    score, sym, power, cons = _predict(pts)
    infer_ms = round((time.perf_counter() - start) * 1000, 3)
    req_id = request.headers.get("x-request-id", "")
    model_uri = os.environ.get("_MODEL_URI", "")
    logging.info(
        json.dumps(
            {
                "service": "ml-py",
                "request_id": req_id,
                "model_uri": model_uri,
                "input_len": len(pts),
                "infer_ms": infer_ms,
            }
        )
    )
    return ScoreOutput(score=score, symmetry=sym, power=power, consistency=cons)


@app.get("/healthz")
def healthz():
    return {"ok": True, "run_id": RUN_ID}


@app.get("/livez")
def livez():
    return {"ok": True, "run_id": RUN_ID}


@app.get("/readiness")
def readiness():
    try:
        _ = get_session()
        return {"ready": True, "run_id": RUN_ID}
    except Exception:
        raise HTTPException(status_code=503, detail="not_ready")
