# 0005 Extended Contract Parity And Stub Flow

- Status: archived
- Date: 2026-04-20
- Owners: repository maintainer

## Objective

Extend contract-parity enforcement to the issuer, verifier, and credential-template contracts, then introduce the first stubbed issuer/verifier flow using those contracts across Go service endpoints, typed TypeScript clients, and the existing console shells.

## Scope

- canonical example fixtures and parity validation for:
  - issuer profile
  - verifier policy request
  - verifier result
  - credential template metadata
- TypeScript schema coverage and parity tests for the above contracts
- issuer-api stub endpoints for issuer profile and credential template metadata
- verifier-api stub endpoints for verifier policy request and stub verifier result
- typed client methods for issuer and verifier stub flows
- issuer and verifier console shell updates to render the stub flow results
- required full threat-model update

## Out of scope

- real credential issuance
- proof generation or verification
- auth or authorization changes
- storage or database integration
- trust-registry runtime behavior changes
- mobile wallet logic

## Affected files, services, or packages

- `docs/plans/`
- `docs/threat-model/full/`
- `schemas/`
- `scripts/validate-schemas.mjs`
- `packages/api-client/`
- `packages/go/foundation/`
- `services/issuer-api/`
- `services/verifier-api/`
- `apps/issuer-console/`
- `apps/verifier-console/`

## Assumptions

- JSON Schema remains the canonical source for transport contracts in this stage
- the stub flow should stay read-only and deterministic
- the flow should demonstrate contract shape and service boundaries without pretending real issuance or verification exists

## Risks

- stub endpoints can be mistaken for real product behavior if not clearly labeled
- contract changes across services and TS clients can drift if the example layer is incomplete
- server-rendered app shells can accidentally hard-code future business rules if the presentation stays too specific

## Validation steps

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `bash scripts/validate.sh`

## Rollback or containment notes

This slice is foundation behavior only.
If the stub flow shape is wrong, revert the endpoints, client methods, and shell rendering before downstream work starts depending on them.

## Open questions

- whether the next slice should add trust-registry participation in the flow
- whether stub payload factories should later move into shared contract fixtures or remain service-local
- when the verifier result should transition from GET-based stub output to a real evaluation request/response flow

## Outcome

- Completed on 2026-04-20
- Extended parity fixtures now cover issuer profile, verifier policy request, verifier result, and credential template metadata
- Stub issuer and verifier endpoints are live in the service skeletons and consumed by the console shells through typed clients
