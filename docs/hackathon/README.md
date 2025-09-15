## Repro (Core)
terraform -chdir=infra init -backend-config=backend.hcl
terraform -chdir=infra apply -var "project=$(gcloud config get-value project)" -var "region=asia-northeast1"

## Checks
git grep -nE 'picca-dev-464810|asia-northeast1' -- ':!docs/**' ':!**/backend.hcl.example'

