#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$root"

bash scripts/check-governance.sh
bash scripts/check-no-secrets.sh
