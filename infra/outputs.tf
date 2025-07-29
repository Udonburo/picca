output "bucket_name" {
  description = "Name of the GCS bucket for ML models"
  value       = google_storage_bucket.model_storage.name
}

output "bucket_url" {
  description = "gs:// URI of the ML model bucket"
  value       = "gs://${google_storage_bucket.model_storage.name}"
}
