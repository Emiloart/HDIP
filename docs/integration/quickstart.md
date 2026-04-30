# Phase 1 Integrator Quickstart

## Purpose

This quickstart lets a fintech or exchange engineer run the HDIP Phase 1 reusable-KYC loop locally:

1. issue a sandbox KYC credential
2. verify it through `verifier-api`
3. revoke it
4. verify that the same credential is denied

This is a sandbox path only.
It uses Hydra OAuth2 client credentials for packaged local issuer/verifier API access.
It does not add wallet flows, selective disclosure, proof verification, or self-service production partner provisioning.

## Prerequisites

- Docker Engine with Compose v2
- Bash, `curl`, and `python3`
- Go toolchain if you want to run `scripts/phase1-sandbox.sh` directly

## Start the local stack

From the repo root:

```bash
docker compose --env-file infra/phase1/.env.example -f infra/phase1/docker-compose.yml up --build
```

The stack starts:

- PostgreSQL with separate `hdip_phase1` and `hydra` databases
- Ory Hydra
- explicit `phase1sql migrate up`
- explicit `phase1sql bootstrap trust`
- `issuer-api`
- `verifier-api`
- private `trust-registry`

Public local endpoints:

- issuer API: `http://127.0.0.1:18081`
- verifier API: `http://127.0.0.1:18082`
- Hydra public: `http://127.0.0.1:4444`
- Hydra admin: `http://127.0.0.1:4445` for local debugging only

## Get sandbox access tokens

The local Compose stack provisions one issuer public client and one verifier public client.
The issuer client ID is the issuer organization identifier, so it uses `client_secret_post` to support the sandbox DID value.

```bash
issuer_token="$(curl -fsS http://127.0.0.1:4444/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  --data-urlencode "client_id=did:web:issuer.hdip.dev" \
  --data-urlencode "client_secret=issuer-public-client-secret" \
  --data-urlencode "scope=issuer.credentials.issue issuer.credentials.read issuer.credentials.status.write" \
  | python3 -c 'import json, sys; print(json.load(sys.stdin)["access_token"])'
)"

verifier_token="$(curl -fsS http://127.0.0.1:4444/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  --data-urlencode "client_id=verifier_org_sandbox" \
  --data-urlencode "client_secret=verifier-public-client-secret" \
  --data-urlencode "scope=verifier.requests.create verifier.results.read" \
  | python3 -c 'import json, sys; print(json.load(sys.stdin)["access_token"])'
)"
```

## Issue a credential

```bash
curl -sS http://127.0.0.1:18081/v1/issuer/credentials \
  -H "Authorization: Bearer ${issuer_token}" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -H "Idempotency-Key: quickstart-issue-001" \
  -d '{
    "templateId": "hdip-passport-basic",
    "subjectReference": "quickstart-subject-001",
    "claims": {
      "fullLegalName": "Quickstart User",
      "dateOfBirth": "1990-01-02",
      "countryOfResidence": "NG",
      "documentCountry": "NG",
      "kycLevel": "basic",
      "verifiedAt": "2026-04-28T10:00:00Z",
      "expiresAt": "2099-01-01T00:00:00Z"
    }
  }' | tee /tmp/hdip-issuance-response.json
```

Extract the values used by the verifier request:

```bash
credential_id="$(python3 - <<'PY'
import json
with open("/tmp/hdip-issuance-response.json", "r", encoding="utf-8") as handle:
    print(json.load(handle)["credentialId"])
PY
)"

python3 - <<'PY' >/tmp/hdip-credential-artifact.json
import json
with open("/tmp/hdip-issuance-response.json", "r", encoding="utf-8") as handle:
    print(json.dumps(json.load(handle)["credentialArtifact"], separators=(",", ":")))
PY
```

## Verify and expect allow

```bash
python3 - "$credential_id" /tmp/hdip-credential-artifact.json <<'PY' >/tmp/hdip-verification-request.json
import json
import sys
credential_id, artifact_path = sys.argv[1:3]
with open(artifact_path, "r", encoding="utf-8") as handle:
    artifact = json.load(handle)
print(json.dumps({
    "policyId": "kyc-passport-basic",
    "credentialId": credential_id,
    "credentialArtifact": artifact,
}, separators=(",", ":")))
PY

curl -sS http://127.0.0.1:18082/v1/verifier/verifications \
  -H "Authorization: Bearer ${verifier_token}" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -H "Idempotency-Key: quickstart-verify-allow-001" \
  --data-binary @/tmp/hdip-verification-request.json
```

Expected `decision`:

```json
"allow"
```

## Revoke and expect deny

```bash
curl -sS "http://127.0.0.1:18081/v1/issuer/credentials/${credential_id}/status" \
  -H "Authorization: Bearer ${issuer_token}" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -H "Idempotency-Key: quickstart-revoke-001" \
  -d '{"status":"revoked"}'

curl -sS http://127.0.0.1:18082/v1/verifier/verifications \
  -H "Authorization: Bearer ${verifier_token}" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -H "Idempotency-Key: quickstart-verify-deny-001" \
  --data-binary @/tmp/hdip-verification-request.json
```

Expected `decision`:

```json
"deny"
```

## One-command public-auth check

To assert the packaged public-auth path against the running Compose stack:

```bash
bash scripts/phase1-public-auth-smoke.sh
```

Expected final line:

```text
final status: PASS
```

## One-command lifecycle check

If you prefer an automated process-run assertion path, use the sandbox runner with real SQL and Hydra:

```bash
export DATABASE_URL="postgres://hdip:hdip_sandbox_password@127.0.0.1:15432/hdip_phase1?sslmode=disable"
export HYDRA_ADMIN_URL="http://127.0.0.1:4445"
export HYDRA_PUBLIC_URL="http://127.0.0.1:4444"
export VERIFIER_TRUST_CLIENT_SECRET="verifier-runtime-secret"
export TRUST_REGISTRY_INTROSPECTION_CLIENT_SECRET="trust-registry-introspection-secret"

bash scripts/phase1-sandbox.sh
```

That runner starts local service processes in deprecated header-auth mode for automation compatibility.
The packaged Compose quickstart above is the public-auth path.

Expected final line:

```text
final status: PASS
```

## Security notes

- Packaged local issuer/verifier API calls use Hydra bearer tokens.
- The `X-HDIP-*` headers are retained only for local process-run sandbox automation and tests.
- Keep verifier calls server-side only.
- Do not log opaque artifacts, KYC claims, bearer tokens, or client secrets.
- Do not treat the Phase 1 artifact as a signed credential or proof.

## Stop the stack

```bash
docker compose --env-file infra/phase1/.env.example -f infra/phase1/docker-compose.yml down
```

To remove local data:

```bash
docker compose --env-file infra/phase1/.env.example -f infra/phase1/docker-compose.yml down -v
```
