# 0017 Phase 1 Fintech Exchange Product Layer

- Status: active
- Date: 2026-04-28
- Owners: repository maintainer

## Objective

Design the first Phase 1 product integration loop for fintech and exchange onboarding without changing runtime code in this slice.

The product target is narrow:

- an issuer operator can create and manage a reusable KYC credential
- a user can carry the Phase 1 credential artifact from issuer to verifier
- a fintech or exchange backend can call the verifier API and receive a deterministic `allow` or `deny`
- the first deployment can run the existing Phase 1 service set with SQL-primary state and Hydra-backed internal trust reads

## Scope

- design the Phase 1 issuer console UI and operator flows
- design the verifier SDK and integration documentation surface
- design the first cloud deployment architecture for a fintech/exchange pilot
- document the current contract constraint that `credentialArtifact` is the verifier submission artifact, while `credentialId` is an optional traceability aid
- add the required deployment-topology ADR
- add the required threat delta for product integration and deployment exposure

## Out of scope

- service, database, SDK, UI, or deployment code changes
- production public issuer/verifier auth implementation
- wallet flows
- selective disclosure, proof verification, or signed credential artifacts
- QR or short-token resolver implementation
- KYC vendor integration
- AI scoring, reputation graphing, chain anchoring, or cross-vertical trust products
- Kubernetes or multi-region deployment topology

## Affected files, services, or packages

- `docs/plans/active/0017-phase1-fintech-exchange-product-layer.md`
- `docs/product/phase1-issuer-console.md`
- `docs/integration/phase1-verifier-sdk-and-api-guide.md`
- `docs/deployment/phase1-fintech-exchange-cloud-infra.md`
- `docs/adr/0011-phase1-fintech-exchange-deployment-topology.md`
- `docs/threat-model/delta/0006-phase1-product-integration-and-deployment.md`
- `docs/repo-structure.md`

Implementation areas for later slices:

- `apps/issuer-console/`
- `apps/developer-portal/`
- `packages/api-client/`
- future SDK package under `packages/`
- `infra/`

## Current state summary

The repo has moved beyond foundation-only scaffolding.
Current accepted docs and workspace files indicate:

- deterministic Phase 1 issuer and verifier logic exists behind the real Phase 1 endpoints
- credential status mutation and idempotency are in scope
- SQL-primary Phase 1 persistence and a `phase1sql` migration/bootstrap lifecycle are governed by ADR 0010
- Hydra client credentials plus introspection are governed for internal verifier-to-trust-registry runtime reads
- public issuer/verifier operator auth remains a service-edge attribution boundary and is not yet a production auth rollout

This plan treats those facts as the current baseline.

## Product flow

1. Issuer operator verifies a user through an internal KYC process outside HDIP runtime scope.
2. Issuer console creates an HDIP reusable KYC credential with normalized claims.
3. HDIP returns a `credentialId`, current status, and opaque Phase 1 `credentialArtifact`.
4. User carries the artifact through the temporary bridge.
5. Exchange or fintech backend calls `POST /v1/verifier/verifications`.
6. HDIP evaluates credential existence, status, expiry, template compatibility, and issuer trust state.
7. Verifier receives `allow` or `deny` and onboards or routes the user to manual KYC.

## Contract constraint

The user-facing "credential ID only" experience is a product goal, but it is not the current trusted Phase 1 verifier contract.

Current accepted contracts require:

- `credentialArtifact` as the verifier submission artifact
- optional `credentialId` for traceability

The first bridge therefore exposes the opaque artifact directly or encodes it into a QR/manual transfer payload.
A short-token or credential-ID resolver is a later governed slice because it changes trust, replay, and privacy properties.

## Assumptions

- The issuer console is an internal KYC operations tool, not a public consumer wallet.
- The verifier SDK runs only server-side in partner fintech/exchange backends.
- The developer portal can host integration docs before a full console exists.
- The first deployment uses one SQL instance, one Hydra instance, and the three Phase 1 services: `issuer-api`, `verifier-api`, and `trust-registry`.
- `trust-registry` is not public in the first deployment.
- `phase1sql migrate up` runs before trust bootstrap and service startup.
- Public edge protection and partner onboarding controls are deployment responsibilities until public issuer/verifier auth is governed in a later slice.

## Risks

- Product copy can imply production cryptographic credentials before signed artifacts are governed.
- A credential-ID-only bridge can accidentally bypass the accepted `credentialArtifact` contract.
- Verifier SDK examples can encourage frontend calls or client-side secret exposure if not explicit.
- Issuer console search can become a broad PII query surface if it is not bounded to operational identifiers.
- First deployment can hide trust-registry or Hydra failures if readiness gates are bypassed.
- One SQL instance is simple but becomes a single operational dependency for the pilot.

## Validation steps

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `bash scripts/validate.sh`

## Rollback or containment notes

This slice is documentation-only.
If review rejects the product-layer direction, remove this plan and the linked design docs before implementing UI, SDK, or deployment assets.
Do not partially implement a credential-ID-only verifier bridge unless a follow-up ADR changes the verifier submission contract.

## Open questions

- Whether the first temporary user bridge should expose raw opaque artifact text, a QR encoding of the same artifact, or a governed short-token resolver.
- Whether issuer console search should initially support only `credentialId` and `subjectReference`, or require a separate search endpoint with stricter audit controls.
- Whether the first verifier SDK should be TypeScript-only or include a Go SDK at the same time.
- Which cloud provider will host the first pilot deployment.

## Exact implementation sequence after this design slice

### Slice 18: Issuer console operational UI

- add issuer console screens for credential creation, credential lookup, status view, and status mutation
- use existing typed client boundaries or extend them for current Phase 1 endpoints
- keep auth placeholder treatment explicit until public/operator auth is governed
- add UI tests for success, loading, error, and status-transition states

### Slice 19: Verifier SDK and developer docs

- add a server-side verifier SDK package or API-client extension for verification create/read
- add developer portal docs for onboarding, cURL, TypeScript usage, errors, idempotency, and webhook-free polling
- add contract tests against canonical schema examples

### Slice 20: Temporary user bridge

- add an issuer-console copy/QR view for the opaque `credentialArtifact`
- document that QR/manual artifact transfer is transitional and not a wallet or proof flow
- defer short-token resolver unless separately governed

### Slice 21: First deployment assets

- add minimal deployment assets for one SQL instance, one Hydra instance, and three services
- document `phase1sql migrate up` and `phase1sql bootstrap trust` ordering
- wire health/readiness and rollback commands

### Slice 22: Pilot readiness check

- run one end-to-end pilot scenario:
  - create credential
  - carry artifact
  - verifier API returns `allow`
  - revoke credential
  - verifier API returns `deny`
  - suspended issuer returns `deny`
