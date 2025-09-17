#!/usr/bin/env bash
set -euo pipefail
git grep -nE 'picca-dev-464810|asia-northeast1' -- ':!docs/' ':!/backend.hcl.example' || echo "OK: no hard-coded literals found"

