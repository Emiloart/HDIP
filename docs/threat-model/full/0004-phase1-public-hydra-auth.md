# 0004 Phase 1 Public Hydra Auth Threat Model

- Status: accepted
- Date: 2026-04-29
- Owners: repository maintainer

## Change Summary

This threat model covers the Phase 1 public issuer/verifier authentication transition from trusted header attribution to Hydra OAuth2 client-credentials bearer tokens validated by token introspection.

This change does not alter Phase 1 API contracts, credential artifact semantics, persistence schema, trust-registry decision rules, wallet behavior, or proof verification behavior.

## Assets

- issuer public OAuth clients and client secrets
- verifier public OAuth clients and client secrets
- issuer/verifier bearer access tokens
- Hydra introspection client credentials for `issuer-api`
- Hydra introspection client credentials for `verifier-api`
- issuer/verifier attribution records
- audit records containing client-derived attribution
- existing KYC credential records, opaque artifacts, and verification results protected by the auth boundary

## Trust Boundaries

- public issuer client to Hydra token endpoint
- public verifier client to Hydra token endpoint
- public issuer client to `issuer-api`
- public verifier client to `verifier-api`
- `issuer-api` to Hydra introspection
- `verifier-api` to Hydra introspection
- service-edge auth context extraction boundary
- existing `verifier-api` to `trust-registry` internal Hydra trust-runtime boundary
- existing service-to-SQL persistence boundaries

## Attacker Classes

- external attackers attempting unauthorized issuance
- external attackers attempting unauthorized verification
- compromised issuer client credentials
- compromised verifier client credentials
- malicious verifier attempting over-collection or repeated correlation
- attackers replaying stolen bearer tokens
- insiders misprovisioning OAuth clients or scopes
- attackers attempting to smuggle trusted `X-HDIP-*` headers in Hydra mode
- attackers attempting denial-of-service through token introspection load

## Entry Points And Privileged Actions

- Hydra token endpoint for public issuer/verifier clients
- Hydra introspection endpoint for issuer/verifier resource servers
- `POST /v1/issuer/credentials`
- `GET /v1/issuer/credentials/{credentialId}`
- `POST /v1/issuer/credentials/{credentialId}/status`
- `POST /v1/verifier/verifications`
- `GET /v1/verifier/verifications/{verificationId}`

## Abuse And Misuse Cases

- issuing credentials with a stolen or over-scoped issuer client
- verifying credentials with a stolen or over-scoped verifier client
- requesting issuer scopes on verifier clients or verifier scopes on issuer clients
- accepting caller identity from headers while Hydra mode is enabled
- treating inactive or failed introspection responses as authenticated
- caching token validation past token expiry without a governed cache policy
- logging bearer tokens, client secrets, opaque artifacts, or KYC claims
- using header mode in production as a shortcut
- provisioning an issuer client whose `client_id` does not match the governed issuer organization identifier

## Failure Modes

- Hydra token endpoint outage prevents partners from obtaining tokens
- Hydra introspection outage causes protected issuer/verifier endpoints to fail closed
- introspection returns inactive token and services reject the request
- introspection returns missing `client_id` and services reject the request
- public client has missing scope and handlers reject the action with insufficient scope
- production service configured with header mode fails startup
- local header-mode sandbox continues to work only outside production

## Privacy Harms

- bearer tokens or client secrets leaking into logs or support channels
- opaque artifacts being logged by partner integrations while debugging auth failures
- verifier organization identifiers being used for cross-context correlation beyond the Phase 1 audit need
- audit records becoming broader than necessary if raw token claims are stored

## Mitigations

- use Hydra client credentials for public issuer/verifier clients
- validate bearer tokens through Hydra introspection at the service edge
- fail closed on missing bearer token, inactive token, missing client identity, or introspection error
- derive `PrincipalID` and `OrganizationID` from introspected `client_id`
- preserve existing Phase 1 scope checks at handlers
- reject production startup when `HDIP_PUBLIC_AUTH_MODE=header`
- keep header attribution only for local development and tests
- do not log request bodies, bearer tokens, client secrets, opaque artifacts, or raw KYC claims
- store only bounded auth references in audit attribution, not raw tokens
- provision issuer and verifier public clients with non-overlapping scopes
- keep Hydra admin/introspection paths private or localhost-bound in local packaging

## Residual Risks

- client credentials remain high-value secrets until richer provisioning, rotation, and partner-management tooling exist
- no rate limiting is added in this slice, so edge/API gateway controls remain necessary before public exposure
- no audience check is governed yet
- no self-service client provisioning or approval workflow exists
- token introspection adds synchronous dependency latency and availability risk

## Validation Impact

This slice must add or preserve tests for:

- Hydra introspection extractor accepts active bearer tokens
- extractor rejects missing, inactive, malformed, and unavailable introspection cases
- issuer handler accepts Hydra-attributed issuer tokens
- verifier handler accepts Hydra-attributed verifier tokens
- missing scopes still fail with `insufficient_scope`
- production header mode fails config validation
- Compose local stack provisions public clients and uses bearer tokens in quickstart smoke
- governance, secret scan, and full validation pass

## Related ADRs, Plans, PRs, And Issues

- `docs/plans/active/0018-phase1-public-hydra-auth.md`
- `docs/adr/0012-phase1-public-hydra-auth.md`
- `docs/adr/0008-phase1-auth-and-attribution-boundary.md`
- `docs/adr/0010-phase1-internal-trust-service-identity-and-sql-lifecycle.md`
- `docs/threat-model/full/0003-phase1-kyc-issuance-verification-and-auth.md`
