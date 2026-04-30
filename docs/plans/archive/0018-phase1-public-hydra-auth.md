# 0018 Phase 1 Public Hydra Auth

## Objective

Implement Hydra-backed public Phase 1 issuer and verifier authentication for the existing reusable-KYC endpoints, without changing public API request/response contracts, trust decision rules, storage schemas, wallet behavior, or proof semantics.

The implementation must replace the current header-only sandbox attribution path for packaged/pilot deployments while preserving a deprecated header mode for local tests and process-run sandbox automation.

## Scope

- Add ADR governance for public issuer/verifier OAuth2 client-credentials auth.
- Add a full threat model artifact for the public auth boundary.
- Add Hydra token introspection based attribution extraction under the existing `authctx` service-edge boundary.
- Wire `issuer-api` and `verifier-api` to use Hydra public auth when configured.
- Keep current header attribution as a deprecated non-production mode.
- Update Compose provisioning to create public issuer/verifier clients and resource-server introspection clients.
- Update quickstart and integration docs to use bearer tokens instead of trusted `X-HDIP-*` headers for packaged local flows.
- Add tests for active token acceptance, inactive/malformed token rejection, missing scope rejection, and config fail-fast behavior.

## Out Of Scope

- End-user holder authentication.
- Wallet flows.
- Passkeys.
- Browser redirects or authorization-code flow.
- API keys.
- Relationship-based authorization.
- Public client self-service provisioning UI.
- Rate limiting or edge WAF configuration.
- Public contract/schema changes.
- Trust-registry trust decision changes.

## Affected Files, Services, Or Packages

- `docs/adr/0012-phase1-public-hydra-auth.md`
- `docs/threat-model/full/0004-phase1-public-hydra-auth.md`
- `packages/go/foundation/authctx/`
- `services/issuer-api/internal/config/`
- `services/issuer-api/internal/httpapi/`
- `services/verifier-api/internal/config/`
- `services/verifier-api/internal/httpapi/`
- `infra/phase1/.env.example`
- `infra/phase1/docker-compose.yml`
- `infra/phase1/hydra/bootstrap-clients.sh`
- `docs/integration/quickstart.md`
- `docs/integration/phase1-verifier-sdk-and-api-guide.md`
- `docs/deployment/phase1-fintech-exchange-cloud-infra.md`
- `docs/runbooks/phase1-sandbox.md` if wording must distinguish header-mode process sandbox from Hydra-mode packaged flows

## Assumptions

- Ory Hydra remains the Phase 1 OAuth2 server.
- Public issuer/verifier API callers use OAuth2 client credentials.
- Hydra introspection is the governed resource-server validation path for this slice.
- The OAuth `client_id` is the Phase 1 organization identifier for public callers.
- The existing accepted Phase 1 scopes remain canonical:
  - `issuer.credentials.issue`
  - `issuer.credentials.read`
  - `issuer.credentials.status.write`
  - `verifier.requests.create`
  - `verifier.results.read`
- Header-based attribution remains available only for local development/tests and must be rejected when `HDIP_ENVIRONMENT=production`.

## Risks

- Misconfigured Hydra introspection could make services appear ready while auth is unusable.
- Public client IDs become organization identifiers, so client provisioning must be controlled.
- Client credentials are high-value secrets and must not appear in logs, fixtures, screenshots, or browser code.
- Keeping header mode for local workflows can be misused if production config validation is weak.
- Hydra outage will fail public write/read auth closed for protected endpoints.

## Validation Steps

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `bash scripts/validate.sh`
- targeted Go tests for `packages/go/foundation/authctx`
- targeted Go tests for issuer/verifier config and handlers
- Compose config validation
- packaged local smoke using Hydra bearer tokens for issuer and verifier calls

## Rollback Or Containment Notes

Rollback by reverting this slice.
If Hydra public auth is misconfigured during a pilot, stop public ingress for issuer/verifier services and keep SQL state intact.
Do not fall back to header attribution in production; use header mode only for local development and tests.

## Open Questions

- Whether later public auth should add audience checks once Hydra audience conventions are locked for HDIP.
- Whether public auth should move from introspection to local JWT validation after token shape, JWKS handling, and audience policy are governed.
- Whether issuer console human delegation should be added through a separate authorization-code or passkey-backed flow.

## Implementation Sequence

1. Add ADR 0012 and full threat model 0004.
2. Add shared Hydra introspection attribution extractor in `authctx`.
3. Add issuer/verifier config for `HDIP_PUBLIC_AUTH_MODE=header|hydra`, Hydra introspection settings, and production fail-fast rules.
4. Wire issuer/verifier handlers to use Hydra auth when configured and include auth provider readiness checks.
5. Add unit tests for extractor, config, and handler auth behavior.
6. Update Compose Hydra client provisioning and service env.
7. Update quickstart/integration/deployment docs.
8. Run validation and packaged smoke.
