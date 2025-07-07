# ── ci.tf ────────────────────────────────────────────────────────────────
# Cloud Build Trigger と GitHub Repository 連携

# プロジェクト情報（番号だけ欲しい）
data "google_project" "this" {}

locals {
  # Cloud Build デフォルトSA
  cloudbuild_sa_name = "projects/-/serviceAccounts/${data.google_project.this.number}@cloudbuild.gserviceaccount.com"
}

# GitHub ↔ Cloud Build Repository 接続
resource "google_cloudbuildv2_repository" "picca_repository" {
  provider          = google-beta
  project           = local.project_id   # ← main.tf の locals を参照
  location          = local.region
  name              = "picca-repo"

  parent_connection = "projects/${local.project_id}/locations/${local.region}/connections/github-connection"
  remote_uri        = "https://github.com/Udonburo/picca.git"
}

# メインブランチ用 Cloud Build Trigger
resource "google_cloudbuild_trigger" "main-branch-trigger" {
  provider        = google-beta
  project         = local.project_id
  location        = local.region
  name            = "main-branch-trigger"


  repository_event_config {
    repository = google_cloudbuildv2_repository.picca_repository.id
    push {
      branch = "^main$"
    }
  }

  filename = "infra/cloudbuild.yaml"
}
