#!/usr/bin/env bash
set -euo pipefail

hdip_prepend_path_once() {
  local dir="${1:-}"
  if [[ -z "$dir" || ! -d "$dir" ]]; then
    return 0
  fi

  case ":$PATH:" in
    *":$dir:"*) ;;
    *) export PATH="$dir:$PATH" ;;
  esac
}

hdip_require_command() {
  local command_name="${1:-}"
  if [[ -z "$command_name" ]]; then
    echo "error: command name is required" >&2
    return 1
  fi

  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "error: required command not found on PATH: $command_name" >&2
    return 1
  fi
}

hdip_resolve_repo_root() {
  local input_root="${1:-}"
  if [[ -z "$input_root" ]]; then
    echo "error: repository root path is required" >&2
    return 1
  fi

  local root parent base candidate sibling entry
  root="$(cd "$input_root" && pwd -P)"
  parent="$(dirname "$root")"
  base="$(basename "$root")"
  candidate=""

  for sibling in "$parent"/*; do
    if [[ ! -d "$sibling" ]]; then
      continue
    fi

    entry="$(basename "$sibling")"
    if [[ "${entry,,}" == "${base,,}" ]]; then
      candidate="$sibling"
      break
    fi
  done

  if [[ -n "$candidate" ]]; then
    root="$candidate"
  fi

  printf '%s\n' "$root"
}

hdip_assert_posix_node_toolchain() {
  hdip_require_command node
  hdip_require_command npm

  local node_bin npm_bin
  node_bin="$(command -v node)"
  npm_bin="$(command -v npm)"

  if [[ "$node_bin" == *.exe || "$npm_bin" == *.exe || "$npm_bin" == *.cmd ]]; then
    echo "error: bash validation requires POSIX node and npm on PATH, got node=$node_bin npm=$npm_bin" >&2
    return 1
  fi
}

hdip_prepend_path_once "$HOME/.cargo/bin"
hdip_prepend_path_once "$HOME/sdk/go/bin"
hdip_prepend_path_once "$HOME/bin"

if [[ -s "$HOME/.nvm/nvm.sh" ]]; then
  export NVM_DIR="$HOME/.nvm"
  # shellcheck source=/dev/null
  source "$NVM_DIR/nvm.sh" >/dev/null 2>&1
  nvm use 22 >/dev/null 2>&1 || true
fi

if command -v node >/dev/null 2>&1; then
  hdip_prepend_path_once "$(dirname "$(command -v node)")"
fi

if command -v npm >/dev/null 2>&1; then
  hdip_prepend_path_once "$(dirname "$(command -v npm)")"
fi
