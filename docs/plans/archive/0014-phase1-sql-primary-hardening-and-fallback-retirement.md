# 0014 Phase 1 SQL Primary Hardening And Fallback Retirement

- Status: completed
- Date: 2026-04-22
- Owners: repository maintainer

## Objective

Constrain the remaining transitional JSON runtime fallback out of the governed primary Phase 1 path and harden the operational lifecycle around Hydra-backed trust reads and the explicit `phase1sql` migration/bootstrap commands, without changing the public Phase 1 issuer or verifier contracts.

## Scope

- make the explicit `phase1sql` lifecycle the operationally primary path in local and deployment documentation
- tighten startup, readiness, and deployment expectations around migrated and bootstrapped SQL state
- make SQL-primary the default governed runtime path and constrain JSON fallback to an explicit transitional mode only
- harden the Hydra trust-runtime path with narrow operational safeguards such as clearer config errors, timeout handling, and failure visibility
- add validation and integration coverage for migrated SQL startup and fallback-retirement behavior

## Out of scope

- public Phase 1 contract changes
- broader issuer/verifier public auth rollout
- wallet flows, proof verification, selective disclosure, chain anchoring, or cross-vertical behavior
- broader trust-registry product APIs
- full platform-wide IAM or migration framework redesign

## Affected files, services, or packages

- `docs/plans/archive/0013-phase1-hydra-internal-trust-auth-and-phase1sql-lifecycle.md`
- `docs/plans/active/0014-phase1-sql-primary-hardening-and-fallback-retirement.md`
- `docs/threat-model/full/0003-phase1-kyc-issuance-verification-and-auth.md`
- `services/internal/phase1sql/`
- `services/issuer-api/internal/phase1/`
- `services/trust-registry/internal/phase1/`
- `services/verifier-api/internal/phase1/`
- deployment or runtime docs only if strictly required for primary-path clarity

## Assumptions

- ADR 0010 remains the governing source for internal trust identity and SQL lifecycle shape
- Hydra client-credentials plus introspection is now the implemented trust-runtime auth model
- the explicit `phase1sql` CLI with versioned SQL assets is now the implemented primary SQL lifecycle
- public Phase 1 contracts remain unchanged

## Risks

- constraining fallback paths too aggressively can break local workflows that have not adopted the explicit SQL lifecycle
- Hydra dependency outages still fail trust runtime reads closed and can become more visible as fallback paths shrink
- operator mis-ordering of migration and trust bootstrap can still block startup until deployment automation hardens

## Validation steps

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `npm run schema:validate`
- `bash scripts/validate.sh`

## Rollback or containment notes

If fallback constraint or SQL-primary hardening is incorrect, restore the transitional compatibility path explicitly and document the rollback boundary.
Do not leave services in a mixed state where startup assumptions differ between issuer, verifier, and trust-registry.

## Open questions

- when the now-explicit transitional JSON mode can be removed entirely rather than remaining available for compatibility and tests
- how much deployment sequencing automation should exist around `phase1sql migrate up` and `phase1sql bootstrap trust` before broader rollout work is governed
