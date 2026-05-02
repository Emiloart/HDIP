# 0011 Phase 1 Fintech Exchange Deployment Topology

- Status: accepted
- Date: 2026-04-28
- Owners: repository maintainer

## Context

HDIP Phase 1 now has governed issuer, verifier, trust, SQL, and internal service-identity boundaries.
The next product step is a fintech/exchange onboarding pilot where one user can complete KYC once and a second platform can verify that KYC without repeating document collection.

The repo needs a first deployment topology before adding deployment assets or integration-specific UI.
This is a deployment topology decision, so it requires an ADR.

## Decision

The first Phase 1 fintech/exchange pilot deployment uses:

- one PostgreSQL-compatible SQL instance
- one Hydra instance
- `issuer-api`
- `verifier-api`
- `trust-registry`
- protected issuer console access
- partner-facing verifier API ingress
- private trust-registry and SQL access

The SQL instance may use separate logical databases, schemas, or users for:

- HDIP Phase 1 runtime state
- Hydra state

This topology does not supersede the long-term CockroachDB-compatible global architecture direction.
It is a first-pilot deployment shape that remains compatible with the accepted relational Phase 1 state model and `pgx` SQL path.

`phase1sql migrate up` and `phase1sql bootstrap trust` are required release steps before service startup.
Normal service startup must not be the primary schema migration or trust bootstrap mechanism.

Hydra remains the governed identity service for internal verifier-to-trust-registry runtime trust reads from ADR 0010.
Public issuer and verifier client-credentials auth is governed by ADR 0012.

## Alternatives considered

### Kubernetes-first deployment

Rejected for Phase 1 pilot because it adds operational surface before the first product validation loop requires it.

### Single monolith service

Rejected because the current repo already has real service boundaries for issuer, verifier, and trust-registry behavior.
Collapsing them would blur trust ownership.

### No Hydra in the first deployment

Rejected because ADR 0010 already governs Hydra client credentials plus introspection for internal runtime trust reads.

### Credential-ID-only verifier bridge

Rejected for this deployment design because accepted Phase 1 contracts require `credentialArtifact` as the verifier submission artifact.
A later resolver may make credential-ID-only or short-token UX possible.

## Security impact

Positive.
This keeps `trust-registry` internal, keeps SQL private, preserves Hydra-backed internal trust identity, and avoids adding public surfaces beyond the issuer/verifier product loop.

## Privacy impact

Positive with operational caveats.
The topology does not broaden data collection, but first-pilot deployment must keep SQL backups, logs, and support tooling from becoming secondary copies of KYC claims or opaque artifacts.

## Migration / rollback

If the first pilot topology fails operational review, rollback by removing deployment assets and keeping the existing local/runtime service architecture intact.

If later production requires Kubernetes or multi-region SQL, create a follow-up ADR instead of mutating this pilot topology silently.

Do not run a pilot where services start against unmigrated or unbootstrapped SQL.

## Consequences

- `docs/deployment/phase1-fintech-exchange-cloud-infra.md` becomes the deployment design source for the first pilot
- deployment assets should target a small process/container topology first
- public verifier integration docs must describe server-side API use only
- issuer console deployment must remain protected access, not public consumer UX
- short-token or QR resolver behavior remains deferred

## Open questions

- the first controlled pilot operations profile is documented in `docs/plans/active/0020-phase1-cloudflare-vault-vm-pilot-hardening.md`: Cloudflare edge, Vault KV v2, and a single Linux VM running Docker Compose
- which cloud provider hosts later production environments
- whether the first deployment uses managed PostgreSQL or a Cockroach-compatible managed SQL option
- whether issuer console access is protected by Zero Trust, VPN, or later delegated user auth

## Related plans, PRs, and issues

- `docs/plans/archive/0017-phase1-fintech-exchange-product-layer.md`
- `docs/product/phase1-issuer-console.md`
- `docs/integration/phase1-verifier-sdk-and-api-guide.md`
- `docs/deployment/phase1-fintech-exchange-cloud-infra.md`
- `docs/adr/0010-phase1-internal-trust-service-identity-and-sql-lifecycle.md`
- `docs/threat-model/delta/0006-phase1-product-integration-and-deployment.md`
