locals {
  cloudbuild_sa = "${data.google_project.this.number}@cloudbuild.gserviceaccount.com"
}

# Allow Cloud Build to deploy Cloud Run
resource "google_project_iam_member" "cb_run_admin" {
  project = data.google_project.this.project_id
  role    = "roles/run.admin"
  member  = "serviceAccount:${local.cloudbuild_sa}"
}

# Allow Cloud Build to invoke the service during smoke test
resource "google_project_iam_member" "cb_run_invoker" {
  project = data.google_project.this.project_id
  role    = "roles/run.invoker"
  member  = "serviceAccount:${local.cloudbuild_sa}"
}

# Allow Cloud Build to push images to Artifact Registry
resource "google_artifact_registry_repository_iam_member" "cb_repo_writer" {
  project    = data.google_project.this.project_id
  location   = "asia-northeast1"
  repository = "picca-backend"
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:${local.cloudbuild_sa}"
}

# If a dedicated runtime SA is used, let Cloud Build target it
variable "runtime_sa_email" {
  type        = string
  description = "Service account email used by picca-ml-py-prod (optional)."
  default     = ""
}

resource "google_service_account_iam_member" "cb_sa_user" {
  count              = var.runtime_sa_email == "" ? 0 : 1
  service_account_id = "projects/${data.google_project.this.project_id}/serviceAccounts/${var.runtime_sa_email}"
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${local.cloudbuild_sa}"
}