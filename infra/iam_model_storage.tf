resource "google_storage_bucket_iam_member" "model_viewer" {
  bucket = google_storage_bucket.model_storage.name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:ml-py-stg-sa@${var.project}.iam.gserviceaccount.com"
}
