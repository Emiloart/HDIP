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

if [[ "${HDIP_VALIDATE_PHASE1_SANDBOX:-0}" == "1" ]]; then
  DATABASE_URL="dry-run" \
    HYDRA_ADMIN_URL="http://127.0.0.1:4445" \
    HYDRA_PUBLIC_URL="http://127.0.0.1:4444" \
    VERIFIER_TRUST_CLIENT_SECRET="dry-run" \
    TRUST_REGISTRY_INTROSPECTION_CLIENT_SECRET="dry-run" \
    HDIP_PHASE1_SANDBOX_DRY_RUN=1 \
    bash scripts/phase1-sandbox.sh
fi
