# 0006 Phase 1 Product Integration And Deployment

- Status: accepted
- Date: 2026-04-28
- Owners: repository maintainer

## Change summary

This delta covers the design of the first fintech/exchange product integration loop:

- issuer console operational workflows
- verifier SDK and integration documentation
- temporary user bridge using the opaque Phase 1 `credentialArtifact`
- first pilot deployment topology with one SQL instance, one Hydra instance, and the three Phase 1 services

No runtime code is changed in this slice.

Current note: public issuer/verifier auth was later implemented under ADR 0012 and threat model 0004.
Pilot operations and partner provisioning are tracked in threat delta 0007.

## New or changed entry points

Designed but not implemented:

- issuer console credential creation flow
- issuer console credential lookup flow
- issuer console status mutation flow
- verifier SDK `verifyCredential`
- verifier SDK `getVerification`
- developer portal integration docs
- protected first-pilot deployment ingress

Existing runtime endpoints remain:

- `POST /v1/issuer/credentials`
- `GET /v1/issuer/credentials/{credentialId}`
- `POST /v1/issuer/credentials/{credentialId}/status`
- `POST /v1/verifier/verifications`
- `GET /v1/verifier/verifications/{verificationId}`

## New or changed privileged actions

Designed but not implemented:

- issuer operator creates credentials through a console
- issuer operator views credential details through a console
- issuer operator revokes or supersedes credentials through a console
- verifier integrator calls HDIP from a partner backend through an SDK
- deployment operator runs `phase1sql migrate up`
- deployment operator runs `phase1sql bootstrap trust`

## Threat delta

- issuer console can become a high-value operator surface for fraudulent issuance
- issuer search can become a PII enumeration surface if query capabilities are too broad
- verifier SDK examples can leak partner credentials if they encourage browser/mobile execution
- temporary artifact bridge can be mistaken for a signed credential or wallet proof
- credential-ID-only UX can accidentally bypass the accepted `credentialArtifact` submission boundary
- first deployment can expose `trust-registry`, Hydra admin, introspection, or SQL paths if network boundaries are too loose
- SQL backups and logs can become secondary sensitive-data stores
- partner verifiers can store opaque artifacts longer than needed for onboarding

## Privacy delta

- issuer console increases the number of operator-facing views of normalized KYC claims
- verifier SDK docs can influence partner retention behavior
- temporary artifact transfer adds a user-visible sensitive artifact handling step
- first deployment introduces backups and operational logs that must be minimized

## Mitigations

- keep issuer console protected-access only in the first deployment
- require issuer console workflows to use existing issuer scopes and audit requirements
- keep issuer search bounded to operational identifiers until stricter search audit is governed
- publish verifier SDK examples as server-side only
- require idempotency keys in verifier SDK write examples
- make docs explicit that `credentialArtifact` is opaque, non-cryptographic, and not a signed credential
- keep credential-ID-only UX deferred until a resolver or token bridge is governed
- keep `trust-registry`, Hydra admin/introspection, and SQL on private network paths
- run SQL migration and trust bootstrap explicitly before service startup
- prohibit logs of opaque artifacts, normalized KYC claims, bearer tokens, and raw request bodies
- keep partner retention guidance narrow: store decisions and audit identifiers, not artifacts

## Residual risks

- public issuer/verifier auth is implemented; controlled partner provisioning and edge operations remain pilot-readiness work
- first deployment has a single SQL operational dependency
- temporary artifact transfer has more user-handling risk than a later wallet flow
- partner systems can still mishandle result retention outside HDIP unless contracts and onboarding reviews catch it

## Validation impact

Future implementation slices must add:

- issuer console tests for create, lookup, status view, status mutation, and error states
- SDK tests for successful verification, `deny` outcomes, idempotency, typed errors, and malformed responses
- deployment validation for SQL migration/bootstrap ordering
- readiness checks for Hydra token acquisition and introspection
- ingress checks proving `trust-registry`, Hydra admin/introspection, and SQL are not publicly reachable
- log checks proving artifacts, claims, tokens, and raw request bodies are not emitted

## Related ADRs, plans, PRs, and issues

- `docs/plans/archive/0017-phase1-fintech-exchange-product-layer.md`
- `docs/threat-model/delta/0007-phase1-pilot-operations.md`
- `docs/adr/0011-phase1-fintech-exchange-deployment-topology.md`
- `docs/adr/0010-phase1-internal-trust-service-identity-and-sql-lifecycle.md`
- `docs/threat-model/full/0003-phase1-kyc-issuance-verification-and-auth.md`
- `docs/product/phase1-issuer-console.md`
- `docs/integration/phase1-verifier-sdk-and-api-guide.md`
- `docs/deployment/phase1-fintech-exchange-cloud-infra.md`
