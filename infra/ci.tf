data "google_project" "this" {}

locals {
  project_number = data.google_project.this.number
  cicd_sa_name   = "projects/${local.project_number}/serviceAccounts/terraform-sa@${local.project_id}.iam.gserviceaccount.com"
}

# GitHub ↔ Cloud Build Repository
resource "google_cloudbuildv2_repository" "picca_repository" {
  provider = google-beta
  project  = local.project_id
  location = local.region
  name     = "picca-repo"

  parent_connection = "projects/${local.project_id}/locations/${local.region}/connections/github-connection"
  remote_uri        = "https://github.com/Udonburo/picca.git"
}

# Trigger
resource "google_cloudbuild_trigger" "main-branch-trigger" {
  provider = google-beta
  project  = local.project_id
  location = local.region
  name     = "main-branch-trigger"


  service_account = "projects/${local.project_id}/serviceAccounts/terraform-sa@${local.project_id}.iam.gserviceaccount.com"

  repository_event_config {
    repository = google_cloudbuildv2_repository.picca_repository.id
    push {
      branch = "^main$"
    }
  }

  filename = "infra/cloudbuild.yaml"
}

# Trigger for infrastructure changes via iac/ branches
resource "google_cloudbuild_trigger" "iac-branch-trigger" {
  provider = google-beta
  project  = local.project_id
  location = local.region
  name     = "iac-branch-trigger"

  service_account = "projects/${local.project_id}/serviceAccounts/terraform-sa@${local.project_id}.iam.gserviceaccount.com"

  repository_event_config {
    repository = google_cloudbuildv2_repository.picca_repository.id
    push {
      branch = "^iac/.*"
    }
  }

  filename = "infra/cloudbuild-iac.yaml"
}

# Trigger for infrastructure tags like infra-*
resource "google_cloudbuild_trigger" "iac-tag-trigger" {
  provider = google-beta
  project  = local.project_id
  location = local.region
  name     = "iac-tag-trigger"

  service_account = "projects/${local.project_id}/serviceAccounts/terraform-sa@${local.project_id}.iam.gserviceaccount.com"

  repository_event_config {
    repository = google_cloudbuildv2_repository.picca_repository.id
    push {
      tag = "^infra-.*"
    }
  }

  filename = "infra/cloudbuild-iac.yaml"
}
