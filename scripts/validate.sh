#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
root="$(cd "$root" && pwd -P)"
cd "$root"

source scripts/toolchain-env.sh
root="$(hdip_resolve_repo_root "$root")"
cd "$root"
hdip_assert_posix_node_toolchain

bash scripts/check-governance.sh
bash scripts/check-no-secrets.sh
bash scripts/validate-rust.sh
bash scripts/validate-go.sh
npm run schema:validate
bash scripts/validate-web.sh
