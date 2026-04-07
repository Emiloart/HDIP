# 0004 Foundation Service And Schema Baseline

- Status: accepted
- Date: 2026-04-06
- Owners: repository maintainer

## Context

The first HDIP service slice needs stable operational boundaries before any real issuance or verification logic is written.
It also needs shared contracts early enough to prevent drift between Go services, Rust core, and TypeScript apps.

## Decision

Adopt the following foundation service baseline:

- `services/issuer-api`
- `services/verifier-api`
- `services/trust-registry`

Each service must start with:

- startup config loading and validation
- structured safe logging
- request ID middleware
- timeout middleware
- panic recovery middleware
- `/healthz`
- `/readyz`
- graceful shutdown

No database or external dependency is required in this slice unless startup shape truly depends on it.

For cross-service contracts, adopt schema-first JSON contract files under `schemas/json/` for:

- health
- error envelope
- issuer metadata
- verifier request envelope
- verifier result
- credential template metadata

Go skeleton concerns may share a narrow internal library under `packages/go/foundation`, limited to startup, middleware, and response helpers.

## Alternatives considered

### Service-first with no shared contracts

Rejected because drift would happen before the first real flow.

### One generic gateway first

Rejected because it adds speculative routing structure before real duplication exists.

### Immediate database-backed service startup

Rejected because it adds operational coupling before the service interfaces are stable.

## Security impact

Positive.
Standardized startup and response handling reduce accidental unsafe behavior.

## Privacy impact

Positive.
Shared contracts and error envelopes make overexposure easier to spot early.

## Migration / rollback

These are foundation-only boundaries.
If the service split changes before real business logic lands, the scaffolding can be reworked with limited cost.

## Consequences

- Go code will have a small shared foundation package.
- Schemas become the canonical source for basic transport contracts.
- Services must remain boring and consistent until real domain logic justifies divergence.

## Open questions

- whether JSON Schema alone will remain sufficient once Rust and Go type generation becomes necessary
- whether future internal RPC contracts should live beside or separate from public HTTP contracts

## Related plans, PRs, and issues

- `docs/plans/active/0002-foundation-workspace-and-service-skeletons.md`
