from fastapi import FastAPI, Request, HTTPException
import os, time, json, logging

from schemas import KeypointsInput, ScoreOutput
from .model import get_session, predict as _predict
logging.basicConfig(level=logging.INFO)

app = FastAPI()

@app.on_event("startup")
def preload_model():
    # 起動時にモデルをロードして cold-start を回避
    _ = get_session()

@app.post("/predict", response_model=ScoreOutput)
async def predict_endpoint(payload: KeypointsInput, request: Request) -> ScoreOutput:
    start = time.time()
    pts = [[pt.x, pt.y] for pt in payload.keypoints]
    score, sym, power, cons = _predict(pts)
    infer_ms = int((time.time() - start) * 1000)
    req_id = request.headers.get("x-request-id", "")
    model_uri = os.environ.get("_MODEL_URI", "")
    logging.info(json.dumps({
        "service": "ml-py",
        "request_id": req_id,
        "model_uri": model_uri,
        "input_len": len(pts),
        "infer_ms": infer_ms
    }))
    return ScoreOutput(score=score, symmetry=sym, power=power, consistency=cons)

@app.get("/healthz")
def healthz():
    # 単純生存確認
    return {"status": "ok"}

@app.get("/readiness")
def readiness():
    # モデルが読み込める状態かを検査（失敗なら 503）
    try:
        _ = get_session()
        return {"ready": True}
    except Exception:
        raise HTTPException(status_code=503, detail="not_ready")
