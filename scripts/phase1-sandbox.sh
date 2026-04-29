#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
root="$(cd "$root" && pwd -P)"
cd "$root"

source scripts/toolchain-env.sh
root="$(hdip_resolve_repo_root "$root")"
cd "$root"

DATABASE_DRIVER="${DATABASE_DRIVER:-pgx}"
ISSUER_API_HOST="${ISSUER_API_HOST:-127.0.0.1}"
ISSUER_API_PORT="${ISSUER_API_PORT:-18081}"
VERIFIER_API_HOST="${VERIFIER_API_HOST:-127.0.0.1}"
VERIFIER_API_PORT="${VERIFIER_API_PORT:-18082}"
TRUST_REGISTRY_HOST="${TRUST_REGISTRY_HOST:-127.0.0.1}"
TRUST_REGISTRY_PORT="${TRUST_REGISTRY_PORT:-18083}"
ISSUER_ID="${ISSUER_ID:-did:web:issuer.hdip.dev}"
TEMPLATE_ID="${TEMPLATE_ID:-hdip-passport-basic}"
POLICY_ID="${POLICY_ID:-kyc-passport-basic}"
VERIFIER_TRUST_CLIENT_ID="${VERIFIER_TRUST_CLIENT_ID:-verifier-api}"
TRUST_REGISTRY_INTROSPECTION_CLIENT_ID="${TRUST_REGISTRY_INTROSPECTION_CLIENT_ID:-trust-registry}"
TRUST_RUNTIME_SCOPE="${TRUST_RUNTIME_SCOPE:-trust.runtime.read}"
RUN_SUSPEND_CHECK="${HDIP_PHASE1_SANDBOX_RUN_SUSPEND_CHECK:-1}"
DRY_RUN="${HDIP_PHASE1_SANDBOX_DRY_RUN:-0}"

ISSUER_API_BASE_URL="http://${ISSUER_API_HOST}:${ISSUER_API_PORT}"
VERIFIER_API_BASE_URL="http://${VERIFIER_API_HOST}:${VERIFIER_API_PORT}"
TRUST_REGISTRY_BASE_URL="http://${TRUST_REGISTRY_HOST}:${TRUST_REGISTRY_PORT}"

service_pids=()
workdir=""
credential_id=""
first_decision=""
second_decision=""
suspended_decision=""
restore_trust_on_exit=0

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

require_env() {
  local name="$1"
  if [[ -z "${!name:-}" ]]; then
    fail "$name is required"
  fi
}

require_command() {
  local name="$1"
  command -v "$name" >/dev/null 2>&1 || fail "$name is required"
}

normalize_url() {
  local value="$1"
  value="${value%/}"
  printf '%s' "$value"
}

cleanup() {
  local status=$?

  if [[ "$restore_trust_on_exit" == "1" && -n "${DATABASE_URL:-}" && -n "$workdir" && -f "$workdir/trust-active.json" ]]; then
    (
      cd services/internal/phase1sql
      go run ./cmd/phase1sql bootstrap trust \
        --driver "$DATABASE_DRIVER" \
        --dsn "$DATABASE_URL" \
        --file "$workdir/trust-active.json" >/dev/null 2>&1 || true
    )
  fi

  for pid in "${service_pids[@]}"; do
    if kill -0 "$pid" >/dev/null 2>&1; then
      kill "$pid" >/dev/null 2>&1 || true
      wait "$pid" >/dev/null 2>&1 || true
    fi
  done

  if [[ -n "$workdir" && -d "$workdir" ]]; then
    rm -rf "$workdir"
  fi

  exit "$status"
}
trap cleanup EXIT

preflight() {
  require_env DATABASE_URL
  require_env HYDRA_ADMIN_URL
  require_env HYDRA_PUBLIC_URL
  require_env VERIFIER_TRUST_CLIENT_SECRET
  require_env TRUST_REGISTRY_INTROSPECTION_CLIENT_SECRET

  require_command go
  require_command curl
  require_command python3

  HYDRA_ADMIN_URL="$(normalize_url "$HYDRA_ADMIN_URL")"
  HYDRA_PUBLIC_URL="$(normalize_url "$HYDRA_PUBLIC_URL")"
}

write_trust_bootstrap() {
  local trust_state="$1"
  local target_file="$2"

  python3 - "$ISSUER_ID" "$trust_state" "$TEMPLATE_ID" "$target_file" <<'PY'
import json
import sys

issuer_id, trust_state, template_id, target_file = sys.argv[1:5]
payload = {
    "issuers": [
        {
            "issuerId": issuer_id,
            "displayName": "HDIP Sandbox Issuer",
            "trustState": trust_state,
            "allowedTemplateIds": [template_id],
            "verificationKeyReferences": ["phase1-opaque-artifact-reference"],
        }
    ]
}
with open(target_file, "w", encoding="utf-8") as handle:
    json.dump(payload, handle, indent=2)
    handle.write("\n")
PY
}

phase1sql() {
  (
    cd services/internal/phase1sql
    go run ./cmd/phase1sql "$@"
  )
}

setup_sql() {
  write_trust_bootstrap "active" "$workdir/trust-active.json"

  echo "==> migrate Phase 1 SQL"
  phase1sql migrate up --driver "$DATABASE_DRIVER" --dsn "$DATABASE_URL"

  echo "==> bootstrap active issuer trust"
  phase1sql bootstrap trust \
    --driver "$DATABASE_DRIVER" \
    --dsn "$DATABASE_URL" \
    --file "$workdir/trust-active.json"
}

start_service() {
  local name="$1"
  local dir="$2"
  local log_file="$workdir/${name}.log"
  shift 2

  echo "==> start $name"
  (
    cd "$dir"
    env "$@" go run "./cmd/$name"
  ) >"$log_file" 2>&1 &

  service_pids+=("$!")
}

start_services() {
  start_service "trust-registry" "services/trust-registry" \
    HDIP_HOST="$TRUST_REGISTRY_HOST" \
    HDIP_PORT="$TRUST_REGISTRY_PORT" \
    HDIP_PHASE1_DATABASE_DRIVER="$DATABASE_DRIVER" \
    HDIP_PHASE1_DATABASE_URL="$DATABASE_URL" \
    HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_URL="${HYDRA_ADMIN_URL}/admin/oauth2/introspect" \
    HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_CLIENT_ID="$TRUST_REGISTRY_INTROSPECTION_CLIENT_ID" \
    HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_CLIENT_SECRET="$TRUST_REGISTRY_INTROSPECTION_CLIENT_SECRET" \
    HDIP_TRUST_RUNTIME_HYDRA_EXPECTED_CLIENT_ID="$VERIFIER_TRUST_CLIENT_ID" \
    HDIP_TRUST_RUNTIME_HYDRA_REQUIRED_SCOPE="$TRUST_RUNTIME_SCOPE"

  start_service "issuer-api" "services/issuer-api" \
    HDIP_HOST="$ISSUER_API_HOST" \
    HDIP_PORT="$ISSUER_API_PORT" \
    HDIP_PHASE1_DATABASE_DRIVER="$DATABASE_DRIVER" \
    HDIP_PHASE1_DATABASE_URL="$DATABASE_URL"

  start_service "verifier-api" "services/verifier-api" \
    HDIP_HOST="$VERIFIER_API_HOST" \
    HDIP_PORT="$VERIFIER_API_PORT" \
    HDIP_PHASE1_DATABASE_DRIVER="$DATABASE_DRIVER" \
    HDIP_PHASE1_DATABASE_URL="$DATABASE_URL" \
    HDIP_TRUST_REGISTRY_BASE_URL="$TRUST_REGISTRY_BASE_URL" \
    HDIP_TRUST_RUNTIME_HYDRA_TOKEN_URL="${HYDRA_PUBLIC_URL}/oauth2/token" \
    HDIP_TRUST_RUNTIME_HYDRA_CLIENT_ID="$VERIFIER_TRUST_CLIENT_ID" \
    HDIP_TRUST_RUNTIME_HYDRA_CLIENT_SECRET="$VERIFIER_TRUST_CLIENT_SECRET" \
    HDIP_TRUST_RUNTIME_HYDRA_SCOPE="$TRUST_RUNTIME_SCOPE"
}

wait_ready() {
  local name="$1"
  local url="$2"
  local log_file="$workdir/${name}.log"

  for _ in $(seq 1 60); do
    if curl -fsS "${url}/readyz" >/dev/null 2>&1; then
      echo "==> $name ready"
      return 0
    fi
    sleep 1
  done

  echo "==> $name log" >&2
  sed -n '1,160p' "$log_file" >&2 || true
  fail "$name did not become ready"
}

wait_for_services() {
  wait_ready "trust-registry" "$TRUST_REGISTRY_BASE_URL"
  wait_ready "issuer-api" "$ISSUER_API_BASE_URL"
  wait_ready "verifier-api" "$VERIFIER_API_BASE_URL"
}

curl_json() {
  local method="$1"
  local url="$2"
  local body_file="$3"
  local output_file="$4"
  local expected_status="$5"
  shift 5

  local status
  local curl_args=(
    -sS
    -o "$output_file"
    -w "%{http_code}"
    -X "$method"
    -H "Accept: application/json"
  )

  while [[ "$#" -gt 0 ]]; do
    curl_args+=("$1")
    shift
  done

  if [[ -n "$body_file" ]]; then
    curl_args+=(
      -H "Content-Type: application/json"
      --data-binary "@$body_file"
    )
  fi

  curl_args+=("$url")
  status="$(curl "${curl_args[@]}")"

  if [[ "$status" != "$expected_status" ]]; then
    echo "Unexpected HTTP status for $method $url: got $status expected $expected_status" >&2
    sed -n '1,200p' "$output_file" >&2 || true
    exit 1
  fi
}

json_value() {
  local file="$1"
  local expression="$2"
  python3 - "$file" "$expression" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as handle:
    value = json.load(handle)

for part in sys.argv[2].split("."):
    value = value[part]

if isinstance(value, (dict, list)):
    print(json.dumps(value, separators=(",", ":")))
else:
    print(value)
PY
}

issuer_headers=(
  -H "X-HDIP-Principal-ID: issuer_operator_sandbox"
  -H "X-HDIP-Organization-ID: ${ISSUER_ID}"
  -H "X-HDIP-Auth-Reference: phase1-sandbox"
  -H "X-HDIP-Scopes: issuer.credentials.issue, issuer.credentials.read, issuer.credentials.status.write"
)

verifier_headers=(
  -H "X-HDIP-Principal-ID: verifier_integrator_sandbox"
  -H "X-HDIP-Organization-ID: verifier_org_sandbox"
  -H "X-HDIP-Auth-Reference: phase1-sandbox"
  -H "X-HDIP-Scopes: verifier.requests.create, verifier.results.read"
)

create_credential() {
  local body="$workdir/issue-request.json"
  local response="$workdir/issue-response.json"

  cat >"$body" <<JSON
{
  "templateId": "${TEMPLATE_ID}",
  "subjectReference": "sandbox-subject-001",
  "claims": {
    "fullLegalName": "Sandbox User",
    "dateOfBirth": "1990-01-02",
    "countryOfResidence": "NG",
    "documentCountry": "NG",
    "kycLevel": "basic",
    "verifiedAt": "2026-04-28T10:00:00Z",
    "expiresAt": "2099-01-01T00:00:00Z"
  }
}
JSON

  curl_json "POST" "${ISSUER_API_BASE_URL}/v1/issuer/credentials" "$body" "$response" "201" \
    "${issuer_headers[@]}" \
    -H "Idempotency-Key: phase1-sandbox-issue-$(date +%s%N)"

  credential_id="$(json_value "$response" "credentialId")"
  json_value "$response" "credentialArtifact" >"$workdir/credential-artifact.json"
}

write_verification_request() {
  local target_file="$1"
  python3 - "$POLICY_ID" "$credential_id" "$workdir/credential-artifact.json" "$target_file" <<'PY'
import json
import sys

policy_id, credential_id, artifact_file, target_file = sys.argv[1:5]
with open(artifact_file, "r", encoding="utf-8") as handle:
    artifact = json.load(handle)

payload = {
    "policyId": policy_id,
    "credentialId": credential_id,
    "credentialArtifact": artifact,
}
with open(target_file, "w", encoding="utf-8") as handle:
    json.dump(payload, handle, separators=(",", ":"))
    handle.write("\n")
PY
}

verify_credential() {
  local label="$1"
  local expected_decision="$2"
  local body="$workdir/verify-${label}-request.json"
  local response="$workdir/verify-${label}-response.json"

  write_verification_request "$body"
  curl_json "POST" "${VERIFIER_API_BASE_URL}/v1/verifier/verifications" "$body" "$response" "201" \
    "${verifier_headers[@]}" \
    -H "Idempotency-Key: phase1-sandbox-verify-${label}-$(date +%s%N)"

  local decision
  decision="$(json_value "$response" "decision")"
  if [[ "$decision" != "$expected_decision" ]]; then
    echo "Verification response:" >&2
    sed -n '1,200p' "$response" >&2 || true
    fail "verification $label returned $decision, expected $expected_decision"
  fi

  printf '%s' "$decision"
}

revoke_credential() {
  local body="$workdir/revoke-request.json"
  local response="$workdir/revoke-response.json"

  cat >"$body" <<'JSON'
{"status":"revoked"}
JSON

  curl_json "POST" "${ISSUER_API_BASE_URL}/v1/issuer/credentials/${credential_id}/status" "$body" "$response" "200" \
    "${issuer_headers[@]}" \
    -H "Idempotency-Key: phase1-sandbox-revoke-$(date +%s%N)"
}

suspend_issuer() {
  write_trust_bootstrap "suspended" "$workdir/trust-suspended.json"
  phase1sql bootstrap trust \
    --driver "$DATABASE_DRIVER" \
    --dsn "$DATABASE_URL" \
    --file "$workdir/trust-suspended.json" >/dev/null
  restore_trust_on_exit=1
}

main() {
  echo "HDIP Phase 1 sandbox"
  preflight

  if [[ "$DRY_RUN" == "1" ]]; then
    echo "dry run: preflight passed"
    echo "final status: PASS"
    return
  fi

  workdir="$(mktemp -d)"

  setup_sql
  start_services
  wait_for_services

  echo "==> create credential"
  create_credential
  echo "issued credentialId: $credential_id"

  echo "==> verify active credential"
  first_decision="$(verify_credential "allow" "allow")"
  echo "first verification result: $first_decision"

  echo "==> revoke credential"
  revoke_credential

  echo "==> verify revoked credential"
  second_decision="$(verify_credential "revoked-deny" "deny")"
  echo "second verification result: $second_decision"

  if [[ "$RUN_SUSPEND_CHECK" == "1" ]]; then
    echo "==> suspend issuer and verify deny"
    suspend_issuer
    suspended_decision="$(verify_credential "suspended-deny" "deny")"
    echo "suspended issuer verification result: $suspended_decision"
  fi

  echo "final status: PASS"
}

main "$@"
