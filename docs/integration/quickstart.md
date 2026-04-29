# Phase 1 Integrator Quickstart

## Purpose

This quickstart lets a fintech or exchange engineer run the HDIP Phase 1 reusable-KYC loop locally:

1. issue a sandbox KYC credential
2. verify it through `verifier-api`
3. revoke it
4. verify that the same credential is denied

This is a sandbox path only.
It does not add wallet flows, selective disclosure, proof verification, or public production auth.

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

## Issue a credential

```bash
curl -sS http://127.0.0.1:18081/v1/issuer/credentials \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -H "Idempotency-Key: quickstart-issue-001" \
  -H "X-HDIP-Principal-ID: issuer_operator_quickstart" \
  -H "X-HDIP-Organization-ID: did:web:issuer.hdip.dev" \
  -H "X-HDIP-Auth-Reference: quickstart" \
  -H "X-HDIP-Scopes: issuer.credentials.issue, issuer.credentials.read, issuer.credentials.status.write" \
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
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -H "Idempotency-Key: quickstart-verify-allow-001" \
  -H "X-HDIP-Principal-ID: verifier_integrator_quickstart" \
  -H "X-HDIP-Organization-ID: verifier_org_quickstart" \
  -H "X-HDIP-Auth-Reference: quickstart" \
  -H "X-HDIP-Scopes: verifier.requests.create, verifier.results.read" \
  --data-binary @/tmp/hdip-verification-request.json
```

Expected `decision`:

```json
"allow"
```

## Revoke and expect deny

```bash
curl -sS "http://127.0.0.1:18081/v1/issuer/credentials/${credential_id}/status" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -H "Idempotency-Key: quickstart-revoke-001" \
  -H "X-HDIP-Principal-ID: issuer_operator_quickstart" \
  -H "X-HDIP-Organization-ID: did:web:issuer.hdip.dev" \
  -H "X-HDIP-Auth-Reference: quickstart" \
  -H "X-HDIP-Scopes: issuer.credentials.issue, issuer.credentials.read, issuer.credentials.status.write" \
  -d '{"status":"revoked"}'

curl -sS http://127.0.0.1:18082/v1/verifier/verifications \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -H "Idempotency-Key: quickstart-verify-deny-001" \
  -H "X-HDIP-Principal-ID: verifier_integrator_quickstart" \
  -H "X-HDIP-Organization-ID: verifier_org_quickstart" \
  -H "X-HDIP-Auth-Reference: quickstart" \
  -H "X-HDIP-Scopes: verifier.requests.create, verifier.results.read" \
  --data-binary @/tmp/hdip-verification-request.json
```

Expected `decision`:

```json
"deny"
```

## One-command lifecycle check

If you prefer an automated assertion path, use the sandbox runner with real SQL and Hydra:

```bash
export DATABASE_URL="postgres://hdip:hdip_sandbox_password@127.0.0.1:15432/hdip_phase1?sslmode=disable"
export HYDRA_ADMIN_URL="http://127.0.0.1:4445"
export HYDRA_PUBLIC_URL="http://127.0.0.1:4444"
export VERIFIER_TRUST_CLIENT_SECRET="verifier-runtime-secret"
export TRUST_REGISTRY_INTROSPECTION_CLIENT_SECRET="trust-registry-introspection-secret"

bash scripts/phase1-sandbox.sh
```

Expected final line:

```text
final status: PASS
```

## Security notes

- The `X-HDIP-*` headers are the current governed service-edge attribution boundary for sandbox flows, not production public auth.
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
