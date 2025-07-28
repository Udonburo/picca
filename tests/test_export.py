import subprocess
from pathlib import Path
import onnx


def test_export_and_hash(tmp_path: Path):
    ckpt = tmp_path / "model.pt"
    out = tmp_path / "model.onnx"

    subprocess.run([
        "python",
        "scripts/export_onnx.py",
        "--ckpt",
        str(ckpt),
        "--out",
        str(out),
    ], check=True)

    model = onnx.load(out)
    onnx.checker.check_model(model)

    subprocess.run(["bash", "scripts/hash_model.sh", str(out)], check=True)
    sha = Path(str(out) + ".sha256")
    assert sha.exists()
