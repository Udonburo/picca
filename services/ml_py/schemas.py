from __future__ import annotations

from pydantic import BaseModel


class XY(BaseModel):
    x: float
    y: float


class KeypointsInput(BaseModel):
    keypoints: list[XY]
    fps: int


class ScoreOutput(BaseModel):
    score: int
    symmetry: float
    power: float
    consistency: float
    analysis: dict | None = None

