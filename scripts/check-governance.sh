#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$root"

missing=0

required_dirs=(
  ".codex"
  ".github"
  ".github/workflows"
  "docs"
  "docs/adr"
  "docs/threat-model"
  "docs/plans"
  "docs/plans/active"
  "docs/plans/archive"
  "scripts"
)

required_files=(
  "AGENTS.md"
  "README.md"
  "CONTRIBUTING.md"
  "SECURITY.md"
  "CODEOWNERS"
  "Makefile"
  ".codex/config.toml"
  ".codex/hooks.json"
  ".github/pull_request_template.md"
  ".github/workflows/ci.yml"
  ".github/workflows/docs-check.yml"
  ".github/workflows/security.yml"
  "scripts/check-governance.sh"
  "scripts/check-no-secrets.sh"
  "scripts/codex/session_start_context.sh"
  "docs/repo-structure.md"
  "docs/adr/README.md"
  "docs/threat-model/README.md"
  "docs/plans/README.md"
  "docs/standards/README.md"
  "docs/privacy/README.md"
)

if [[ -f .codex ]]; then
  echo "error: .codex must be a directory, not a file"
  missing=1
fi

for dir in "${required_dirs[@]}"; do
  if [[ ! -d "$dir" ]]; then
    echo "error: missing directory $dir"
    missing=1
  fi
done

for file in "${required_files[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "error: missing file $file"
    missing=1
  fi
done

if grep -RIn --exclude-dir=.git --exclude-dir=.next --exclude-dir=node_modules '\[SETUP_COMMAND\]\|\[FORMAT_COMMAND\]\|\[LINT_COMMAND\]\|\[TYPECHECK_COMMAND\]\|\[FULL_VALIDATE_COMMAND\]' . >/dev/null 2>&1; then
  echo "error: unresolved command placeholders remain in repository files"
  missing=1
fi

if [[ "$missing" -ne 0 ]]; then
  exit 1
fi

echo "Governance structure check passed."
