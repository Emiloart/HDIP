# 0011 Phase 1 Production Persistence And Trust Runtime Reads

- Status: completed
- Date: 2026-04-21
- Owners: repository maintainer

## Objective

Replace the transitional shared JSON-backed Phase 1 runtime with the governed relational production persistence path already accepted by the repo, move verifier trust reads to trust-registry-owned runtime reads behind the existing narrow trust boundary, and harden idempotency from completed-write replay into reservation-based write protection.

## Scope

- add a Cockroach-compatible relational Phase 1 persistence adapter behind the existing issuer and verifier repository boundaries
- make relational persistence the primary runtime path for credential records, verification request records, verification result records, audit records, and bounded idempotency records
- keep the shared JSON-backed runtime only as a clearly transitional fallback where still needed for local compatibility or targeted tests
- add reservation-state idempotency handling so overlapping same-key writes cannot silently create duplicate completed state
- keep replay and conflict handling caller-bound and operation-bound
- add trust-registry-owned runtime reads for verifier trust lookups through the existing narrow verifier trust-read adapter
- add the minimum trust-registry internal read endpoint and store wiring needed to serve issuer trust snapshots for deterministic Phase 1 verification
- update config and startup wiring so the relational path is explicit and primary
- add tests for relational persistence round trips, trust-registry-owned verifier reads, reservation collisions, replay, and cross-service continuity
- update the full Phase 1 threat model and any plan/runtime notes needed to keep repo truth precise

## Out of scope

- production auth or token validation rollout
- holder wallet flows, proof verification, selective disclosure, or chain anchoring
- broader trust-registry product APIs or trust-registry write administration
- full production topology rollout beyond the narrow relational adapter and initialization required for this slice
- changing the public Phase 1 request or response shapes
- changing opaque `credentialArtifact` semantics from ADR 0009
- event streaming, Temporal workflows, or cross-vertical logic

## Affected files, services, or packages

- `docs/plans/active/0011-phase1-production-persistence-and-trust-runtime-reads.md`
- `docs/plans/archive/0010-phase1-persistence-hardening-idempotency-and-trust-read-adapter.md`
- `docs/threat-model/full/0003-phase1-kyc-issuance-verification-and-auth.md`
- `services/internal/phase1runtime/`
- new shared internal relational persistence package under `services/internal/`
- `services/issuer-api/internal/app/`
- `services/issuer-api/internal/config/`
- `services/issuer-api/internal/httpapi/`
- `services/issuer-api/internal/phase1/`
- `services/verifier-api/internal/app/`
- `services/verifier-api/internal/config/`
- `services/verifier-api/internal/httpapi/`
- `services/verifier-api/internal/phase1/`
- `services/trust-registry/internal/app/`
- `services/trust-registry/internal/config/`
- `services/trust-registry/internal/httpapi/`
- new trust-registry local Phase 1 runtime helpers if needed

## Assumptions

- ADR 0007 already locks a relational persistence model behind service-owned repository interfaces for Phase 1
- ADR 0002 already locks CockroachDB as the primary transactional store, so a Cockroach-compatible SQL adapter does not require a new storage-engine ADR
- stronger idempotency reservation state is justified by the accepted Phase 1 threat model and by the already-landed replay-defense behavior
- trust-registry-owned verifier reads can stay bounded to issuer trust state, allowed template identifiers, and verification-key references
- public Phase 1 HTTP contracts remain unchanged in this slice

## Risks

- introducing SQL without clear fallback semantics can make local execution brittle if the transitional JSON path is not demoted carefully
- trust-registry-owned reads can accidentally widen into a product API surface if the internal response is not kept minimal
- reservation-state idempotency can become privacy-invasive if raw request payloads or excess response data are stored
- schema bootstrap can imply a final migration story if it is not clearly documented as the narrow Phase 1 relational initialization path

## Validation steps

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `npm run schema:validate`
- `bash scripts/validate.sh`

## Rollback or containment notes

If the relational adapter or trust-registry-owned read path is incorrect, revert the SQL-primary wiring and the new trust-registry runtime read path together.
Do not keep a mixed state where verifier trust comes from the trust-registry boundary but issuer and verifier writes still rely on partially migrated persistence.

## Open questions

- whether the eventual production migration story should keep the same transitional JSON fallback at all once Cockroach-backed local workflows exist
- whether future trust-registry runtime reads should add explicit template-policy compatibility data beyond the current minimal issuer trust snapshot
