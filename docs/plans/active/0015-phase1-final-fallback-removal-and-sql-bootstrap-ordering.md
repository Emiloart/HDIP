# 0015 Phase 1 Final Fallback Removal And SQL Bootstrap Ordering

- Status: active
- Date: 2026-04-22
- Owners: repository maintainer

## Objective

Retire the remaining explicit `transitional-json` compatibility mode once the repo is ready to rely on SQL-primary Phase 1 state everywhere in scope, and tighten operational sequencing around `phase1sql migrate up` and `phase1sql bootstrap trust` without changing public Phase 1 issuer or verifier contracts.

## Scope

- evaluate and, if governed repo truth still permits, remove the remaining explicit `transitional-json` service runtime mode
- replace any remaining compatibility-only runtime dependencies on the JSON state path with test-local helpers or fixtures
- tighten primary SQL startup and readiness expectations around migrated schema and trust bootstrap ordering
- harden deployment-facing documentation or narrow automation only as needed to keep `phase1sql` migration and trust bootstrap reproducible
- preserve current Hydra trust-runtime auth boundaries and current public Phase 1 contracts

## Out of scope

- public Phase 1 issuer or verifier contract changes
- broader issuer/verifier auth rollout
- wallet flows, proof verification, selective disclosure, chain anchoring, or cross-vertical behavior
- broader trust-registry product APIs
- platform-wide migration or rollout framework redesign

## Affected files, services, or packages

- `docs/plans/archive/0014-phase1-sql-primary-hardening-and-fallback-retirement.md`
- `docs/plans/active/0015-phase1-final-fallback-removal-and-sql-bootstrap-ordering.md`
- `docs/threat-model/full/0003-phase1-kyc-issuance-verification-and-auth.md`
- `services/internal/phase1sql/`
- `services/issuer-api/internal/`
- `services/verifier-api/internal/`
- `services/trust-registry/internal/`
- deployment or runtime docs only if strictly required for SQL-primary ordering clarity

## Assumptions

- ADR 0010 remains the governing source for internal trust identity and SQL lifecycle shape
- SQL-primary is now the default governed Phase 1 runtime path
- `transitional-json` is now explicit opt-in only and is no longer a silent fallback
- public Phase 1 contracts remain unchanged

## Risks

- removing the explicit transitional mode too early can break compatibility workflows and test seams that still rely on it
- operator mis-ordering of schema migration and trust bootstrap can still leave services unavailable until deployment sequencing is tightened
- overreaching rollout automation would widen scope beyond the current Phase 1 need

## Validation steps

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `npm run schema:validate`
- `bash scripts/validate.sh`

## Rollback or containment notes

If full fallback removal or bootstrap-ordering hardening is incorrect, restore the explicit transitional mode rather than reintroducing an implicit fallback.
Do not leave services in a mixed state where startup and readiness semantics differ unpredictably between SQL-primary and compatibility modes.

## Open questions

- whether the remaining explicit transitional mode can be removed immediately or should survive one more slice for compatibility and tests
- whether SQL bootstrap-ordering hardening needs repo-local automation, documentation only, or a narrower runtime preflight in the next slice
