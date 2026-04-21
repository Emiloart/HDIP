# 0010 Phase 1 Persistence Hardening Idempotency And Trust Read Adapter

- Status: completed
- Date: 2026-04-21
- Owners: repository maintainer

## Objective

Harden the current replaceable Phase 1 shared runtime by adding persisted idempotency and replay handling for write operations, introducing an explicit verifier trust read adapter, and removing misleading runtime-state naming that implies a backend which is not in use.

## Scope

- add bounded persisted idempotency records for Phase 1 write operations under the shared runtime state boundary
- make `issuer-api` issuance and status writes replay prior successful results when `Idempotency-Key` is reused with the same caller, operation, and semantically equivalent request
- make `verifier-api` verification writes replay prior successful results when `Idempotency-Key` is reused with the same caller, operation, and semantically equivalent request
- reject conflicting idempotency-key reuse cleanly and audibly without creating duplicate state
- introduce a verifier-local trust read adapter that exposes only the minimum deterministic Phase 1 trust view needed for evaluation
- keep the explicit trust read adapter backed by the current shared file-backed runtime until dedicated trust-registry runtime reads land
- rename misleading shared-runtime path defaults and config names to neutral state-oriented names while preserving backward compatibility where needed
- add continuity and replay tests that prove shared-state visibility across runtime instances and service boundaries
- update the active threat model for replay, conflict handling, and the explicit trust-read boundary

## Out of scope

- replacing the current file-backed runtime with the final production storage backend
- adding database migrations or locking in a production persistence topology
- production auth, token validation, or new auth mechanisms
- trust-registry write ownership or broader trust-registry product APIs
- wallet flows, proof verification, selective disclosure, chain anchoring, AI scoring, or cross-vertical logic
- changes to the approved opaque `credentialArtifact` semantics from ADR 0009
- public Phase 1 contract changes beyond narrow internal naming or documentation clarifications required by this slice

## Affected files, services, or packages

- `docs/plans/active/0010-phase1-persistence-hardening-idempotency-and-trust-read-adapter.md`
- `docs/plans/archive/0009-phase1-persistence-trust-read-and-status-mutation.md`
- `docs/threat-model/full/0003-phase1-kyc-issuance-verification-and-auth.md`
- `services/internal/phase1runtime/`
- `services/issuer-api/internal/config/`
- `services/issuer-api/internal/httpapi/`
- `services/issuer-api/internal/phase1/`
- `services/verifier-api/internal/config/`
- `services/verifier-api/internal/httpapi/`
- `services/verifier-api/internal/phase1/`

## Assumptions

- ADR 0006 and ADR 0008 make caller-supplied idempotency keys optional but acceptable on the current write endpoints
- ADR 0007 already requires persisted verification-request fields sufficient for attribution, replay defense, and audit, so bounded idempotency metadata fits the accepted Phase 1 persistence direction
- ADR 0009 remains the source of truth for opaque artifact semantics and the required `deny` behavior for suspended or otherwise non-active issuers
- the current shared runtime remains intentionally replaceable, so this slice must harden it without presenting it as the final production persistence backend
- trust reads can stay backed by the shared runtime for now if verifier logic no longer conceptually depends on a generic issuer-record read path

## Risks

- replay handling can become privacy-invasive if idempotency persistence stores raw request payloads rather than bounded fingerprints and response snapshots
- conflicting key reuse can become ambiguous if the operation or caller binding is underspecified
- verifier trust handling can silently drift back to generic issuer-record behavior if tests do not force the explicit adapter path
- renaming runtime state config or defaults can break local execution if backward compatibility with the prior runtime-path environment variable is not preserved

## Validation steps

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `bash scripts/validate.sh`
- `npm run schema:validate`

## Rollback or containment notes

If replay handling or the trust adapter is incorrect, revert the shared-runtime hardening and the issuer/verifier idempotency wiring together.
Do not keep a partially hardened state where duplicate writes are possible but replay metadata or trust reads disagree across services.

## Open questions

- whether a later production storage slice should keep replay response snapshots in the operational store or move them to a dedicated idempotency table with stricter lifecycle controls
- whether future trust-registry runtime reads should preserve the same verifier-local trust adapter shape exactly or add explicit template-policy compatibility data
