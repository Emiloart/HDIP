# 0008 Phase 1 Deterministic Service Logic

- Status: completed
- Date: 2026-04-21
- Owners: repository maintainer

## Objective

Implement the first real deterministic Phase 1 issuer and verifier service logic behind the existing HTTP handlers using the accepted Phase 1 contracts, the ADR 0009 opaque `credentialArtifact` semantics, the approved suspended-issuer `deny` policy, and the current in-memory repository boundary.

## Scope

- implement deterministic issuance record creation in `issuer-api`
- implement deterministic opaque artifact materialization and digest continuity
- implement deterministic verification request persistence and evaluation in `verifier-api`
- enforce the approved issuer trust outcomes for `active`, `suspended`, and other non-active trust states
- append auditable issuer and verifier records for privileged reads and writes
- extend in-memory repository interfaces and tests only as needed for the approved behavior

## Out of scope

- durable storage, DB adapters, or migrations
- new auth mechanisms or real token validation
- wallet delivery, OpenID transport, selective disclosure, or proof verification
- broader trust policy redesign
- new public endpoints beyond the landed Phase 1 handler surface
- changes to stub metadata endpoint behavior
- production signing, proof-bearing artifacts, or key materialization

## Affected files, services, or packages

- `docs/plans/archive/0008-phase1-deterministic-service-logic.md`
- `services/issuer-api/internal/httpapi/`
- `services/issuer-api/internal/phase1/`
- `services/verifier-api/internal/httpapi/`
- `services/verifier-api/internal/phase1/`
- `packages/go/foundation/authctx/` if narrow auth-context helpers need adjustment
- `packages/go/foundation/httpx/` only if existing request handling helpers need narrow extension

## Assumptions

- the schema layer and TypeScript contract parity already reflect the accepted Phase 1 `credentialArtifact` model
- header-based attribution extraction remains the approved temporary auth-context boundary for this slice
- in-memory repositories are sufficient for deterministic Phase 1 logic and tests, provided they reflect the accepted ownership and minimization rules
- trust lookup may remain repository-backed inside each service for this slice without activating a runtime `trust-registry` integration

## Risks

- handler code can still over-assume future production trust or signing semantics if artifact materialization is not kept explicitly opaque
- deterministic verification can drift from ADR 0009 if non-active issuer trust states are not normalized consistently
- audit records can become over-broad if tests do not guard against credential payload duplication
- in-memory repositories can hide continuity bugs if artifact digest and lookup behavior are not asserted directly

## Validation steps

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `bash scripts/validate.sh`

## Rollback or containment notes

If deterministic service behavior proves incorrect, revert the Phase 1 service-logic implementation while preserving ADR 0009 and the contract layer.
Do not keep partially real issuer or verifier behavior that disagrees with the accepted opaque artifact or suspended-issuer policy.

## Open questions

- whether Phase 1 should emit additional deterministic `review` outcomes beyond the required `deny` cases once concrete policy definitions exist
- whether credential continuity should remain artifact-digest-based only in this slice or gain a service-local artifact parser helper in a later slice
