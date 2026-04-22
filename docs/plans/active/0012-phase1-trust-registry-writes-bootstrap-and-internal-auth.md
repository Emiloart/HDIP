# 0012 Phase 1 Trust Registry Writes Bootstrap And Internal Auth

- Status: active
- Date: 2026-04-22
- Owners: repository maintainer

## Objective

Make `trust-registry` the real owner of deterministic Phase 1 issuer trust writes and bootstrap on the primary SQL-backed runtime path, and add governed internal auth for the trust runtime reads consumed by `verifier-api`, without changing current public issuer or verifier contracts.

## Scope

- add a trust-registry-owned Phase 1 trust bootstrap and update flow over the existing issuer trust record model
- keep the SQL-backed runtime path primary for trust ownership while preserving the JSON fallback only as a transitional compatibility path
- add bounded audit records for trust write or bootstrap actions
- add governed internal bearer auth for trust-registry internal runtime-read endpoints
- update `verifier-api` trust client config and transport so runtime trust reads authenticate explicitly and fail closed
- preserve the existing narrow verifier trust-read boundary and deterministic issuer trust policy
- add tests for bootstrap or update continuity, internal auth success and failure, and verifier deny behavior through the authenticated trust-registry read path
- update threat-model and slice accounting docs needed to keep repo truth precise

## Out of scope

- broad trust-registry product APIs or external admin surfaces
- broader production auth rollout for issuer or verifier public APIs
- wallet flows, proof verification, selective disclosure, chain anchoring, or cross-vertical trust behavior
- changes to current public Phase 1 issuer or verifier request and response shapes
- replacing the current SQL adapter with a platform-wide migration system
- final production token issuance or service identity infrastructure

## Affected files, services, or packages

- `docs/plans/active/0012-phase1-trust-registry-writes-bootstrap-and-internal-auth.md`
- `docs/plans/archive/0011-phase1-production-persistence-and-trust-runtime-reads.md`
- `docs/threat-model/full/0003-phase1-kyc-issuance-verification-and-auth.md`
- `services/internal/phase1sql/`
- `services/trust-registry/internal/app/`
- `services/trust-registry/internal/config/`
- `services/trust-registry/internal/httpapi/`
- `services/trust-registry/internal/phase1/`
- `services/verifier-api/internal/config/`
- `services/verifier-api/internal/httpapi/`
- `services/verifier-api/internal/phase1/`

## Assumptions

- ADR 0007 already locks `trust-registry` ownership of issuer trust state, allowed template registrations, and verification-key references
- ADR 0008 already permits a narrow bearer-style auth abstraction without requiring the full long-term auth platform to land in this slice
- the narrowest acceptable internal auth mechanism is a service-to-service bearer token enforced only on trust-registry internal Phase 1 routes
- the narrowest acceptable trust bootstrap path is trust-registry-local bootstrap or apply logic over the existing issuer trust record shape rather than a broad external admin API
- public Phase 1 issuer and verifier contracts remain unchanged in this slice

## Risks

- internal trust auth can become an accidental second public auth model if route scoping is not strict
- bootstrap or update logic can drift into a broad admin surface if it is exposed beyond the minimum trust-registry-owned path
- trust write audits can over-collect if they store raw configuration payloads rather than bounded trust metadata
- SQL-primary startup can become misleading if the fallback path is not clearly documented as transitional only

## Validation steps

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `npm run schema:validate`
- `bash scripts/validate.sh`

## Rollback or containment notes

If the trust bootstrap or internal auth layer is incorrect, revert the trust-registry write path and verifier trust-client auth changes together.
Do not keep a mixed state where verifier runtime reads require internal auth but trust-registry ownership of trust data is still partially ad hoc.

## Open questions

- whether a later trust-registry product slice should expose a richer administrative write surface or keep trust writes behind dedicated operator tooling
- whether the future production internal auth layer should replace the static bearer token with a stronger service identity scheme without changing the current verifier trust-read boundary
