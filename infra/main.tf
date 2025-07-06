terraform {
  backend "gcs" {
    bucket = "terraform-state-picca-dev-464810" # さっき作ったバケット名
    prefix = "terraform/state"
  }

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

locals {
  project_id = "picca-dev-464810"
  region     = "asia-northeast1"
}

provider "google" {
  project = local.project_id
  region  = local.region
}

provider "google" {
  alias   = "beta"
  project = local.project_id
  region  = local.region
}

# --- Artifact Registry ---

resource "google_artifact_registry_repository" "backend" {
  provider      = google.beta
  location      = local.region
  repository_id = "picca-backend"
  description   = "Docker repository for picca backend"
  format        = "DOCKER"
}