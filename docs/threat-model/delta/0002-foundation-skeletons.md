# 0002 Foundation Skeletons Threat Delta

- Status: accepted
- Date: 2026-04-06
- Owners: repository maintainer

## Change summary

This delta covers the first executable foundation slice:

- Rust core crate skeletons
- Go service skeletons with health and readiness endpoints
- frontend shells for issuer and verifier consoles
- schema-first transport contracts

## New or changed entry points

- service HTTP endpoints for `/healthz` and `/readyz`
- service startup config ingestion
- frontend environment loading
- frontend API client boundaries

## New or changed privileged actions

None yet.
No auth, admin operations, credential issuance, or verification actions are part of this slice.

## Threat delta

- health endpoints must not leak sensitive runtime details
- structured logs must stay free of secrets and raw request payloads
- request ID middleware must not trust unsafe inbound values blindly
- config loading must fail closed on malformed required values
- shared packages must not quietly become homes for privileged logic

## Privacy delta

- frontend shells must not collect or emit real user identity data
- example and test data must remain synthetic and non-sensitive
- schemas must not imply broader disclosure than intended

## Mitigations

- standardized error envelope
- safe logging defaults
- minimal health responses
- no business logic in this slice
- governance and validation gates retained

## Residual risks

- toolchain setup complexity may slow consistent local validation
- early shared packages can still become catch-all utilities if not watched

## Validation impact

Validation must expand to include Rust, Go, TypeScript, and schema checks in this slice.

## Related ADRs, plans, PRs, and issues

- `docs/adr/0003-rust-core-crate-boundaries.md`
- `docs/adr/0004-foundation-service-and-schema-baseline.md`
- `docs/adr/0005-web-surface-separation.md`
- `docs/plans/active/0002-foundation-workspace-and-service-skeletons.md`
