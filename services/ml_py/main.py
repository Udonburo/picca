from fastapi import FastAPI

from schemas import KeypointsInput, ScoreOutput
from .model import predict

app = FastAPI()


@app.post("/predict", response_model=ScoreOutput)
async def predict_endpoint(payload: KeypointsInput) -> ScoreOutput:
    score, sym, power, cons = predict([[pt.x, pt.y] for pt in payload.keypoints])
    return ScoreOutput(score=score, symmetry=sym, power=power, consistency=cons)
