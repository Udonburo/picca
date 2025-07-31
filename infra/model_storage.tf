resource "google_storage_bucket" "model_storage" {
  name                        = "picca-models"
  project                     = var.project
  location                    = var.region
  storage_class               = "NEARLINE"
  uniform_bucket_level_access = true

  lifecycle_rule {
    condition {
      age = 90
    }

    action {
      type          = "SetStorageClass"
      storage_class = "COLDLINE"
    }
  }
}


