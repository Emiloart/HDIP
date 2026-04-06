#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$root"

pattern='BEGIN (RSA|EC|DSA|OPENSSH|PGP) PRIVATE KEY|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36}|github_pat_[A-Za-z0-9_]{20,}|sk-[A-Za-z0-9]{20,}'

if command -v rg >/dev/null 2>&1; then
  if rg -n --hidden --glob '!.git' --glob '!node_modules' --glob '!target' -e "$pattern" .; then
    echo "error: possible secret material detected"
    exit 1
  fi
else
  if grep -RInE --exclude-dir=.git --exclude-dir=node_modules --exclude-dir=target "$pattern" .; then
    echo "error: possible secret material detected"
    exit 1
  fi
fi

echo "Heuristic secret scan passed."
