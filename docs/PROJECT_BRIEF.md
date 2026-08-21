# Picca — Project Brief

> Archived engineering log · Motion-scoring proof of concept · 2025

Picca explored one end-to-end question: can a short sequence of body keypoints be turned into feedback that is immediate, reproducible, and explainable? Its central design choice was to keep artifact-verified numerical scoring separate from generative explanation. The repository preserves the resulting prototype and the engineering decisions behind it; its hosted services have been retired.

## At a glance

| | |
| --- | --- |
| **Input** | Frame rate and a time-ordered sequence of 2D keypoints |
| **Output** | Overall score plus symmetry, power, and consistency indicators; optional natural-language explanation |
| **System** | Separate score and explain API paths over ONNX Runtime and Vertex AI / Gemini |
| **Engineering scope** | API boundaries, model packaging, deployment, observability, and CI |
| **Current status** | Archived learning record; no hosted demo or production claims |

## System boundary

```mermaid
flowchart LR
    Input["Pose keypoints"] -->|"POST /api/v1/score"| Score["Go + Gin<br/>Score route"]
    Score --> Inference["FastAPI + ONNX Runtime<br/>SHA-256 verified artifact"]
    Inference --> Metrics["Score + 3 metrics"]
    Metrics --> Result["Result UI"]
    Metrics -.->|"POST /api/v1/explain<br/>metrics only"| Explain["Go + Gin<br/>Explain route"]
    Explain -.-> Gemini["Vertex AI / Gemini<br/>Natural-language explanation"]
```

The score route delegates artifact-bound numerical inference to FastAPI / ONNX Runtime. The explain route accepts only the four resulting values and asks Vertex AI / Gemini to express them in natural language; it does not determine the score or receive raw keypoints. The default configured model is `gemini-2.5-flash-lite`.

This separation keeps the numerical result useful without a generative dependency and allows the explanation model or prompt to change without redefining the measurement. The broader three-service split made responsibilities explicit, but it also introduced contracts, environment configuration, timeout handling, and cross-service debugging.

## What is implemented

- A static Next.js archive page, a clearly labeled result mock, and API routes for forwarding motion data.
- A Go gateway with distinct `/api/v1/score` and `/api/v1/explain` routes, API-key authentication, request IDs, upstream timeouts, health checks, and graceful shutdown.
- A FastAPI inference service with keypoint preprocessing, ONNX Runtime execution, model SHA-256 verification, and typed numerical responses.
- A Vertex AI / Gemini explanation path that receives the four numerical outputs and returns a short natural-language summary.
- Docker, Terraform, and Cloud Build experiments, plus SLO, runbook, structured logging, and benchmark tooling.
- Automated checks for the web, Python, Go, workflow conventions, and secret leakage.

## Engineering takeaways

| Decision | What it revealed | What I would do now |
| --- | --- | --- |
| Separate web, gateway, and inference services | Clear ownership comes with more contracts and failure points | Start with two boundaries and split only when scaling or security requires it |
| Use ONNX as the inference boundary | A model file is not reproducible without preprocessing and schema versions | Package artifact, hash, schema, and preprocessing version as one manifest |
| Separate scoring from generative explanation | Artifact integrity and flexible wording have different reliability requirements | Version the score schema and prompt independently, with a non-generative fallback |
| Add operations after the happy path | A working response does not explain where or why a request failed | Define request IDs, health semantics, and minimum logs at the start |

## Evidence map

- [`src/app/`](../src/app/) — archive UI, static result mock, and API routes
- [`services/api-go/main.go`](../services/api-go/main.go#L131-L143) — distinct score / explain routes and Vertex AI handling
- [`services/ml_py/model.py`](../services/ml_py/model.py#L23-L54) — model hash verification and ONNX inference
- [`tests/`](../tests/) — export and prediction smoke tests
- [`infra/`](../infra/) — archived infrastructure experiments
- [`ops/lachesis/`](../ops/lachesis/) — SLO and runbook
- [`RETROSPECTIVE.md`](RETROSPECTIVE.md) — detailed decisions, failures, and lessons

---

This brief describes implemented repository evidence, not the broader goals contained in the original hackathon materials.
