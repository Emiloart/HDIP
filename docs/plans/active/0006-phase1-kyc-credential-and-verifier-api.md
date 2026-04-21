# 0006 Phase 1 KYC Credential And Verifier API

- Status: active
- Date: 2026-04-20
- Owners: repository maintainer

## Objective

Lock the Phase 1 design baseline for a reusable KYC credential and a real verifier API so the next implementation slice can start immediately without reopening credential, persistence, auth, attribution, or trust-boundary decisions.

## Scope

- define the Phase 1 reusable KYC credential boundary
- define the Phase 1 issuance and verification contract boundaries
- define the minimum persisted entity set and status model
- define the issuer operator and verifier integrator auth and attribution boundary
- define the full threat model for Phase 1 issuance, verification, persistence, and caller identity
- preserve compatibility with the schema-first, typed-client, and service-boundary direction already landed

## Out of scope

- service, database, UI, or auth implementation code
- holder wallet flows
- OpenID4VCI and OpenID4VP transport
- selective disclosure, proof-based verification, or BBS flows
- AI risk scoring
- cross-vertical trust products
- advanced authorization models beyond the minimum Phase 1 boundary
- regional multi-cluster rollout design

## Affected files, services, or packages

- `docs/plans/active/0006-phase1-kyc-credential-and-verifier-api.md`
- `docs/adr/0006-phase1-credential-and-issuance-boundary.md`
- `docs/adr/0007-phase1-state-and-persistence-model.md`
- `docs/adr/0008-phase1-auth-and-attribution-boundary.md`
- `docs/threat-model/full/0003-phase1-kyc-issuance-verification-and-auth.md`
- next implementation slice targets:
  - `schemas/json/credentials/`
  - `schemas/json/issuer/`
  - `schemas/json/verifier/`
  - `packages/api-client/`
  - `services/issuer-api/`
  - `services/verifier-api/`
  - `services/trust-registry/`
  - `packages/go/foundation/`

## Assumptions

- JSON Schema remains the canonical source for transport contracts in Phase 1.
- Phase 1 builds on the existing `issuer-api`, `verifier-api`, and `trust-registry` service skeletons rather than activating all deferred backend services at once.
- The existing stub metadata endpoints remain foundation/stub behavior and must not be reinterpreted as real issuance or verification APIs.
- Phase 1 can use a relational persistence model compatible with the accepted CockroachDB baseline without forcing full production topology decisions in the first code slice.
- Phase 1 auth can be implemented behind narrow service-edge abstractions that remain compatible with the accepted OIDC/OAuth direction from ADR 0002.
- Raw KYC evidence ingestion is outside the Phase 1 issuance API; the issuer boundary receives normalized, already-verified KYC claims.
- Deterministic Phase 1 slices use the opaque credential artifact and suspended-issuer `deny` policy clarified in ADR 0009.

## Risks

- contract and storage design can overreach into Phase 2 wallet and proof flows if the credential boundary is not kept narrow
- auth and attribution can be underdesigned if caller identity is treated as a later integration detail
- Phase 1 can accidentally normalize unnecessary PII fields if claim scope is not bounded tightly
- stub-era endpoints can be mistaken for production behavior if the real Phase 1 APIs are not additive and explicit
- trust-registry participation can become a hidden dependency if the verifier flow assumes trust lookup semantics that are not documented

## Validation steps

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`

For the next implementation slice, full validation must also include:

- `bash scripts/validate.sh`

## Rollback or containment notes

These artifacts lock design only.
If review finds the Phase 1 boundary too broad, contain the correction within these five governance files before code starts.
Do not partially implement issuance, verification, or auth against a superseded version of this plan.

## Open questions

- whether the Phase 1 KYC claims set should carry one `fullLegalName` field or separate legal name fields
- whether trust-registry lookup in Phase 1 should expose a narrow internal HTTP read contract or a shared repository abstraction behind service boundaries
- whether Phase 1 verification results need an explicit `error` terminal state in addition to `allow`, `deny`, and `review`

## Exact implementation sequence for the first real Phase 1 code slices

### Slice 1: Phase 1 contracts and schema parity

- add schema-first contracts for:
  - reusable KYC credential artifact metadata
  - issuance request
  - issuance response
  - credential status
  - verification submission
  - verification result
- extend schema examples and parity tests in TypeScript and Go
- add typed TS client methods for the new Phase 1 endpoints without changing UI behavior yet

### Slice 2: Persistence boundary and repository interfaces

- add storage-agnostic repository interfaces in `issuer-api`, `verifier-api`, and `trust-registry`
- add the minimum persisted entity mappings for:
  - issuer records
  - credential records
  - verification request records
  - verification result records
  - audit records
- keep the implementation relational and service-owned without introducing new runtime services

### Slice 3: Auth context, attribution, and audit plumbing

- add narrow service-edge auth context extraction and validation hooks
- enforce issuer operator auth on issuance and status-changing endpoints
- enforce verifier integrator auth on verification endpoints
- append auditable attribution metadata for all privileged writes and sensitive reads

### Slice 4: Real issuer flow

- implement issuance request validation against the new schemas
- persist credential records and issuance audit records
- produce the Phase 1 opaque credential artifact defined by ADR 0009
- add issuer-side credential retrieval and status mutation endpoints for the issuing operator

### Slice 5: Real verifier flow

- implement verification submission and result persistence
- evaluate the submitted credential deterministically against:
  - issuer trust lookup
  - non-active issuer trust state -> `deny`
  - credential status
  - expiry
  - template compatibility
- return real `allow`, `deny`, or `review` results with structured reason codes

### Slice 6: Minimal status and trust integration completion

- wire verifier evaluation to the Phase 1 trust-registry read path
- expose the minimum credential status read behavior needed by issuer and verifier flows
- add audit and integration tests across issuance, verification, and status transitions
