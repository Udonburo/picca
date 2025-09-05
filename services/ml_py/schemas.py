from __future__ import annotations

from typing import List, Optional

from pydantic import BaseModel


class XY(BaseModel):
    x: float
    y: float


class KeypointsInput(BaseModel):
    keypoints: List[XY]
    fps: int


class ScoreOutput(BaseModel):
    score: int
    symmetry: float
    power: float
    consistency: float
    analysis: Optional[dict] = None

