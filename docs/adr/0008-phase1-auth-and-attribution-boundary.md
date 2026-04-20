# 0008 Phase 1 Auth And Attribution Boundary

- Status: accepted
- Date: 2026-04-20
- Owners: repository maintainer

## Context

Phase 1 cannot safely define issuance or verification contracts without also defining who is allowed to call them, how caller identity is derived, and which actions must be attributable in audit records.

The current foundation state includes only auth placeholders in the web shells and no runtime caller identity model.
Leaving caller identity undecided would force rework across request schemas, storage records, and audit behavior in the first real code slice.

## Decision

### Issuer operator auth boundary

All Phase 1 issuer write actions are authenticated and attributable.

The Phase 1 issuer caller is an `issuer_operator` principal bound to exactly one issuer organization for the lifetime of the request.
The principal is conveyed to `issuer-api` through authenticated service-edge context rather than trusted request body fields.

The minimum authenticated issuer operator attributes are:

- principal identifier
- issuer organization identifier
- actor type
- granted scopes
- credential or session identifier used to authenticate

### Verifier integrator auth boundary

All Phase 1 verifier write actions are authenticated and attributable.

The Phase 1 verifier caller is a `verifier_integrator` principal bound to exactly one verifier organization for the lifetime of the request.
This is primarily a machine-to-machine caller model, though a later console flow may delegate through the same boundary.

The minimum authenticated verifier attributes are:

- principal identifier
- verifier organization identifier
- actor type
- granted scopes
- credential identifier used to authenticate

### Attribution model for issuance requests

Issuance requests derive authoritative caller identity from the authenticated context only.
`issuerId` in the request body is not authoritative and is not part of the Phase 1 issue request contract.

Every issuance write action must be attributable to:

- issuer organization
- authenticated principal
- request identifier
- idempotency key when present
- execution service
- timestamp

### Attribution model for verification requests

Verification requests derive authoritative caller identity from the authenticated context only.
`verifierId` in the request body is not authoritative and is not part of the Phase 1 verification submission contract.

Every verification write action must be attributable to:

- verifier organization
- authenticated principal
- request identifier
- idempotency key when present
- execution service
- timestamp

If a console-driven flow later acts through the verifier boundary, the delegated human actor may be captured as secondary audit metadata, but Phase 1 does not require that richer model.

### Actions requiring authenticated identity

The following actions require authenticated identity in Phase 1:

- `POST /v1/issuer/credentials`
- issuer-authenticated read of per-credential details
- issuer-authenticated revoke or supersede actions
- `POST /v1/verifier/verifications`
- verifier-authenticated read of per-verification results
- verifier-authenticated read of verifier-specific policy definitions if exposed

Read-only public metadata endpoints from the stub slice may remain unauthenticated until they are deliberately changed, but they are not sufficient for Phase 1 product behavior.

### Actions requiring auditable attribution

The following actions require auditable attribution in Phase 1:

- all authenticated issuer writes
- all authenticated verifier writes
- all status transitions
- all reads of credential details
- all reads of verification result details
- trust-registry writes affecting issuer trust state or active verification-key references

### Minimum authorization boundary

Phase 1 uses a minimum scope-based authorization model:

- issuer scopes:
  - `issuer.credentials.issue`
  - `issuer.credentials.read`
  - `issuer.credentials.status.write`
- verifier scopes:
  - `verifier.requests.create`
  - `verifier.results.read`

Caller identity and scopes are validated at the service edge before request handlers reach business logic.

### Phase 1 transport assumption for auth

Phase 1 assumes bearer-style authenticated service-edge context compatible with the accepted OIDC/OAuth platform direction from ADR 0002.
The next implementation slice may realize this behind narrow internal auth-context abstractions without forcing the full long-term auth platform rollout before issuance and verification code can start.

### Deferred richer auth and authz

Phase 1 explicitly defers:

- end-user holder authentication
- passkey UX and wallet-bound auth
- delegated organization administrators
- support impersonation
- fine-grained relationship-based authorization
- step-up auth
- policy authoring auth
- multi-issuer or multi-verifier org hierarchies

## Alternatives considered

### Anonymous issuance and verification with request-body organization fields

Rejected because it would destroy attribution and make Phase 1 replay and abuse handling indefensible.

### Full long-term authorization stack before first real Phase 1 flow

Rejected because it would block the first issuance and verification slice on a larger platform rollout than Phase 1 requires.

### API-key-only model for both issuer and verifier callers

Rejected because issuer write actions need a clearer human or operator attribution model than a shared static key.

## Security impact

Positive.
This decision makes privileged caller identity explicit, prevents organization identity from being accepted from request bodies, and requires auditable attribution for the Phase 1 write path.

## Privacy impact

Positive overall.
Attribution data is intentionally narrower than full end-user identity data, and the decision avoids adding broad caller metadata into request bodies or unrelated records.

## Migration / rollback

If the Phase 1 auth transport changes, preserve the same authenticated context shape and attribution requirements at the service edge.
Do not weaken caller binding or auditability for convenience during implementation.

Later migration to richer authz systems is allowed only if the minimum issuer and verifier attribution requirements remain satisfied or are strengthened.

## Consequences

- issuance and verification request contracts remain cleaner because authoritative organization identity is not supplied by callers in the body
- issuer and verifier write endpoints cannot be implemented as anonymous or half-authenticated stubs
- audit record design and persistence shape are now coupled correctly to caller identity
- the next implementation slice must add auth-context extraction before real issuance or verification handlers are considered done

## Open questions

- whether the initial issuer operator principal will always represent a human actor or may also represent a service operator in Phase 1
- whether verifier policy reads should be authenticated in the first implementation slice or remain internal-only until policy management exists
- whether Phase 1 should require idempotency keys on both issuance and verification writes or allow them to be optional but strongly recommended

## Related plans, PRs, and issues

- `docs/plans/active/0006-phase1-kyc-credential-and-verifier-api.md`
- `docs/adr/0006-phase1-credential-and-issuance-boundary.md`
- `docs/adr/0007-phase1-state-and-persistence-model.md`
- `docs/threat-model/full/0003-phase1-kyc-issuance-verification-and-auth.md`
