#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$root"

source scripts/toolchain-env.sh

run_validation_steps() {
  rm -f \
    apps/issuer-console/tsconfig.tsbuildinfo \
    apps/verifier-console/tsconfig.tsbuildinfo \
    packages/ui/tsconfig.tsbuildinfo

  npm run lint --workspaces --if-present
  npm run typecheck --workspace @hdip/issuer-console
  npm run typecheck --workspace @hdip/verifier-console
  npm run typecheck --workspace @hdip/api-client
  npm run typecheck --workspace @hdip/ui
  npm run test --workspaces --if-present
  npm run schema:validate
}

windows_root=""
if command -v cygpath >/dev/null 2>&1; then
  windows_root="$(cygpath -w "$root")"
elif [[ "$root" =~ ^/mnt/([a-zA-Z])/(.*)$ ]]; then
  drive_letter="${BASH_REMATCH[1]^^}"
  windows_root="${drive_letter}:/${BASH_REMATCH[2]}"
fi

if [[ -n "$windows_root" ]] && command -v powershell.exe >/dev/null 2>&1; then
  powershell.exe -NoProfile -ExecutionPolicy Bypass -Command "
Set-Location '$windows_root'
Remove-Item 'apps/issuer-console/tsconfig.tsbuildinfo','apps/verifier-console/tsconfig.tsbuildinfo','packages/ui/tsconfig.tsbuildinfo' -ErrorAction SilentlyContinue
npm run lint --workspaces --if-present
if (\$LASTEXITCODE -ne 0) { exit \$LASTEXITCODE }
npm run typecheck --workspace @hdip/issuer-console
if (\$LASTEXITCODE -ne 0) { exit \$LASTEXITCODE }
npm run typecheck --workspace @hdip/verifier-console
if (\$LASTEXITCODE -ne 0) { exit \$LASTEXITCODE }
npm run typecheck --workspace @hdip/api-client
if (\$LASTEXITCODE -ne 0) { exit \$LASTEXITCODE }
npm run typecheck --workspace @hdip/ui
if (\$LASTEXITCODE -ne 0) { exit \$LASTEXITCODE }
npm run test --workspaces --if-present
if (\$LASTEXITCODE -ne 0) { exit \$LASTEXITCODE }
npm run schema:validate
exit \$LASTEXITCODE
"
else
  run_validation_steps
fi
