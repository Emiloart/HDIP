# 0012 Phase 1 Public Hydra Auth

- Status: accepted
- Date: 2026-04-29
- Owners: repository maintainer

## Context

HDIP Phase 1 now has a working reusable-KYC issuer/verifier flow, SQL-primary state, Hydra-backed internal verifier-to-trust-registry identity, local deployment packaging, and an external quickstart.

The remaining rollout blocker is controlled public access for issuer and verifier integrations.
The current service handlers derive attribution from trusted `X-HDIP-*` headers.
That was acceptable for local tests, console shells, and process-run sandbox automation, but it is not acceptable for a fintech/exchange pilot where callers cross a public or partner-facing network boundary.

ADR 0008 already requires authenticated and attributable issuer/verifier callers.
ADR 0010 already accepts Hydra client credentials plus introspection for the internal runtime trust-read boundary.
This ADR extends the same OAuth2/Hydra direction to the public Phase 1 issuer and verifier API boundary.

## Decision

### Public auth model

Phase 1 public issuer and verifier API access uses Ory Hydra OAuth2 client credentials.

The public Phase 1 grant type is:

- `client_credentials`

The public Phase 1 resource-server validation model is:

- token introspection against Hydra
- no trusted public `X-HDIP-*` caller headers
- no request-body organization identity
- no API-key authentication system

### Public issuer clients

An issuer client represents an issuer organization allowed to perform KYC issuance operations.

The OAuth `client_id` is the Phase 1 issuer organization identifier used for attribution.
For the sandbox issuer, this means the client ID is the issuer DID already present in trust bootstrap:

- `did:web:issuer.hdip.dev`

Issuer clients may request only issuer scopes.
The canonical Phase 1 issuer scopes remain the accepted ADR 0008 scopes:

- `issuer.credentials.issue`
- `issuer.credentials.read`
- `issuer.credentials.status.write`

This ADR does not introduce parallel `issuer.credentials.write` or `issuer.credentials.status` aliases.

### Public verifier clients

A verifier client represents a fintech/exchange/verifier organization allowed to submit reusable-KYC verification requests.

The OAuth `client_id` is the Phase 1 verifier organization identifier used for attribution.

Verifier clients may request only verifier scopes.
The canonical Phase 1 verifier scopes remain the accepted ADR 0008 scopes:

- `verifier.requests.create`
- `verifier.results.read`

This ADR does not introduce parallel `verifier.verifications.write` or `verifier.verifications.read` aliases.

### Resource-server introspection

`issuer-api` and `verifier-api` must validate public bearer tokens by Hydra introspection before protected handlers execute.

The minimum acceptance checks are:

- `Authorization: Bearer <access-token>` is present
- token introspection succeeds
- token is active
- introspection response includes a non-empty `client_id`
- required action scope is present

If any check fails, the service must fail closed.
Hydra introspection outages must not allow access.

### Attribution mapping

Hydra introspection output maps into the existing `authctx.Attribution` shape:

- `PrincipalID`: introspected `client_id`
- `OrganizationID`: introspected `client_id`
- `ActorType`: `issuer_operator` for issuer API, `verifier_integrator` for verifier API
- `Scopes`: introspected `scope`
- `AuthenticationReference`: introspected token identifier when available, otherwise the Hydra client identity

The Phase 1 organization binding therefore comes from the Hydra client identity, not request headers or request bodies.

### Header mode deprecation

Header attribution remains available only as a local development and test mode.

The service auth mode is configured explicitly:

- `HDIP_PUBLIC_AUTH_MODE=hydra`
- `HDIP_PUBLIC_AUTH_MODE=header`

`header` mode is deprecated for pilot use and must be rejected when the deployment environment is production.

Production deployments must use:

- `HDIP_ENVIRONMENT=production`
- `HDIP_PUBLIC_AUTH_MODE=hydra`

### Token endpoint client authentication

The local Phase 1 Compose sandbox may provision public clients with `client_secret_post` to support issuer DID client identifiers that contain `:`.
This is a local packaging concern only.
Resource-server introspection clients still authenticate to Hydra introspection using their configured confidential client credentials.

### Deferred auth features

This ADR does not add:

- end-user holder auth
- wallet auth
- authorization-code flow
- passkey flows
- organization admin delegation
- API keys
- relationship-based authorization
- audience checks beyond introspection unless separately governed
- local JWT/JWKS verification

## Alternatives Considered

### API keys

Rejected because API keys would introduce a parallel public auth system with weaker revocation semantics, poorer alignment with existing Hydra direction, and weaker audit attribution.

### Keep trusted headers for pilots

Rejected for any public or partner-facing pilot because trusted headers are only safe behind a separately controlled edge-auth system.
No such public edge-auth system is governed in Phase 1.

### Hydra authorization-code flow

Rejected for this slice because Phase 1 public verifier integration is server-to-server and does not require browser redirects or end-user login.

### Local JWT/JWKS validation

Rejected for this slice because token shape, audience policy, JWKS caching, and rotation behavior are not yet governed for public API callers.
Hydra introspection is the narrower and already-used resource-server validation pattern.

### Rename existing Phase 1 scopes

Rejected because ADR 0008 and existing service handlers already define more precise Phase 1 scopes.
Renaming them would create unnecessary migration and compatibility drift.

## Security Impact

Positive.
This replaces trusted caller headers for packaged/pilot public paths with Hydra-issued bearer tokens, active-token validation, revocable client credentials, and explicit scope checks.

New risk is concentrated in Hydra availability, client provisioning, and client secret handling.
Services must fail closed on missing, inactive, malformed, or unavailable token validation.

## Privacy Impact

Neutral to positive.
No new KYC data is collected.
Attribution records continue to store organization and client identity rather than raw tokens.
Client credentials and bearer tokens must not be logged, stored in fixtures, or exposed to browser/mobile clients.

## Migration / Rollback

Existing local tests and process-run sandbox automation may continue using header mode.
Packaged and production-like deployment should move to Hydra mode.

Rollback by returning local deployment config to `HDIP_PUBLIC_AUTH_MODE=header` only outside production.
Do not use header mode as a production rollback.
If Hydra public auth fails in a pilot, stop public ingress and repair Hydra/client provisioning.

## Consequences

- `issuer-api` and `verifier-api` gain Hydra introspection configuration.
- Public quickstart and Compose provisioning must create public issuer/verifier clients.
- Public quickstart uses bearer tokens instead of `X-HDIP-*` headers.
- Token introspection becomes a readiness dependency when Hydra auth mode is enabled.
- Public client provisioning remains operator-driven for Phase 1 and may use the pilot provisioning script.

## Open Questions

- Whether later production auth should require audience checks.
- Whether public token validation should move to local JWT verification after token shape and key rotation are governed.
- Whether issuer console human operators should use delegated user auth in a later slice.

## Related Plans, PRs, And Issues

- `docs/plans/archive/0018-phase1-public-hydra-auth.md`
- `docs/plans/active/0019-phase1-pilot-readiness-completion.md`
- `docs/adr/0008-phase1-auth-and-attribution-boundary.md`
- `docs/adr/0010-phase1-internal-trust-service-identity-and-sql-lifecycle.md`
- `docs/adr/0011-phase1-fintech-exchange-deployment-topology.md`
- `docs/threat-model/full/0004-phase1-public-hydra-auth.md`
