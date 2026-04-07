#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$root"

source scripts/toolchain-env.sh

bash scripts/check-governance.sh
bash scripts/check-no-secrets.sh
bash scripts/validate-rust.sh
bash scripts/validate-go.sh
bash scripts/validate-web.sh
