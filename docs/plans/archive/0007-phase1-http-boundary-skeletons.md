# 0007 Phase 1 HTTP Boundary Skeletons

- Status: completed
- Date: 2026-04-20
- Owners: repository maintainer

## Objective

Add the first real Phase 1 HTTP boundary skeletons for issuance and verification so HDIP exposes the correct authenticated endpoint shape, request parsing, and deterministic placeholder responses before real business logic, DB adapters, or auth middleware land.

## Scope

- add Phase 1 issuer and verifier HTTP endpoints alongside the existing stub metadata endpoints
- parse and validate Phase 1 request bodies against the new contract layer
- extract issuer and verifier attribution through the new auth-context interfaces at the handler edge
- inject in-memory or no-op repository implementations instead of real persistence
- return deterministic placeholder responses using the real Phase 1 response contracts
- add handler tests covering success and key failure paths

## Out of scope

- real issuance logic
- real verification decision logic
- DB-backed repositories or migrations
- auth middleware, token verification, or OIDC integration
- trust-registry runtime calls
- UI integration
- wallet flows
- proof verification or selective disclosure

## Affected files, services, or packages

- `docs/plans/active/0007-phase1-http-boundary-skeletons.md`
- `docs/threat-model/delta/0005-phase1-http-boundary-skeletons.md`
- `packages/go/foundation/httpx/`
- `services/issuer-api/internal/httpapi/`
- `services/issuer-api/internal/phase1/`
- `services/verifier-api/internal/httpapi/`
- `services/verifier-api/internal/phase1/`
- `README.md`
- `docs/repo-structure.md`

## Assumptions

- placeholder caller attribution can be represented by header-based extraction without claiming real auth validation
- in-memory repositories are sufficient for boundary tests and must not be mistaken for persistence
- the existing stub GET metadata endpoints remain intact and separate from the new Phase 1 POST/GET endpoints

## Risks

- handlers can accidentally encode business logic that belongs in the next slice
- placeholder auth extraction can be mistaken for a real security boundary if not clearly limited
- response placeholders can drift from the canonical schema contracts if mapping is not tested

## Validation steps

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `bash scripts/validate.sh`

## Rollback or containment notes

If the handler skeleton shape is wrong, revert the new Phase 1 routes and local placeholder implementations before any UI or integration code starts depending on them.
Do not partially keep auth extraction or repository scaffolding without the matching endpoint tests.

## Open questions

- whether the first read endpoints should include credential detail retrieval immediately or wait until the issuer-flow slice
- whether idempotency-key headers should already be echoed or only stored later when repositories become real
