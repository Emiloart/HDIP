# 0003 Foundation Hardening Follow-ups

- Status: archived
- Date: 2026-04-18
- Owners: repository maintainer

## Objective

Harden the foundation slice by preserving typed client error boundaries in TypeScript, making Go service config parsing fail fast on malformed environment values, making Go validation non-mutating, and refreshing stale top-level status docs.

## Scope

- `packages/api-client/` error wrapping and tests
- `services/issuer-api/`, `services/verifier-api/`, and `services/trust-registry/` config parsing behavior
- `scripts/validate-go.sh`
- root status docs that still describe the repo as governance/bootstrap-only
- threat delta for this follow-up slice

## Out of scope

- auth, passkeys, issuance, verification, storage, or wallet logic
- service contract generation
- new dependencies
- ADR changes

## Affected files, services, or packages

- `docs/plans/`
- `docs/threat-model/delta/`
- `packages/api-client/src/`
- `packages/api-client/tests/`
- `services/issuer-api/internal/config/`
- `services/verifier-api/internal/config/`
- `services/trust-registry/internal/config/`
- `scripts/validate-go.sh`
- `README.md`
- `CONTRIBUTING.md`

## Assumptions

- the foundation slice in `e19c82b` is the correct base state
- malformed env values should stop service startup rather than silently falling back
- typed client boundaries should return `ServiceClientError` for malformed transport payloads, not raw parser exceptions

## Risks

- stricter env parsing can break ad hoc local startup if contributors rely on invalid values being ignored
- changing client error wrapping can invalidate assumptions in later consumers if codes are poorly chosen
- doc refresh can drift again if future phase changes are not reflected promptly

## Validation steps

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `bash scripts/validate.sh`

## Rollback or containment notes

These changes are limited to foundation behavior and documentation.
If the stricter parsing behavior proves too disruptive, revert this slice before config handling becomes a dependency of higher-level flows.

## Open questions

- whether Go config parsing should later move into a shared foundation helper once service count grows further
- whether typed contract validation errors in the TS client should eventually expose structured causes for UI-level handling
