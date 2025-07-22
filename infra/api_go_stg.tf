# infra/api_go_stg.tf
# Cloud Run (v2) – Go API Staging
resource "google_cloud_run_v2_service" "api_go_stg" {
  name     = "picca-api-go-stg"
  project  = local.project_id      # ← locals に合わせた
  location = local.region          # ← locals に合わせた

  template {
    containers {
      image = "gcr.io/${local.project_id}/picca-api-go-stg:initial"
      ports { container_port = 8080 }
      env   { name = "PORT" value = "8080" }
    }
  }

  traffic { percent = 100  latest_revision = true }
}

# 誰でも叩けるように Invoker 権限
resource "google_cloud_run_v2_service_iam_member" "api_go_stg_invoker" {
  name     = google_cloud_run_v2_service.api_go_stg.name
  project  = local.project_id
  location = local.region

  role   = "roles/run.invoker"
  member = "allUsers"
}

output "api_go_stg_url" {
  description = "URL of the Go API staging service"
  value       = google_cloud_run_v2_service.api_go_stg.uri
}
