resource "google_cloud_run_v2_service" "ml_py_stg" {
  name     = "picca-ml-py-stg"
  project  = var.project
  location = var.region

  template {
    service_account = google_service_account.ml_py_stg_sa.email
    containers {
      image = "${var.region}-docker.pkg.dev/${var.project}/picca-backend/picca-ml-py-stg:latest"
      ports { container_port = 8080 }
    }
    scaling {
      min_instance_count = 0
      max_instance_count = 2
    }
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }

  lifecycle {
    ignore_changes = [template[0].containers[0].image]
  }
}

resource "google_cloud_run_v2_service_iam_member" "ml_py_stg_invoker" {
  name     = google_cloud_run_v2_service.ml_py_stg.name
  project  = var.project
  location = var.region

  role   = "roles/run.invoker"
  member = "allUsers"
}

output "ml_py_stg_url" {
  description = "URL of the ML Python staging service"
  value       = google_cloud_run_v2_service.ml_py_stg.uri
}
