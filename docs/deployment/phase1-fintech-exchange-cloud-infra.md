# Phase 1 Fintech Exchange Deployment Architecture

## Purpose

This document defines the first deployable Phase 1 topology for a fintech or exchange onboarding pilot.
It is intentionally small and must not become the long-term global architecture by accident.

## Target topology

Services:

- `issuer-api`
- `verifier-api`
- `trust-registry`
- Ory Hydra

State:

- one PostgreSQL-compatible SQL instance
- separate logical databases or schemas for HDIP Phase 1 state and Hydra state
- separate least-privileged database users per service where supported

Edge:

- TLS termination at the cloud edge or load balancer
- WAF and rate limiting for public issuer/verifier surfaces
- private network access for `trust-registry`
- private or protected access for issuer console and operational endpoints

## Why this topology

The first product proof is:

> one real user is verified once and onboarded into a second platform without repeating KYC

That does not require Kubernetes, multi-region clusters, event streaming, mobile wallets, proof systems, or AI scoring.
It does require reliable SQL state, explicit service readiness, and clear separation between public partner APIs and internal trust reads.

## Component responsibilities

### issuer-api

Public or private depending on pilot model.
For the first internal issuer console pilot, keep it behind private access or partner-restricted ingress.

Owns:

- credential creation
- credential reads for issuer operators
- status mutation
- issuer-side audit writes

### verifier-api

Public partner-facing API for fintech/exchange backend calls.

Owns:

- verification request intake
- deterministic decision response
- verifier result retrieval
- verifier-side audit writes
- runtime trust reads from `trust-registry`

### trust-registry

Internal service only in the first deployment.

Owns:

- issuer trust state
- allowed templates
- verification-key references
- trust bootstrap and trust update state
- Hydra introspection validation for internal runtime trust reads

### Ory Hydra

Hydra is used for the governed internal verifier-to-trust-registry client-credentials flow from ADR 0010.

It is not yet the public issuer/verifier operator auth rollout.

### SQL

SQL is the primary Phase 1 runtime state.

Required lifecycle:

1. apply schema:

```bash
phase1sql migrate up --dsn "$PHASE1_DATABASE_URL"
```

2. bootstrap trust:

```bash
phase1sql bootstrap trust --dsn "$PHASE1_DATABASE_URL" --file ./trust-bootstrap.json
```

3. start services only after migration and trust bootstrap complete

## Network shape

Public ingress:

- `verifier-api`
- optionally `issuer-api` for a controlled issuer operator pilot
- issuer console through protected access only
- developer portal/docs when published

Private ingress:

- `trust-registry`
- Hydra admin and introspection paths
- SQL

Service egress:

- `verifier-api` to Hydra token endpoint
- `verifier-api` to `trust-registry`
- `trust-registry` to Hydra introspection
- all three services to SQL

## Deployment environments

### Local

Use local process execution, the sandbox runner, or the Phase 1 Docker Compose stack.
Local execution may remain developer-oriented, but it must not reintroduce runtime JSON fallback as a service mode.

The local Compose stack lives at `infra/phase1/docker-compose.yml`.
It is packaging only:

- `phase1sql migrate up` is a separate one-shot job
- `phase1sql bootstrap trust` is a separate one-shot job
- service containers do not mutate schema during startup
- `trust-registry` remains private to the Compose network
- Compose uses `infra/phase1/Dockerfile.services` to build separate runtime images from one shared Go build stage

### Pilot

Minimum viable cloud deployment:

- one managed PostgreSQL-compatible SQL instance
- one Hydra instance
- three service processes or containers
- one edge/load-balancer layer
- one protected issuer console deployment
- one developer portal or static docs surface

### Later production

Deferred:

- Kubernetes
- multi-region SQL topology
- separate audit store
- managed secret rotation automation beyond first-pilot controls
- full public OAuth/OIDC onboarding for issuer and verifier operators

## Readiness gates

Each deployment must verify:

- SQL schema is migrated
- trust bootstrap exists
- Hydra token acquisition works for `verifier-api`
- Hydra introspection works for `trust-registry`
- `issuer-api /readyz` is healthy
- `verifier-api /readyz` is healthy
- `trust-registry /readyz` is healthy
- suspended issuer path returns verifier `deny`
- revoked credential path returns verifier `deny`

## Local packaging command

```bash
docker compose --env-file infra/phase1/.env.example -f infra/phase1/docker-compose.yml up --build
```

Use `docs/integration/quickstart.md` for the external integrator walkthrough and `docs/runbooks/phase1-sandbox.md` for the automated lifecycle check.

## Secrets

Required secret classes:

- SQL credentials
- Hydra system secrets and client credentials
- service-to-service client secret for verifier runtime trust reads
- edge or deployment credentials

Rules:

- no secrets in repository files
- no secrets in screenshots, fixtures, examples, or logs
- use environment injection from the deployment secret manager
- rotate before external pilot access

## Logging and observability

Minimum pilot telemetry:

- structured service logs
- request IDs
- readiness failures
- verification decision counts by decision and reason code
- audit append success/failure counts
- SQL connectivity failures
- Hydra token/introspection failures

Do not log:

- opaque credential artifact values
- normalized KYC claims
- raw request bodies
- bearer tokens
- Hydra client secrets

## Rollback

Rollback must preserve data integrity.

Allowed first-pilot rollback:

- stop public ingress
- keep SQL instance intact
- roll service containers/processes back together
- do not downgrade schema unless an explicit reverse migration exists
- keep trust bootstrap file and applied trust state auditable

## Production documentation references

These external references are operational inputs, not repo rule precedence:

- Ory Hydra overview: https://www.ory.com/docs/network/hydra
- Ory OAuth2 token introspection: https://www.ory.com/docs/hydra/guides/oauth2-token-introspection
- PostgreSQL continuous archiving and point-in-time recovery: https://www.postgresql.org/docs/current/continuous-archiving.html
- Cloudflare WAF: https://developers.cloudflare.com/waf/
- Cloudflare API Shield: https://developers.cloudflare.com/api-shield/
