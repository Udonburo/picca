import hashlib
from pathlib import Path
import sys

sys.path.append(str(Path(__file__).resolve().parents[1]))

import numpy as np
import onnx
from fastapi.testclient import TestClient
from onnx import TensorProto, helper, numpy_helper
import pytest


def _make_const_model(path: Path) -> None:
    out = np.array([[0.5, 0.2, 0.3, 0.4]], dtype=np.float32)
    const = numpy_helper.from_array(out, name="const")
    node = helper.make_node("Constant", inputs=[], outputs=["output"], value=const)
    graph = helper.make_graph(
        [node],
        "test",
        [helper.make_tensor_value_info("input", TensorProto.FLOAT, [None, None])],
        [helper.make_tensor_value_info("output", TensorProto.FLOAT, [1, 4])],
    )
    opset = [helper.make_opsetid("", 17)]
    model = helper.make_model(
        graph, producer_name="test", ir_version=8, opset_imports=opset
    )
    onnx.save(model, path)


def test_predict(tmp_path, monkeypatch):
    model_path = tmp_path / "model.onnx"
    _make_const_model(model_path)
    sha = hashlib.sha256(model_path.read_bytes()).hexdigest()

    monkeypatch.setenv("_MODEL_URI", str(model_path))
    monkeypatch.setenv("_MODEL_SHA256", sha)

    from services.ml_py.model import get_session
    from services.ml_py.main import app

    get_session.cache_clear()
    client = TestClient(app)
    payload = {"keypoints": [{"x": 0.1, "y": 0.2}] * 100, "fps": 30}
    resp = client.post("/predict", json=payload)
    assert resp.status_code == 200
    data = resp.json()
    assert data["score"] == 50
    for key in ["symmetry", "power", "consistency"]:
        assert key in data
    assert data["symmetry"] == pytest.approx(0.2)
    assert data["power"] == pytest.approx(0.3)
    assert data["consistency"] == pytest.approx(0.4)
