from __future__ import annotations

import argparse
import hashlib
from pathlib import Path


def write_sha256(model_path: Path) -> Path:
    digest = hashlib.sha256(model_path.read_bytes()).hexdigest()
    output_path = model_path.with_name(f"{model_path.name}.sha256")
    output_path.write_text(f"{digest}\n", encoding="utf-8")
    return output_path


def main() -> None:
    parser = argparse.ArgumentParser(description="Write a SHA-256 file for a model.")
    parser.add_argument("model", type=Path, help="Path to the model artifact.")
    args = parser.parse_args()

    output_path = write_sha256(args.model)
    print(f"SHA256: {output_path.read_text(encoding='utf-8').strip()}")


if __name__ == "__main__":
    main()
