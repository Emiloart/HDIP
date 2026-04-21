# 0007 Phase 1 State And Persistence Model

- Status: accepted
- Date: 2026-04-20
- Owners: repository maintainer

## Context

Phase 1 requires real issuance, real verification, status handling, and auditable attribution.
The current repo has no runtime persistence model, and Phase 1 cannot be implemented safely if storage shape, state ownership, and minimization rules are left implicit.

At the same time, the repo should not overcommit to the full long-term platform storage topology before the first real reusable KYC flow exists.

ADR 0009 clarifies the deterministic Phase 1 artifact semantics and resolves suspended-issuer verifier behavior for the purposes of this record.

## Decision

### Minimum persisted entity set for Phase 1

Phase 1 persists exactly these logical entity classes:

- `issuer_record`
- `credential_record`
- `verification_request_record`
- `verification_result_record`
- `audit_record`

A dedicated subject table is deferred.
Phase 1 uses `subjectReference` inside `credential_record` rather than introducing a broader subject graph prematurely.

### Issuer records

`issuer_record` is owned by `trust-registry` and is the minimum authoritative trust entry for a Phase 1 issuer.
It includes:

- `issuerId`
- issuer display metadata
- issuer trust state
- allowed template identifiers
- active verification-key references
- created and updated timestamps

`issuer_record` is read by `issuer-api` and `verifier-api`, but its trust state and key references are owned by `trust-registry`.

### Credential records

`credential_record` is owned by `issuer-api`.
It includes the minimum data required to support reuse, verification, status checks, and audit:

- `credentialId`
- `issuerId`
- `templateId`
- `subjectReference`
- normalized KYC claims
- artifact digest
- issued opaque Phase 1 artifact or durable artifact retrieval reference
- `issuedAt`
- `expiresAt`
- current credential status
- status-updated timestamp
- superseding credential reference when applicable

`credential_record` must not persist raw document scans, raw KYC evidence bundles, or free-form issuer notes in Phase 1.

### Verification request records

`verification_request_record` is owned by `verifier-api`.
It includes:

- `verificationId`
- `verifierId`
- submitted credential artifact digest
- resolved `credentialId` when available
- `policyId`
- request timestamp
- attribution fields from the authenticated caller context
- idempotency key when supplied

The full submitted opaque credential artifact should not be duplicated into the verification request record by default.
If temporary raw artifact storage is needed for evaluation, it must be bounded to the request lifecycle rather than becoming the durable request record.

### Verification result records

`verification_result_record` is owned by `verifier-api`.
It includes:

- `verificationId`
- `decision`
- `reasonCodes`
- issuer trust snapshot reference
- credential status snapshot
- evaluation timestamp
- response version metadata

`verification_result_record` is immutable after write except for tightly bounded repair tooling that would itself require explicit future governance.

### Audit records

`audit_record` is append-only and is written by the service performing the action.
Phase 1 audit records must capture:

- actor identity and actor type
- organization identity
- action name
- resource type and resource identifier
- request identifier
- idempotency key when present
- outcome
- timestamp
- service name

Audit records must reference credential and verification artifacts by stable identifiers or digests, not by duplicating raw credential payloads.

### Credential status model

Phase 1 uses the following credential status model:

- `active`
- `revoked`
- `superseded`

`expired` is derived from `expiresAt` and is not a separately persisted primary status value.

Phase 1 does not publish a standalone public status list service.
Instead, `credential_record` owns the authoritative status fields while preserving a status reference shape that can later map to Bitstring Status List or a dedicated status service.

### Ownership of state transitions

`issuer-api` owns:

- creation of `credential_record`
- transitions from `active` to `revoked`
- transitions from `active` to `superseded`

`verifier-api` owns:

- creation of `verification_request_record`
- creation of `verification_result_record`

`trust-registry` owns:

- issuer trust activation or suspension
- verification-key reference changes

For deterministic Phase 1 verifier behavior, non-active issuer trust state results in verifier decision `deny` as clarified by ADR 0009.

`audit_record` is written by whichever service performs the transition or sensitive read.

### Minimal retention and minimization expectations

Phase 1 retention must follow these minimum rules:

- keep only the normalized claims required for the reusable KYC credential
- do not store raw upstream KYC evidence in Phase 1 runtime records
- keep verification request records to the minimum fields required for attribution, replay defense, and audit
- keep verification result records as decision snapshots, not as expanded copies of credential payloads
- keep audit records long enough for compliance and incident review, but without raw credential duplication

Exact retention durations remain deployment-policy concerns and are not fixed by this ADR.

### Storage assumptions

Phase 1 uses a relational persistence model behind service-owned repository interfaces.
The implementation may start with one logical relational database or one operational relational cluster with service-owned tables or schemas.

This decision is intentionally compatible with the accepted CockroachDB direction from ADR 0002, but it does not force full global production topology, event streaming, or workflow orchestration into the first real code slice.

Phase 1 does not require:

- a dedicated persistence service
- a separate status service
- Temporal orchestration
- NATS-backed state transitions

## Alternatives considered

### No persistence until after real API work starts

Rejected because reusable KYC, status handling, and auditable verification are impossible without durable state.

### Separate dedicated services for issuance, status, verification, and audit in Phase 1

Rejected because it would over-split the first real flow before the behavior is proven.

### Storing raw credential evidence for easier debugging

Rejected because it conflicts with data minimization and would normalize unsafe handling of sensitive identity artifacts.

## Security impact

Positive.
This ADR makes state ownership explicit, constrains mutable records, and reduces the chance that privileged flows are implemented against an ad hoc or overly permissive storage shape.

## Privacy impact

Positive.
The decision explicitly minimizes persistent storage of raw credential artifacts and excludes raw KYC proofing evidence from Phase 1 records.

## Migration / rollback

If the initial relational layout needs refinement, change repository and table shape without widening the external contract or minimization boundary.
If later phases require dedicated services, migrate ownership deliberately and preserve audit continuity.

Do not introduce new persisted entity classes casually during Phase 1 implementation without checking whether they belong in a follow-up ADR.

## Consequences

- `trust-registry` becomes a real runtime dependency in Phase 1 for issuer trust state
- `issuer-api` and `verifier-api` gain durable business state instead of remaining stateless skeletons
- a dedicated subject model is deferred, which keeps Phase 1 narrower but means `subjectReference` discipline matters
- the next code slice must add storage interfaces and migration notes before business logic lands

## Open questions

- whether verification request retention needs a shorter operational window than verification result retention
- whether the initial artifact retrieval strategy should store the full opaque artifact inline or in bounded object storage behind the same logical record

## Related plans, PRs, and issues

- `docs/plans/active/0006-phase1-kyc-credential-and-verifier-api.md`
- `docs/adr/0006-phase1-credential-and-issuance-boundary.md`
- `docs/adr/0008-phase1-auth-and-attribution-boundary.md`
- `docs/adr/0009-phase1-opaque-artifact-and-suspended-issuer-policy.md`
- `docs/threat-model/full/0003-phase1-kyc-issuance-verification-and-auth.md`
