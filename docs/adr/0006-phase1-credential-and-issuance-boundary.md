# 0006 Phase 1 Credential And Issuance Boundary

- Status: accepted
- Date: 2026-04-20
- Owners: repository maintainer

## Context

HDIP has a landed foundation slice, hardening fixes, schema/example parity, and a deterministic stub issuer/verifier flow.
That state is sufficient to start Phase 1 work, but it does not yet define what the first real reusable KYC credential is, what is actually issued, what a verifier submits for evaluation, or which current services own that logic.

ADR 0009 clarifies the deterministic Phase 1 artifact semantics and partially supersedes earlier ambiguous "signed credential artifact" wording in this record.

Phase 1 must stay narrow:

- one reusable KYC credential
- one real issuer API boundary
- one real verifier API boundary
- minimal status handling
- minimal trust-registry participation

It must not implicitly pull in holder wallet UX, OpenID transport, selective disclosure, or richer trust products.

## Decision

### Phase 1 reusable KYC credential boundary

Phase 1 defines one real issued credential based on the existing `hdip-passport-basic` template identity.
The credential remains a standards-aligned reusable KYC attestation and carries the two credential types already implied by the landed stub metadata:

- `HDIPPassportCredential`
- `KycCredential`

The Phase 1 reusable KYC credential contains a normalized claim set only:

- `subjectReference`
- `fullLegalName`
- `dateOfBirth`
- `countryOfResidence`
- `documentCountry`
- `kycLevel`
- `verifiedAt`
- `expiresAt`

Phase 1 does not include raw document numbers, document images, liveness media, sanctions evidence bundles, or free-form onboarding payloads in the issued credential boundary.

### What the credential record is in Phase 1

The Phase 1 credential record is the authoritative HDIP control-plane record for one issued reusable KYC credential.
It is not only a copy of the issued Phase 1 credential artifact and it is not only a UI summary.

At the contract boundary, the credential record consists of:

- a stable `credentialId`
- `issuerId`
- `templateId`
- the normalized KYC claims listed above
- issuance timestamps
- current credential status
- a status reference usable by verifier evaluation
- a digest of the issued Phase 1 credential artifact
- the Phase 1 opaque credential artifact itself or a durable retrieval reference owned by `issuer-api`

### What is issued vs referenced

What is issued in Phase 1:

- an issuer-produced opaque Phase 1 credential artifact as clarified by ADR 0009
- a stable `credentialId`
- a status reference tied to that credential

What is referenced in Phase 1 verification:

- the opaque credential artifact is the primary verifier submission payload
- `credentialId` may be supplied as an optimization or traceability aid, but it is not trusted without matching the submitted artifact and issuer context
- verifier evaluation references persisted credential status and issuer trust state; it does not rely on raw KYC onboarding evidence
- the verifier does not treat the Phase 1 artifact as cryptographically verifiable

### Service ownership in Phase 1

`issuer-api` owns in Phase 1:

- credential template lookup for issuance
- issuance request validation
- orchestration of credential materialization
- persistence of the credential record
- issuer-facing credential retrieval
- credential status mutation for the issuing organization
- issuance audit emission

`verifier-api` owns in Phase 1:

- verification request intake
- deterministic evaluation of one submitted KYC credential
- persistence of verification request and result records
- verifier-facing result retrieval
- verification audit emission

`trust-registry` owns in Phase 1:

- issuer trust records
- issuer active/suspended state
- allowed template registrations for trusted issuers
- issuer verification-key references needed by verification

Deferred services such as `credential-issuance`, `verification`, `credential-status`, and `audit-compliance` may take over this logic later, but they do not own it in Phase 1.

### Issuance request and response boundary

Phase 1 issuance uses a dedicated write endpoint under `issuer-api`:

- `POST /v1/issuer/credentials`

The request boundary includes:

- `templateId`
- `subjectReference`
- normalized KYC claims
- optional caller-supplied idempotency key via header or transport metadata

The request body must not carry authoritative `issuerId`.
The issuer organization is derived from the authenticated caller context.

The response boundary includes:

- `credentialId`
- `issuerId`
- `templateId`
- `status`
- `issuedAt`
- `expiresAt`
- `statusReference`
- the opaque Phase 1 credential artifact

### Verification submission and result boundary

Phase 1 verification uses a dedicated write endpoint under `verifier-api`:

- `POST /v1/verifier/verifications`

The request boundary includes:

- `policyId`
- the opaque credential artifact
- optional `credentialId` for traceability
- optional caller-supplied idempotency key via header or transport metadata

The verifier organization is derived from authenticated caller context and is not trusted from the request body.

Phase 1 verification is synchronous and deterministic.
The response boundary includes:

- `verificationId`
- `credentialId` when resolvable
- `issuerId`
- `decision`
- `reasonCodes`
- `evaluatedAt`
- the credential status snapshot used for the decision

Suspended or otherwise non-active issuer trust state returns `deny` in deterministic Phase 1 as clarified by ADR 0009.

`GET /v1/verifier/verifications/{verificationId}` is part of the Phase 1 read boundary for verifier-owned retrieval and auditability.

### What is explicitly deferred to Phase 2

Phase 1 does not include:

- holder wallet delivery or presentation UX
- OpenID4VCI issuance transport
- OpenID4VP presentation transport
- selective disclosure
- proof-based verification
- multi-credential evaluation
- delegated presentations
- asynchronous case-management workflows
- cross-vertical trust and reputation composition
- richer authorization beyond the minimum caller boundary in ADR 0008

## Alternatives considered

### Internal reference-only credential with no issued artifact

Rejected because it would not satisfy the reusable credential goal and would turn Phase 1 into a database-only KYC lookup product.

### Full wallet and presentation flow in Phase 1

Rejected because it would pull OpenID transport, holder UX, and proof semantics into a slice that is explicitly supposed to stay narrow.

### Activating deferred issuance and verification services immediately

Rejected because the current repo reality already has `issuer-api` and `verifier-api` as the active foundation surfaces, and splitting sooner would add orchestration cost before the first real flow exists.

## Security impact

Positive.
This boundary makes issuance and verification explicit privileged actions, defines what data is trusted from caller context versus request bodies, and prevents raw onboarding evidence from quietly entering the first reusable credential model.

## Privacy impact

Positive with explicit limits.
Phase 1 still carries more disclosure than later holder-controlled flows, but this ADR constrains the issued claim set and excludes raw KYC evidence from the credential boundary.

## Migration / rollback

If the Phase 1 template naming or claim set needs refinement before code lands, update this ADR and the paired persistence and auth ADRs together.
Do not implement issuance or verification against a partially updated contract boundary.

Later migration to dedicated issuance or verification services is allowed, but only if the externally visible Phase 1 contract boundary and auditability are preserved or versioned explicitly.

## Consequences

- the existing stub metadata endpoints remain read-only foundation behavior and are not the real Phase 1 issuance or verification APIs
- the first real verifier flow will evaluate a submitted opaque credential artifact, not a proof or wallet presentation
- `issuer-api` and `verifier-api` will own real business logic in Phase 1, which keeps the slice implementable without activating every deferred service
- the next implementation slice must add new schema contracts instead of silently reusing stub result semantics as the product contract
- deterministic Phase 1 slices must not describe the artifact as signed or cryptographically verifiable

## Open questions

- whether the final Phase 1 name claim should remain `fullLegalName` or split into separate name fields
- whether the opaque artifact is returned inline only or also retrievable through an issuer-authenticated read endpoint in the first issuance slice
- whether `reasonCodes` should be paired with optional human-readable messages in the initial verification result contract

## Related plans, PRs, and issues

- `docs/plans/active/0006-phase1-kyc-credential-and-verifier-api.md`
- `docs/adr/0007-phase1-state-and-persistence-model.md`
- `docs/adr/0008-phase1-auth-and-attribution-boundary.md`
- `docs/adr/0009-phase1-opaque-artifact-and-suspended-issuer-policy.md`
- `docs/threat-model/full/0003-phase1-kyc-issuance-verification-and-auth.md`
