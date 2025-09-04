from fastapi.testclient import TestClient
from services.ml_py import main


def test_healthz_only(monkeypatch):
    # readinessの起動前プリロードで get_session() が走るのを無害化
    monkeypatch.setattr(main, "get_session", lambda: object())
    app = main.app
    with TestClient(app) as client:
        r = client.get("/healthz")
        assert r.status_code == 200

