# 0004 Contract Parity Enforcement

- Status: archived
- Date: 2026-04-18
- Owners: repository maintainer

## Objective

Reduce early transport-contract drift by adding canonical contract examples and enforcing parity across JSON Schema, the TypeScript client schemas, and the Go foundation response helpers.

## Scope

- add canonical example fixtures under `schemas/examples/`
- extend schema validation to verify examples against the canonical JSON Schemas
- add TypeScript parity tests for the contracts currently mirrored in `packages/api-client`
- add Go foundation tests for the common HTTP envelope helpers
- document the new contract-example layer in schema docs

## Out of scope

- new production dependencies
- schema generation for Go or Rust
- changes to product flows, auth, issuance, or verification behavior
- changes to non-common schema contracts beyond example coverage

## Affected files, services, or packages

- `docs/plans/`
- `docs/threat-model/delta/`
- `schemas/`
- `scripts/validate-schemas.mjs`
- `packages/api-client/src/`
- `packages/api-client/tests/`
- `packages/go/foundation/httpx/`

## Assumptions

- JSON Schema remains the canonical source for the baseline transport contracts
- the current TypeScript client only needs parity coverage for `error-envelope` and `health`
- Go foundation helpers are the right enforcement point for common HTTP envelopes shared by services

## Risks

- example fixtures can become another drift source if they are not validated against schemas
- partial parity coverage could create a false sense of completeness if not scoped explicitly
- mounted Windows paths remain slower for JS dependency hydration and validation

## Validation steps

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `bash scripts/validate.sh`

## Rollback or containment notes

This slice adds validation and tests only.
If the example/manifest shape proves wrong, remove the new fixtures and validation wiring before downstream code depends on them.

## Open questions

- whether future contract parity should move from example-based checks to code generation
- whether issuer, verifier, and credential template schemas need the same parity fixtures in the next slice
