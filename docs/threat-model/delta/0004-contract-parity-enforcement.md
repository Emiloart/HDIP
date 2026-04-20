# 0004 Contract Parity Enforcement

- Status: accepted
- Date: 2026-04-18
- Owners: repository maintainer

## Change summary

Add canonical contract examples and parity checks so baseline transport contracts are validated across JSON Schema, the TypeScript client schemas, and the Go HTTP envelope helpers.

## New or changed entry points

- none

## New or changed privileged actions

- none

## Threat delta

- Reduces the risk of silent contract drift between schema definitions and runtime clients/helpers.
- Improves confidence that common error and health envelopes stay stable before real product flows depend on them.

## Privacy delta

- none

## Mitigations

- validate example fixtures against the canonical JSON Schemas
- run TypeScript parity tests against the same examples used by schema validation
- run Go foundation tests against the same canonical envelope examples

## Residual risks

- parity remains example-based rather than generated from a single typed source
- only the common contracts are covered in this slice

## Validation impact

- schema validation now covers canonical examples in addition to schema compilation
- TypeScript and Go tests cover shared health and error envelope behavior

## Related ADRs, plans, PRs, and issues

- `docs/adr/0004-foundation-service-and-schema-baseline.md`
- `docs/plans/active/0004-contract-parity-enforcement.md`
