resource "google_cloud_run_v2_service" "api_go_stg" {
  name     = "picca-api-go-stg"
  project  = local.project_id
  location = local.region

  template {
    containers {
      image = "gcr.io/${local.project_id}/picca-api-go-stg:initial"

      ports {
        container_port = 8080
      }
      env {
        name  = "PORT"
        value = "8080"
      }
    }
  }
}

resource "google_cloud_run_v2_service_iam_member" "api_go_stg_invoker" {
  name     = google_cloud_run_v2_service.api_go_stg.name
  project  = local.project_id
  location = local.region

  role   = "roles/run.invoker"
  member = "allUsers"
}

output "api_go_stg_url" {
  value       = google_cloud_run_v2_service.api_go_stg.uri
  description = "URL of the Go API staging service"
}
