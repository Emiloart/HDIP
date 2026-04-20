# 0002 Stubbed Issuer And Verifier Flow Threat Model

- Status: accepted
- Date: 2026-04-20
- Owners: repository maintainer

## Change summary

Extend contract parity across the issuer, verifier, and credential-template contracts and add the first stubbed issuer/verifier flow through foundation service endpoints, typed clients, and console shells.
The new flow is deterministic and read-only, but it introduces issuance- and verification-shaped API surfaces that future product logic will build on.

## Assets

- issuer profile metadata
- credential template metadata
- verifier policy request definitions
- verifier decision results
- transport-contract fixtures and schemas
- typed client contract bindings

## Trust boundaries

- issuer-api boundary
- verifier-api boundary
- TypeScript client boundary
- console server-rendering boundary
- schema and example source-of-truth boundary

## Attacker classes

- external callers probing stub endpoints for contract inconsistencies
- contributors accidentally introducing drift between service payloads and client schemas
- abusive verifiers or issuers using stub payloads as if they were production-safe logic
- insiders weakening transport validation during rapid iteration

## Entry points and privileged actions

- `GET /v1/issuer/profile`
- `GET /v1/issuer/templates/{templateId}`
- `GET /v1/verifier/policy-requests/{policyId}`
- `GET /v1/verifier/results/{requestId}/stub`
- server-rendered issuer and verifier console page fetches

No privileged write actions are added in this slice.

## Abuse and misuse cases

- treating stub outputs as authoritative verification decisions
- adding hidden logic in app shells that diverges from service contracts
- exposing overly broad template or verifier metadata in future iterations
- allowing parity examples to drift from actual service responses

## Privacy harms

- future overexposure if stub payloads normalize unnecessary disclosure fields
- console rendering accidentally presenting verifier intent as real user data handling

This slice does not add live user data or credential payloads.

## Mitigations

- stub endpoints remain read-only and deterministic
- console copy explicitly labels the flow as stubbed foundation behavior
- canonical JSON Schema examples enforce contract parity
- typed client methods validate payloads before UI use
- service tests compare responses to deterministic expectations

## Residual risks

- stubs can still be misread as product commitments if later docs are not updated carefully
- trust-registry is still outside the flow, so end-to-end trust composition remains untested
- no real proof input or verifier submission path exists yet

## Validation impact

- full repo validation remains required
- parity tests must cover the new issuer, verifier, and credential-template fixtures
- service tests must cover the new stub endpoints

## Related ADRs, plans, PRs, and issues

- `docs/adr/0004-foundation-service-and-schema-baseline.md`
- `docs/plans/archive/0005-extended-contract-parity-and-stub-flow.md`
