#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
root="$(cd "$root" && pwd -P)"
cd "$root"

source scripts/toolchain-env.sh
root="$(hdip_resolve_repo_root "$root")"
cd "$root"

hdip_assert_posix_node_toolchain
hdip_require_command rsync
hdip_require_command sha256sum

stage="${1:-all}"
case "$stage" in
  all|lint|typecheck|test) ;;
  *)
    echo "error: unsupported validate-web stage: $stage" >&2
    exit 1
    ;;
esac

lint_workspaces=(
  "@hdip/issuer-console"
  "@hdip/verifier-console"
)

typecheck_workspaces=(
  "@hdip/api-client"
  "@hdip/ui"
  "@hdip/issuer-console"
  "@hdip/verifier-console"
)

test_workspaces=(
  "@hdip/api-client"
  "@hdip/ui"
  "@hdip/issuer-console"
  "@hdip/verifier-console"
)

cache_home="${HDIP_VALIDATION_CACHE_DIR:-${XDG_CACHE_HOME:-$HOME/.cache}/hdip}"
repo_key="$(printf '%s\n' "$root" | sha256sum | awk '{print substr($1, 1, 16)}')"
mirror_base="$cache_home/validate-web/$repo_key"
mirror_root="$mirror_base/root"
npm_cache_dir="$cache_home/npm"
install_fingerprint_file="$mirror_base/node_modules.fingerprint"

export npm_config_cache="$npm_cache_dir"
export npm_config_audit=false
export npm_config_fund=false
export npm_config_update_notifier=false

workspace_path_for() {
  local workspace="${1:?workspace is required}"

  case "$workspace" in
    "@hdip/api-client")
      printf '%s\n' "packages/api-client"
      ;;
    "@hdip/ui")
      printf '%s\n' "packages/ui"
      ;;
    "@hdip/issuer-console")
      printf '%s\n' "apps/issuer-console"
      ;;
    "@hdip/verifier-console")
      printf '%s\n' "apps/verifier-console"
      ;;
    *)
      echo "error: unsupported workspace: $workspace" >&2
      return 1
      ;;
  esac
}

compute_install_fingerprint() {
  (
    cd "$root"
    {
      node -v
      npm -v
      sha256sum package-lock.json package.json
      find apps packages -name package.json -print0 | sort -z | xargs -0 sha256sum
    } | sha256sum | awk '{print $1}'
  )
}

sync_mirror() {
  mkdir -p "$mirror_root" "$npm_cache_dir" "$mirror_base/eslint"

  echo "==> sync web validation mirror"
  rsync -a --delete \
    --exclude '.next/' \
    --exclude 'coverage/' \
    --exclude 'dist/' \
    --exclude 'node_modules/' \
    --exclude '*.tsbuildinfo' \
    "$root/package.json" \
    "$root/package-lock.json" \
    "$root/apps" \
    "$root/packages" \
    "$root/schemas" \
    "$mirror_root"/
}

ensure_mirror_dependencies() {
  local expected_fingerprint current_fingerprint
  expected_fingerprint="$(compute_install_fingerprint)"
  current_fingerprint=""

  if [[ -f "$install_fingerprint_file" ]]; then
    current_fingerprint="$(<"$install_fingerprint_file")"
  fi

  if [[ ! -d "$mirror_root/node_modules" || "$current_fingerprint" != "$expected_fingerprint" ]]; then
    echo "==> install web validation dependencies"
    rm -rf "$mirror_root/node_modules"
    (
      cd "$mirror_root"
      npm ci --prefer-offline --no-audit --no-fund >/dev/null
    )
    printf '%s\n' "$expected_fingerprint" > "$install_fingerprint_file"
  fi
}

clear_mirror_tsbuildinfo() {
  rm -f \
    "$mirror_root/apps/issuer-console/tsconfig.tsbuildinfo" \
    "$mirror_root/apps/verifier-console/tsconfig.tsbuildinfo" \
    "$mirror_root/packages/ui/tsconfig.tsbuildinfo"
}

run_workspace_script() {
  local script_name="${1:?script name is required}"
  shift

  local workspace
  for workspace in "$@"; do
    echo "==> $script_name $workspace"
    (
      cd "$mirror_root"
      npm run "$script_name" --workspace "$workspace" --if-present
    )
  done
}

run_workspace_lint() {
  local workspace workspace_path cache_name
  local -a targets

  for workspace in "${lint_workspaces[@]}"; do
    workspace_path="$(workspace_path_for "$workspace")"
    cache_name="${workspace#@hdip/}"

    case "$workspace" in
      "@hdip/issuer-console"| "@hdip/verifier-console")
        targets=(app lib next.config.ts eslint.config.mjs)
        ;;
      *)
        echo "error: unsupported lint workspace: $workspace" >&2
        return 1
        ;;
    esac

    echo "==> lint $workspace"
    (
      cd "$mirror_root/$workspace_path"
      "$mirror_root/node_modules/.bin/eslint" \
        --cache \
        --cache-location "$mirror_base/eslint/$cache_name.cache" \
        "${targets[@]}"
    )
  done
}

sync_mirror
ensure_mirror_dependencies
clear_mirror_tsbuildinfo

if [[ "$stage" == "all" || "$stage" == "lint" ]]; then
  run_workspace_lint
fi

if [[ "$stage" == "all" || "$stage" == "typecheck" ]]; then
  run_workspace_script typecheck "${typecheck_workspaces[@]}"
fi

if [[ "$stage" == "all" || "$stage" == "test" ]]; then
  run_workspace_script test "${test_workspaces[@]}"
fi
