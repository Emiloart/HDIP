#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$root"

source scripts/toolchain-env.sh

cargo fmt --check
cargo clippy --all-targets --all-features -- -D warnings
cargo test
