# Phase 1 Pilot Readiness Runbook

## Purpose

This runbook is the preflight checklist before giving HDIP Phase 1 access to a controlled fintech/exchange pilot.

Phase 1 readiness means one approved partner can use Hydra client credentials to verify a reusable KYC credential through HDIP without repeating KYC.
It does not mean global production scale.

## Required Runtime Shape

- one PostgreSQL-compatible SQL instance
- one Ory Hydra instance
- `issuer-api`
- `verifier-api`
- private `trust-registry`
- protected issuer console
- public or partner-restricted verifier API ingress

## Required Environment Separation

Use `infra/phase1/.env.example` only for local Compose.
Use `infra/phase1/production.env.template` as the production/pilot variable checklist and inject real values from the deployment secret manager.

Production requirements:

- `HDIP_ENVIRONMENT=production`
- `HDIP_PUBLIC_AUTH_MODE=hydra`
- no local sample client secrets
- no trusted-header public access
- Hydra admin and introspection paths are private
- SQL is private
- `trust-registry` is private

## Required Release Order

1. apply SQL schema
2. bootstrap issuer trust
3. provision Hydra internal runtime clients
4. provision Hydra public issuer/verifier clients
5. start `trust-registry`
6. start `issuer-api`
7. start `verifier-api`
8. verify `/readyz` for all services
9. run public-auth smoke validation

## Edge Controls

Enforce these outside the issuer/verifier business logic:

- TLS at the edge
- WAF or equivalent request filtering
- request body size limits
- per-IP and per-client rate limits where the edge supports client identity
- private access for Hydra admin, Hydra introspection, SQL, and `trust-registry`
- no public access to debug, admin, database, or bootstrap surfaces

Do not implement Phase 1 rate limiting inside credential issuance or verification business logic.

## Minimum Observability

The current service baseline includes JSON logs, request IDs, access logs, readiness checks, and audit records.

Before pilot, confirm operators can inspect:

- request IDs across issuer/verifier/trust-registry logs
- readiness failures
- Hydra token and introspection failures
- SQL connectivity failures
- verification decisions by `allow` or `deny`
- audit append failures

Do not log:

- bearer tokens
- client secrets
- opaque credential artifacts
- normalized KYC claims
- raw request bodies

## Public-Auth Smoke

Against a running local Compose or pilot stack:

```bash
bash scripts/phase1-public-auth-smoke.sh
```

The script must print:

```text
token acquisition: ok
verifier token cannot issue: ok
invalid token fails closed: ok
first verification result: allow
second verification result: deny
final status: PASS
```

For non-local endpoints, set:

```bash
export HYDRA_PUBLIC_URL="<hydra-public-url>"
export ISSUER_API_BASE_URL="<issuer-api-url>"
export VERIFIER_API_BASE_URL="<verifier-api-url>"
export ISSUER_CLIENT_ID="<issuer-client-id>"
export ISSUER_CLIENT_SECRET="<issuer-client-secret>"
export VERIFIER_CLIENT_ID="<verifier-client-id>"
export VERIFIER_CLIENT_SECRET="<verifier-client-secret>"
```

## Rollback And Containment

Allowed first-pilot rollback:

- stop public ingress
- delete compromised Hydra clients
- keep SQL state intact
- roll all service images/processes back together
- do not use header mode in production
- do not downgrade schema without an explicit reverse migration

If credential misuse is suspected:

1. stop public ingress
2. delete the affected Hydra client
3. preserve SQL and audit logs
4. review audit records and request IDs
5. re-provision only after root cause is understood

## Completion Criteria

Pilot readiness is satisfied when:

- partner clients are provisioned through the runbook
- public-auth smoke passes
- private surfaces are not publicly reachable
- edge limits are configured
- rollback path is rehearsed
- partner integration docs are delivered
- no secrets or artifacts are present in logs, screenshots, docs, or repository files
