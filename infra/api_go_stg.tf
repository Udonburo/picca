resource "google_cloud_run_v2_service" "api_go_stg" {
  name     = "picca-api-go-stg"
  project  = var.project
  location = var.region

  template {
    containers {
      image = "gcr.io/${var.project}/picca-api-go-stg:initial"
      ports {
        container_port = 8080
      }
    }
  }

  lifecycle {
    ignore_changes = [
      template[0].containers[0].image # ← リテラル参照
    ]
  }
}

resource "google_cloud_run_v2_service_iam_member" "api_go_stg_invoker" {
  name     = google_cloud_run_v2_service.api_go_stg.name
  project  = var.project
  location = var.region

  role   = "roles/run.invoker"
  member = "allUsers"
}

output "api_go_stg_url" {
  description = "URL of the Go API staging service"
  value       = google_cloud_run_v2_service.api_go_stg.uri
}
