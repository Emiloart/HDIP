# Phase 1 Sandbox Runbook

## Purpose

This runbook proves the narrow Phase 1 product loop without manual UI steps:

1. migrate SQL
2. bootstrap issuer trust
3. start `trust-registry`, `issuer-api`, and `verifier-api`
4. issue a reusable KYC credential
5. verify it and expect `allow`
6. revoke it
7. verify it again and expect `deny`
8. optionally suspend the issuer and expect `deny`

This is not a wallet flow, proof flow, selective-disclosure flow, or production signing flow.
The sandbox uses the accepted opaque Phase 1 `credentialArtifact`.

## Prerequisites

Run from a WSL-native workspace when possible.
The mounted Windows checkout may work for editing, but WSL-native execution is the lower-friction path for Go, Node, and validation.

Required local dependencies:

- PostgreSQL-compatible SQL database dedicated to sandbox use
- Ory Hydra with the verifier runtime client and trust-registry introspection client already configured
- Go toolchain
- `curl`
- `python3`

Required environment:

```bash
export DATABASE_URL="<postgresql-dsn-for-sandbox>"
export HYDRA_ADMIN_URL="http://127.0.0.1:4445"
export HYDRA_PUBLIC_URL="http://127.0.0.1:4444"
export VERIFIER_TRUST_CLIENT_SECRET="<verifier-runtime-client-secret>"
export TRUST_REGISTRY_INTROSPECTION_CLIENT_SECRET="<trust-registry-introspection-client-secret>"
```

Optional environment:

```bash
export DATABASE_DRIVER="pgx"
export VERIFIER_TRUST_CLIENT_ID="verifier-api"
export TRUST_REGISTRY_INTROSPECTION_CLIENT_ID="trust-registry"
export TRUST_RUNTIME_SCOPE="trust.runtime.read"
export HDIP_PHASE1_SANDBOX_RUN_SUSPEND_CHECK="1"
```

Do not commit real DSNs, bearer tokens, client secrets, opaque artifacts, or KYC claims.

## Command

```bash
bash scripts/phase1-sandbox.sh
```

The script starts services in the background, waits for readiness, runs the lifecycle, prints structured results, and cleans up service processes.

## Expected Output

The exact credential and verification IDs vary.
The important outputs are:

```text
HDIP Phase 1 sandbox
==> migrate Phase 1 SQL
==> bootstrap active issuer trust
==> start trust-registry
==> start issuer-api
==> start verifier-api
==> trust-registry ready
==> issuer-api ready
==> verifier-api ready
==> create credential
issued credentialId: credential_hdip_passport_basic_001
==> verify active credential
first verification result: allow
==> revoke credential
==> verify revoked credential
second verification result: deny
==> suspend issuer and verify deny
suspended issuer verification result: deny
final status: PASS
```

## What The Script Proves

- `phase1sql migrate up` initializes the governed SQL schema.
- `phase1sql bootstrap trust` creates an active trusted issuer.
- `issuer-api` can issue a credential through the existing Phase 1 contract.
- `verifier-api` can verify the issued opaque artifact and return `allow`.
- `issuer-api` can revoke the credential through the existing status endpoint.
- `verifier-api` observes the revoked state and returns `deny`.
- issuer suspension returns `deny` through the existing trust decision rule.

## Non-Destructive Validation Mode

To run a dry preflight without touching SQL or starting services:

```bash
DATABASE_URL="dry-run" \
HYDRA_ADMIN_URL="http://127.0.0.1:4445" \
HYDRA_PUBLIC_URL="http://127.0.0.1:4444" \
VERIFIER_TRUST_CLIENT_SECRET="dry-run" \
TRUST_REGISTRY_INTROSPECTION_CLIENT_SECRET="dry-run" \
HDIP_PHASE1_SANDBOX_DRY_RUN=1 \
bash scripts/phase1-sandbox.sh
```

Repo validation can invoke this path by setting:

```bash
export HDIP_VALIDATE_PHASE1_SANDBOX=1
```

## Go E2E Test

The real HTTP E2E test is opt-in because it requires live SQL and Hydra:

```bash
HDIP_PHASE1_E2E=1 \
DATABASE_URL="$DATABASE_URL" \
HYDRA_ADMIN_URL="$HYDRA_ADMIN_URL" \
HYDRA_PUBLIC_URL="$HYDRA_PUBLIC_URL" \
VERIFIER_TRUST_CLIENT_SECRET="$VERIFIER_TRUST_CLIENT_SECRET" \
TRUST_REGISTRY_INTROSPECTION_CLIENT_SECRET="$TRUST_REGISTRY_INTROSPECTION_CLIENT_SECRET" \
go test ./... ./services/e2e
```

The process-run sandbox uses the deprecated local header attribution path for issuer/verifier public calls.
Use `docs/integration/quickstart.md` and `infra/phase1/docker-compose.yml` to exercise the packaged Hydra public-auth path.
Use `scripts/phase1-public-auth-smoke.sh` to assert the packaged Hydra public-auth path without manual curl steps.

The test starts real service processes and uses real HTTP clients.
It does not mock HDIP services.

## Troubleshooting

### Hydra not reachable

Symptoms:

- `trust-registry` never becomes ready
- `verifier-api` never becomes ready
- verifier calls fail closed with trust runtime errors

Checks:

- `HYDRA_ADMIN_URL` points to Hydra admin
- `HYDRA_PUBLIC_URL` points to Hydra public
- verifier client credentials exist in Hydra
- trust-registry introspection credentials are valid

### SQL not migrated

Symptoms:

- services fail startup
- `/readyz` never returns healthy
- logs mention missing Phase 1 schema

Fix:

```bash
(
  cd services/internal/phase1sql
  source ../../../scripts/toolchain-env.sh
  go run ./cmd/phase1sql migrate up --dsn "$DATABASE_URL"
)
```

### Trust bootstrap missing

Symptoms:

- issuer credential creation returns `issuer_not_trusted`
- verifier returns `deny` with issuer trust reason codes

Fix:

- run the full sandbox script again, or
- apply `phase1sql bootstrap trust` with an active record for `did:web:issuer.hdip.dev`

### Port already in use

The sandbox defaults to:

- `18081` for `issuer-api`
- `18082` for `verifier-api`
- `18083` for `trust-registry`

Override if needed:

```bash
export ISSUER_API_PORT=28081
export VERIFIER_API_PORT=28082
export TRUST_REGISTRY_PORT=28083
```

## Integrator Handling Rules

- Store verifier decisions and audit identifiers, not raw KYC evidence.
- Do not log opaque artifacts, normalized claims, tokens, or raw request bodies.
- Treat the opaque artifact as temporary Phase 1 bridge material only.
- Do not treat the artifact as a signed credential, proof, wallet credential, or selective-disclosure artifact.
