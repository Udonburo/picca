terraform {
  backend "gcs" {}

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }

    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 5.0"
    }
  }

  required_version = ">= 1.5.0"
}

provider "google" {
  project = var.project
  region  = var.region
}

provider "google-beta" {
  project = var.project
  region  = var.region
}

resource "google_artifact_registry_repository" "backend" {
  location      = var.region
  repository_id = "picca-backend"
  description   = "Docker repository for picca backend"
  format        = "DOCKER"
}
