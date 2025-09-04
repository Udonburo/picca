from __future__ import annotations

import hashlib
import os
from functools import lru_cache
from typing import Tuple

import gcsfs
import numpy as np
import onnxruntime as ort

from .utils.preproc import uniform_sample


def _load_model_bytes(uri: str) -> bytes:
    if uri.startswith("gs://"):
        fs = gcsfs.GCSFileSystem()
        with fs.open(uri, "rb") as f:
            return f.read()
    with open(uri, "rb") as f:
        return f.read()


@lru_cache(maxsize=1)
def get_session() -> ort.InferenceSession:
    # Read env variables at call-time so tests can monkeypatch them
    model_uri = os.getenv("_MODEL_URI", "model.onnx")
    model_sha = os.getenv("_MODEL_SHA256")

    data = _load_model_bytes(model_uri)
    if model_sha:
        digest = hashlib.sha256(data).hexdigest()
        if digest != model_sha:
            raise ValueError("model sha256 mismatch")
    return ort.InferenceSession(data, providers=["CPUExecutionProvider"])


def predict(arr: np.ndarray | list[list[float]]) -> Tuple[int, float, float, float]:
    session = get_session()
    in_shape = session.get_inputs()[0].shape
    input_size = in_shape[1]

    if input_size is None:
        target = 75
    else:
        target = input_size // 2
        
    sampled = uniform_sample(arr, target).astype(np.float32)
    flat    = sampled.reshape(1, -1)

    outputs = session.run(None, {session.get_inputs()[0].name: flat})[0][0]
    score   = int(np.clip(float(outputs[0]) * 100, 0, 100))
    metrics = np.clip(outputs[1:4], 0.0, 1.0)

    return (score, float(metrics[0]), float(metrics[1]), float(metrics[2]))

