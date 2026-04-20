#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$root"

source scripts/toolchain-env.sh

mapfile -t modules < <(find packages/go services -name go.mod -print | sort)

for module in "${modules[@]}"; do
  dir="$(dirname "$module")"
  mapfile -t go_files < <(find "$dir" -name '*.go' -print | sort)
  if [[ "${#go_files[@]}" -gt 0 ]]; then
    gofmt_output="$(gofmt -l "${go_files[@]}")"
    if [[ -n "$gofmt_output" ]]; then
      echo "error: gofmt drift detected"
      echo "$gofmt_output"
      exit 1
    fi
  fi
done

for module in "${modules[@]}"; do
  dir="$(dirname "$module")"
  (cd "$dir" && go test ./... && go vet ./...)
done
