# Infrastructure

This directory contains Terraform configuration for the picca project.

## Model Storage Bucket

The `model_storage.tf` file provisions a Google Cloud Storage bucket named
`picca-models` to store machine learning models. The bucket is created in the
region defined by the `region` variable and uses the `NEARLINE` storage class.
A lifecycle rule moves objects to the `COLDLINE` storage class after 90 days.

Access to objects is granted to the service account
`ml-py-stg-sa@<project>.iam.gserviceaccount.com` with the
`roles/storage.objectViewer` role.

## Quick start

```bash
terraform init
terraform plan
terraform apply
terraform destroy
```
