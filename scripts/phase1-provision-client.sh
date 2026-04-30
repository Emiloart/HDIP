#!/usr/bin/env bash
set -euo pipefail

issuer_scopes="issuer.credentials.issue issuer.credentials.read issuer.credentials.status.write"
verifier_scopes="verifier.requests.create verifier.results.read"

usage() {
  cat >&2 <<'EOF'
usage:
  HYDRA_ADMIN_URL=<url> bash scripts/phase1-provision-client.sh create issuer --client-id <issuer-org-id> [--client-secret <secret>] [--token-auth-method client_secret_post|client_secret_basic]
  HYDRA_ADMIN_URL=<url> bash scripts/phase1-provision-client.sh create verifier --client-id <verifier-org-id> [--client-secret <secret>] [--token-auth-method client_secret_post|client_secret_basic]
  HYDRA_ADMIN_URL=<url> bash scripts/phase1-provision-client.sh delete --client-id <client-id>

Outputs JSON to stdout. Generated client secrets are printed once and are not stored by this script.
EOF
}

fail() {
  echo "error: $*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || fail "$1 is required"
}

normalize_url() {
  local value="$1"
  value="${value%/}"
  printf '%s' "$value"
}

generate_secret() {
  python3 - <<'PY'
import secrets
print(secrets.token_urlsafe(48))
PY
}

json_string() {
  python3 - "$1" <<'PY'
import json
import sys
print(json.dumps(sys.argv[1]))
PY
}

print_created_json() {
  local client_type="$1"
  local client_id="$2"
  local client_secret="$3"
  local scopes="$4"
  local token_auth_method="$5"

  python3 - "$client_type" "$client_id" "$client_secret" "$scopes" "$token_auth_method" <<'PY'
import json
import sys

client_type, client_id, client_secret, scopes, token_auth_method = sys.argv[1:6]
print(json.dumps({
    "status": "created",
    "clientType": client_type,
    "clientId": client_id,
    "clientSecret": client_secret,
    "grantType": "client_credentials",
    "tokenEndpointAuthMethod": token_auth_method,
    "scopes": scopes.split(),
}, indent=2))
PY
}

print_deleted_json() {
  local client_id="$1"
  printf '{\n  "status": "deleted",\n  "clientId": %s\n}\n' "$(json_string "$client_id")"
}

validate_client_id() {
  local client_id="$1"
  [[ -n "$client_id" ]] || fail "--client-id is required"
  [[ "$client_id" != *$'\n'* && "$client_id" != *$'\r'* ]] || fail "--client-id must be a single line"
}

hydra_admin_url="${HYDRA_ADMIN_URL:-${HYDRA_ADMIN_ENDPOINT:-}}"
command_name="${1:-}"

if [[ -z "$command_name" || "$command_name" == "-h" || "$command_name" == "--help" ]]; then
  usage
  exit 0
fi

[[ -n "$hydra_admin_url" ]] || fail "HYDRA_ADMIN_URL is required"
hydra_admin_url="$(normalize_url "$hydra_admin_url")"

require_command hydra
require_command python3

case "$command_name" in
  create)
    [[ "$#" -ge 2 ]] || fail "create requires client type issuer or verifier"
    client_type="${2:-}"
    shift 2

    case "$client_type" in
      issuer)
        scopes="$issuer_scopes"
        ;;
      verifier)
        scopes="$verifier_scopes"
        ;;
      *)
        fail "create requires client type issuer or verifier"
        ;;
    esac

    client_id=""
    client_secret=""
    token_auth_method="client_secret_post"

    while [[ "$#" -gt 0 ]]; do
      case "$1" in
        --client-id)
          client_id="${2:-}"
          shift 2
          ;;
        --client-secret)
          client_secret="${2:-}"
          shift 2
          ;;
        --token-auth-method)
          token_auth_method="${2:-}"
          shift 2
          ;;
        *)
          fail "unknown create option: $1"
          ;;
      esac
    done

    validate_client_id "$client_id"
    case "$token_auth_method" in
      client_secret_post|client_secret_basic) ;;
      *) fail "--token-auth-method must be client_secret_post or client_secret_basic" ;;
    esac

    if [[ -z "$client_secret" ]]; then
      client_secret="$(generate_secret)"
    fi

    hydra create oauth2-client \
      --endpoint "$hydra_admin_url" \
      --id "$client_id" \
      --secret "$client_secret" \
      --grant-type client_credentials \
      --response-type token \
      --scope "$scopes" \
      --token-endpoint-auth-method "$token_auth_method" \
      --format json >/dev/null

    print_created_json "$client_type" "$client_id" "$client_secret" "$scopes" "$token_auth_method"
    ;;

  delete)
    shift
    client_id=""

    while [[ "$#" -gt 0 ]]; do
      case "$1" in
        --client-id)
          client_id="${2:-}"
          shift 2
          ;;
        *)
          fail "unknown delete option: $1"
          ;;
      esac
    done

    validate_client_id "$client_id"
    hydra delete oauth2-client "$client_id" --endpoint "$hydra_admin_url" >/dev/null
    print_deleted_json "$client_id"
    ;;

  *)
    fail "unknown command: $command_name"
    ;;
esac
