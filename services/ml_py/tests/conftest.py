"""Pytest configuration for ml_py tests.

Ensures the repository root is on `sys.path` so imports like
`from services.ml_py import main` work regardless of the working
directory used to invoke pytest in CI.
"""

from __future__ import annotations

import sys
from pathlib import Path


def _ensure_repo_root_on_syspath() -> None:
    # tests/ -> ml_py/ -> services/ -> repo root
    repo_root = Path(__file__).resolve().parents[3]
    repo_root_str = str(repo_root)
    if repo_root_str not in sys.path:
        sys.path.insert(0, repo_root_str)


_ensure_repo_root_on_syspath()

