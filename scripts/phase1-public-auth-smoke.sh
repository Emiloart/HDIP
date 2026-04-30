#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
root="$(cd "$root" && pwd -P)"
cd "$root"

source scripts/toolchain-env.sh
root="$(hdip_resolve_repo_root "$root")"
cd "$root"

ISSUER_API_BASE_URL="${ISSUER_API_BASE_URL:-http://127.0.0.1:18081}"
VERIFIER_API_BASE_URL="${VERIFIER_API_BASE_URL:-http://127.0.0.1:18082}"
HYDRA_PUBLIC_URL="${HYDRA_PUBLIC_URL:-http://127.0.0.1:4444}"
ISSUER_CLIENT_ID="${ISSUER_CLIENT_ID:-did:web:issuer.hdip.dev}"
ISSUER_CLIENT_SECRET="${ISSUER_CLIENT_SECRET:-issuer-public-client-secret}"
VERIFIER_CLIENT_ID="${VERIFIER_CLIENT_ID:-verifier_org_sandbox}"
VERIFIER_CLIENT_SECRET="${VERIFIER_CLIENT_SECRET:-verifier-public-client-secret}"
ISSUER_SCOPE="issuer.credentials.issue issuer.credentials.read issuer.credentials.status.write"
VERIFIER_SCOPE="verifier.requests.create verifier.results.read"
TEMPLATE_ID="${TEMPLATE_ID:-hdip-passport-basic}"
POLICY_ID="${POLICY_ID:-kyc-passport-basic}"

workdir=""

fail() {
  echo "FAIL: $*" >&2
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

cleanup() {
  if [[ -n "$workdir" && -d "$workdir" ]]; then
    rm -rf "$workdir"
  fi
}
trap cleanup EXIT

request_token() {
  local client_id="$1"
  local client_secret="$2"
  local scope="$3"
  local response_file="$4"

  local status
  status="$(
    curl -sS \
      -o "$response_file" \
      -w "%{http_code}" \
      -X POST "${HYDRA_PUBLIC_URL}/oauth2/token" \
      -H "Content-Type: application/x-www-form-urlencoded" \
      -d "grant_type=client_credentials" \
      --data-urlencode "client_id=${client_id}" \
      --data-urlencode "client_secret=${client_secret}" \
      --data-urlencode "scope=${scope}"
  )"

  if [[ "$status" != "200" ]]; then
    echo "Hydra token response:" >&2
    sed -n '1,120p' "$response_file" >&2 || true
    fail "token request failed with HTTP $status"
  fi

  python3 - "$response_file" <<'PY'
import json
import sys
with open(sys.argv[1], "r", encoding="utf-8") as handle:
    print(json.load(handle)["access_token"])
PY
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
    sed -n '1,160p' "$output_file" >&2 || true
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

write_issue_request() {
  local target_file="$1"
  cat >"$target_file" <<JSON
{
  "templateId": "${TEMPLATE_ID}",
  "subjectReference": "public-auth-smoke-subject",
  "claims": {
    "fullLegalName": "Public Auth Smoke User",
    "dateOfBirth": "1990-01-02",
    "countryOfResidence": "NG",
    "documentCountry": "NG",
    "kycLevel": "basic",
    "verifiedAt": "2026-04-28T10:00:00Z",
    "expiresAt": "2099-01-01T00:00:00Z"
  }
}
JSON
}

write_verification_request() {
  local credential_id="$1"
  local artifact_file="$2"
  local target_file="$3"

  python3 - "$POLICY_ID" "$credential_id" "$artifact_file" "$target_file" <<'PY'
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

main() {
  require_command curl
  require_command python3

  HYDRA_PUBLIC_URL="$(normalize_url "$HYDRA_PUBLIC_URL")"
  ISSUER_API_BASE_URL="$(normalize_url "$ISSUER_API_BASE_URL")"
  VERIFIER_API_BASE_URL="$(normalize_url "$VERIFIER_API_BASE_URL")"
  workdir="$(mktemp -d)"
  run_id="phase1-public-auth-smoke-$(date +%s%N)"

  echo "HDIP Phase 1 public-auth smoke"

  issuer_token="$(request_token "$ISSUER_CLIENT_ID" "$ISSUER_CLIENT_SECRET" "$ISSUER_SCOPE" "$workdir/issuer-token.json")"
  verifier_token="$(request_token "$VERIFIER_CLIENT_ID" "$VERIFIER_CLIENT_SECRET" "$VERIFIER_SCOPE" "$workdir/verifier-token.json")"
  echo "token acquisition: ok"

  write_issue_request "$workdir/issue-request.json"

  curl_json "POST" "${ISSUER_API_BASE_URL}/v1/issuer/credentials" "$workdir/issue-request.json" "$workdir/verifier-cannot-issue.json" "403" \
    -H "Authorization: Bearer ${verifier_token}" \
    -H "Idempotency-Key: ${run_id}-verifier-cannot-issue"
  echo "verifier token cannot issue: ok"

  curl_json "GET" "${VERIFIER_API_BASE_URL}/v1/verifier/verifications/missing" "" "$workdir/invalid-token.json" "401" \
    -H "Authorization: Bearer invalid-token"
  echo "invalid token fails closed: ok"

  curl_json "POST" "${ISSUER_API_BASE_URL}/v1/issuer/credentials" "$workdir/issue-request.json" "$workdir/issue-response.json" "201" \
    -H "Authorization: Bearer ${issuer_token}" \
    -H "Idempotency-Key: ${run_id}-issue"

  credential_id="$(json_value "$workdir/issue-response.json" "credentialId")"
  json_value "$workdir/issue-response.json" "credentialArtifact" >"$workdir/credential-artifact.json"
  echo "issued credentialId: $credential_id"

  write_verification_request "$credential_id" "$workdir/credential-artifact.json" "$workdir/verify-request.json"

  curl_json "POST" "${VERIFIER_API_BASE_URL}/v1/verifier/verifications" "$workdir/verify-request.json" "$workdir/verify-allow-response.json" "201" \
    -H "Authorization: Bearer ${verifier_token}" \
    -H "Idempotency-Key: ${run_id}-verify-allow"
  first_decision="$(json_value "$workdir/verify-allow-response.json" "decision")"
  [[ "$first_decision" == "allow" ]] || fail "expected allow, got $first_decision"
  echo "first verification result: $first_decision"

  printf '{"status":"revoked"}\n' >"$workdir/revoke-request.json"
  curl_json "POST" "${ISSUER_API_BASE_URL}/v1/issuer/credentials/${credential_id}/status" "$workdir/revoke-request.json" "$workdir/revoke-response.json" "200" \
    -H "Authorization: Bearer ${issuer_token}" \
    -H "Idempotency-Key: ${run_id}-revoke"

  curl_json "POST" "${VERIFIER_API_BASE_URL}/v1/verifier/verifications" "$workdir/verify-request.json" "$workdir/verify-deny-response.json" "201" \
    -H "Authorization: Bearer ${verifier_token}" \
    -H "Idempotency-Key: ${run_id}-verify-deny"
  second_decision="$(json_value "$workdir/verify-deny-response.json" "decision")"
  [[ "$second_decision" == "deny" ]] || fail "expected deny, got $second_decision"
  echo "second verification result: $second_decision"

  echo "final status: PASS"
}

main "$@"
