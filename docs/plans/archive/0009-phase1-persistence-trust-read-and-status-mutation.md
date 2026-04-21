# 0009 Phase 1 Persistence Trust Read And Status Mutation

- Status: completed
- Date: 2026-04-21
- Owners: repository maintainer

## Objective

Replace the ad hoc service-local Phase 1 runtime placeholders with a repository-backed shared runtime state boundary, add the first real issuer-authenticated credential status mutation path, and make verifier trust and credential reads come from that same explicit runtime boundary.

## Scope

- add a shared repository-backed Phase 1 runtime state package under `services/internal/`
- persist issuer trust records, credential records, verification requests, verification results, and audit records through that runtime boundary
- make `issuer-api` issuance and credential reads use the shared runtime repository adapters
- make `verifier-api` verification evaluation and result reads use the shared runtime repository adapters
- add the first issuer-authenticated credential status mutation endpoint and contract for Phase 1 at `POST /v1/issuer/credentials/{credentialId}/status`
- enforce the accepted `active -> revoked` and `active -> superseded` transition rules
- add tests for persistence continuity, status transitions, trust-driven verifier outcomes, and append-only audit behavior
- update the Phase 1 threat model and contract parity only where required by the new endpoint

## Out of scope

- production auth, token validation, or OIDC rollout
- proof verification, selective disclosure, or wallet flows
- chain anchoring, AI scoring, or cross-vertical logic
- broad `trust-registry` product implementation or admin write APIs
- repurposing stub metadata endpoints
- changes to opaque `credentialArtifact` semantics from ADR 0009
- database migrations or production topology lock-in beyond the narrow runtime adapter needed for this slice

## Affected files, services, or packages

- `docs/plans/active/0009-phase1-persistence-trust-read-and-status-mutation.md`
- `docs/plans/archive/0008-phase1-deterministic-service-logic.md`
- `docs/threat-model/full/0003-phase1-kyc-issuance-verification-and-auth.md`
- `go.work`
- `schemas/json/credentials/`
- `schemas/examples/credentials/`
- `schemas/examples/issuer/`
- `schemas/examples/manifest.json`
- `packages/api-client/src/`
- `services/internal/phase1runtime/`
- `services/issuer-api/internal/app/`
- `services/issuer-api/internal/config/`
- `services/issuer-api/internal/httpapi/`
- `services/issuer-api/internal/phase1/`
- `services/verifier-api/internal/app/`
- `services/verifier-api/internal/config/`
- `services/verifier-api/internal/httpapi/`
- `services/verifier-api/internal/phase1/`

## Assumptions

- ADR 0007 fixes the logical Phase 1 entity set and the accepted status model, but does not require the final production relational engine to land in this slice
- ADR 0008 already authorizes issuer-authenticated revoke and supersede actions using the existing scope boundary
- ADR 0009 remains the source of truth for opaque `credentialArtifact` semantics and the `deny` behavior for suspended or otherwise non-active issuers
- the narrowest acceptable runtime adapter for this slice must stay replaceable behind repository interfaces without changing the public contracts
- trust state may still be bootstrapped with deterministic Phase 1 records in the shared runtime boundary until dedicated `trust-registry` writes land, provided verifier reads no longer use service-local placeholders

## Risks

- a narrow runtime adapter can still be mistaken for long-term production storage if its replaceable role is not documented
- status mutation can drift beyond accepted transitions if handler validation and tests do not guard the terminal-state rules
- verifier trust reads can silently regress back to seeded local behavior if tests do not force shared runtime continuity across services
- audit persistence can over-collect if the write path stores raw credential payloads instead of bounded identifiers and digests

## Validation steps

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `bash scripts/validate.sh`

## Rollback or containment notes

If the shared runtime adapter or status mutation path is incorrect, revert the new repository-backed runtime wiring and the status endpoint together.
Do not keep a half-migrated state where issuer and verifier read different credential or trust sources.

## Open questions

- whether the future `trust-registry` read path should replace the shared runtime issuer record table directly or through a narrower internal read adapter
