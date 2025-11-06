import hashlib
import pathlib
import sys


def short_sha1(path: str) -> str:
    p = pathlib.Path(path)
    h = hashlib.sha1(p.read_bytes()).hexdigest()
    return h[:8]


def run_id(script_path: str, model_ckpt_sha: str) -> str:
    return f"{short_sha1(script_path)}·{model_ckpt_sha}"


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("usage: python tools/run_id.py <script_path> <model_ckpt_sha>")
        sys.exit(1)
    if hasattr(sys.stdout, "reconfigure"):
        sys.stdout.reconfigure(encoding="utf-8")
    print(run_id(sys.argv[1], sys.argv[2]))
