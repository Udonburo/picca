resource "google_service_account" "ml_py_stg_sa" {
  account_id   = "ml-py-stg-sa"
  display_name = "Picca ML-Py Cloud Run (stg)"
}
