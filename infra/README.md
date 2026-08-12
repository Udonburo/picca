# Infrastructure

> **Archived:** these files document the infrastructure explored during the
> prototype. They are not connected to an active deployment pipeline and are
> not maintained as a ready-to-apply production stack.

This directory contains the Terraform and Cloud Build configuration used while
exploring the Picca deployment architecture.

## Model Storage Bucket

The `model_storage.tf` file provisions a Google Cloud Storage bucket named
`picca-models` to store machine learning models. The bucket is created in the
region defined by the `region` variable and uses the `NEARLINE` storage class.
A lifecycle rule moves objects to the `COLDLINE` storage class after 90 days.

Access to objects is granted to the service account
`ml-py-stg-sa@<project>.iam.gserviceaccount.com` with the
`roles/storage.objectViewer` role.

## Read-only validation

The configuration can be inspected without creating cloud resources:

```bash
cd infra
terraform init -backend=false
terraform validate
```

No automated GCP deployment workflow is retained in this repository.
