# services/ml-py/main.py
from fastapi import FastAPI, Request

app = FastAPI()

@app.post("/predict")
async def predict(_: Request):
    # ↳ Task-6 で本物のモデルに置換
    return {"score": 0.87}
