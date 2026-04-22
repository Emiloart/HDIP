# 0010 Phase 1 Internal Trust Service Identity And SQL Lifecycle

- Status: accepted
- Date: 2026-04-22
- Owners: repository maintainer

## Context

The current Phase 1 runtime trust-read slice is landed, but it is still using a narrow static bearer token between `verifier-api` and `trust-registry`.
The current shared SQL adapter also still relies on startup-time `ensureSchema()` as the operational initialization path.

That state was intentionally acceptable for the prior slice, but it is no longer precise enough for the next implementation step.
Accepted ADRs already lock:

- `trust-registry` ownership of issuer trust state in ADR 0007
- bearer-style Phase 1 auth compatibility with the OIDC/OAuth platform direction from ADR 0002 in ADR 0008
- CockroachDB-compatible relational persistence behind service-owned repository boundaries in ADR 0007 and ADR 0002

What is not yet locked is the concrete governed shape of:

- the internal service identity used for runtime trust reads
- the operational lifecycle used to initialize and bootstrap the primary SQL path

Without those decisions, the next implementation slice would need to improvise an auth model or migration model in a sensitive area.

## Decision

### Scope of this ADR

This ADR governs only:

- internal service-to-service identity for `verifier-api` runtime trust reads against `trust-registry`
- the primary SQL migration and bootstrap lifecycle for the currently approved Phase 1 relational state

It does not change:

- public Phase 1 issuer or verifier request and response contracts
- deterministic Phase 1 trust-policy outcomes
- holder, wallet, proof, or public auth behavior

### Internal trust runtime service identity

The governed internal service identity model for runtime trust reads is:

- Ory Hydra client-credentials access tokens
- used only for `verifier-api` to `trust-registry` runtime trust reads
- validated by `trust-registry` through Hydra token introspection

`verifier-api` is the confidential OAuth client for this boundary.
The authenticated service principal is the verifier service itself, not a human operator and not an end-user.

The required internal scope is:

- `trust.runtime.read`

The trust authority remains sourced from `trust-registry` runtime state.
`verifier-api` request bodies do not become authoritative for issuer trust or organization identity.

### Trust-registry validation requirements

For every authenticated runtime trust-read request, `trust-registry` must validate the presented Hydra access token by introspection.
The minimum acceptance checks are:

- token is active
- token represents the configured `verifier-api` confidential client identity
- token contains scope `trust.runtime.read`

If any of those checks fail, `trust-registry` must fail closed.

This ADR rejects the following shapes for this slice:

- static shared bearer tokens as the governed end state
- local JWT verification against Hydra JWKS
- forwarded trusted edge-auth context in place of resource-server validation
- mTLS-only identity without OAuth client identity

The currently landed static bearer-token mechanism is explicitly transitional and is superseded by this ADR for the next implementation slice.

### Phase 1 SQL lifecycle ownership

The shared `services/internal/phase1sql` layer is the schema owner for the current approved Phase 1 relational state.

It owns the versioned SQL lifecycle for:

- `phase1_sequences`
- `trust_registry_issuer_records`
- `issuer_api_credential_records`
- `verifier_api_verification_request_records`
- `verifier_api_verification_result_records`
- `phase1_audit_records`
- `phase1_idempotency_records`
- any indexes already required by that approved table set

No service may define a competing primary migration source for those Phase 1 tables.

### Primary SQL initialization shape

The primary SQL initialization model is an explicit shared CLI with versioned SQL assets.

The governed operational entrypoint is a `phase1sql` CLI under the shared SQL package boundary.
That CLI must expose, at minimum:

- `migrate up` to apply versioned schema migrations for the current approved Phase 1 relational state
- `bootstrap trust --file <path>` to apply trust-registry-owned issuer trust bootstrap data after schema migration

Although `bootstrap trust` is executed through the shared operational CLI, semantic ownership of the trust data remains with `trust-registry`.
The bootstrap document shape remains bounded to the existing deterministic Phase 1 trust record model:

- `issuerId`
- `displayName`
- `trustState`
- `allowedTemplateIds`
- `verificationKeyReferences`

Schema migration and trust bootstrap are separate explicit steps.

### Startup behavior after the follow-up implementation lands

Once the follow-up implementation slice lands:

- normal service startup must no longer be the primary migration mechanism
- the primary SQL path must fail closed when the required schema is absent or behind
- trust-registry-owned bootstrap must not rely on implicit startup mutation for the primary SQL path

The JSON-backed runtime may remain as a clearly transitional fallback, but it is not the primary lifecycle model governed by this ADR.

## Alternatives considered

### Keep static bearer tokens and document rotation only

Rejected because that would preserve the current transitional mechanism as architecture in a boundary that the repo already describes as requiring a stronger service identity model.

### Hydra access tokens with local JWT verification

Rejected for this slice because it would require locking additional token-shape, issuer, audience, and key-distribution details that the repo has not otherwise needed yet.
Hydra introspection is the narrower governed resource-server path.

### Forwarded edge-auth context instead of trust-registry token validation

Rejected because it would widen this slice into a broader edge-auth rollout and would weaken the explicit resource-server trust boundary.

### Keep startup-time `ensureSchema()` as the primary SQL lifecycle

Rejected because it leaves schema mutation implicit, weakens operational reproducibility, and does not satisfy ADR 0007's requirement for migration notes before business logic expands.

### Infra-owned SQL lifecycle with no shared service CLI

Rejected for this slice because the repo does not yet have a standardized infra migration framework, and forcing one here would widen scope beyond the current approved Phase 1 need.

## Security impact

Positive.
This ADR replaces an underspecified internal shared-secret pattern with a governed service identity model and removes ambiguity about who owns SQL schema mutation on the primary path.

## Privacy impact

Neutral to positive.
The decision does not broaden data collection and keeps trust bootstrap bounded to the already approved issuer trust record shape.

## Migration / rollback

Until the follow-up implementation lands, the current static bearer token and startup schema initialization remain the transitional behavior in code.

The next implementation slice must replace the static token path and explicit startup schema mutation together on the primary SQL path.
Do not keep a mixed steady state where:

- Hydra client-credentials are partly introduced but static bearer tokens remain the undeclared fallback
- primary SQL schema is partly migrated by CLI and partly mutated implicitly at service startup

If the implementation is rolled back, roll back the governed internal identity and explicit SQL lifecycle changes together.

## Consequences

- the next implementation slice must add Hydra client-credentials token acquisition in `verifier-api`
- the next implementation slice must add Hydra introspection-based token validation in `trust-registry`
- the next implementation slice must introduce versioned SQL assets and a shared `phase1sql` operational CLI
- public Phase 1 issuer and verifier HTTP contracts remain unchanged
- the current static bearer-token mechanism is no longer acceptable as the long-term Phase 1 internal trust auth model

## Open questions

- whether a later broader internal service-identity model should continue to use Hydra introspection for other internal resource-server boundaries or adopt a stronger shared pattern once that broader system is governed explicitly
- whether the eventual JSON fallback removal should occur immediately after the SQL CLI lifecycle stabilizes or in a later persistence hardening slice

## Related plans, PRs, and issues

- `docs/plans/active/0013-phase1-hydra-internal-trust-auth-and-phase1sql-lifecycle.md`
- `docs/plans/archive/0012-phase1-trust-registry-writes-bootstrap-and-internal-auth.md`
- `docs/adr/0007-phase1-state-and-persistence-model.md`
- `docs/adr/0008-phase1-auth-and-attribution-boundary.md`
- `docs/threat-model/full/0003-phase1-kyc-issuance-verification-and-auth.md`
