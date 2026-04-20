# 0005 Phase 1 HTTP Boundary Skeletons

- Status: accepted
- Date: 2026-04-20
- Owners: repository maintainer

## Change summary

This delta covers the first real Phase 1 handler layer:

- issuer and verifier write/read endpoint skeletons
- request parsing and validation for the Phase 1 contract set
- header-based auth-context extraction placeholders at the service edge
- in-memory or no-op repositories standing in for real persistence

## New or changed entry points

- `POST /v1/issuer/credentials`
- issuer credential detail read endpoint
- `POST /v1/verifier/verifications`
- verifier result read endpoint

## New or changed privileged actions

- skeleton issuance requests now require issuer attribution context
- skeleton verification requests now require verifier attribution context

## Threat delta

- malformed or over-broad request bodies must be rejected before placeholder business flow starts
- missing or mismatched attribution must not silently fall through to handler success paths
- placeholder in-memory repositories must not be mistaken for durable state
- new Phase 1 endpoints must stay clearly separate from the stub metadata endpoints

## Privacy delta

- placeholder handlers must not echo extra request fields or duplicate raw credential payloads into error paths
- attribution extraction must stay bounded to organization, principal, scopes, and authentication reference

## Mitigations

- strict JSON decoding with unknown-field rejection
- explicit attribution validation at the handler edge
- deterministic placeholder responses based on the real contract layer
- tests covering unauthenticated, invalid-payload, and success paths

## Residual risks

- header-based auth extraction is only a skeleton and provides no cryptographic assurance
- in-memory repositories can hide persistence concerns until the next slice if their limits are not kept explicit

## Validation impact

- full repo validation remains required
- new handler tests must cover contract parsing and auth-context failures

## Related ADRs, plans, PRs, and issues

- `docs/plans/archive/0007-phase1-http-boundary-skeletons.md`
- `docs/adr/0006-phase1-credential-and-issuance-boundary.md`
- `docs/adr/0007-phase1-state-and-persistence-model.md`
- `docs/adr/0008-phase1-auth-and-attribution-boundary.md`
- `docs/threat-model/full/0003-phase1-kyc-issuance-verification-and-auth.md`
