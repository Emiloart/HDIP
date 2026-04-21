# 0003 Phase 1 KYC Issuance Verification And Auth Threat Model

- Status: accepted
- Date: 2026-04-20
- Owners: repository maintainer

## Change summary

This threat model covers the Phase 1 transition from a read-only stub issuer/verifier flow to a real reusable KYC credential and verifier API.
The Phase 1 boundary adds authenticated issuer writes, authenticated verifier writes, persistent credential and verification state, audit records, credential status handling, and minimum trust-registry participation.

## Assets

- issuer operator identities and credentials
- verifier integrator identities and credentials
- issuer trust records and verification-key references
- opaque Phase 1 credential artifacts
- credential records and credential status
- verification request records
- verification result records
- audit records
- schema and typed-client contract definitions that drive privileged flows

## Trust boundaries

- issuer operator to `issuer-api`
- verifier integrator to `verifier-api`
- `issuer-api` to credential persistence
- `verifier-api` to verification persistence
- `issuer-api` and `verifier-api` to `trust-registry`
- service edge auth-context extraction boundary
- audit append boundary
- typed client to backend API boundary
- stub-era metadata endpoints that continue to coexist beside real Phase 1 endpoints

## Attacker classes

- external attackers attempting unauthorized issuance
- attackers replaying or forging verification requests
- malicious or overreaching verifiers
- compromised issuer operators
- insiders tampering with persistence or audit records
- attackers exploiting stub-era assumptions during migration
- compromised or stale trust-registry operators or data feeds

## Entry points and privileged actions

- `POST /v1/issuer/credentials`
- issuer credential detail reads
- issuer revoke and supersede actions
- `POST /v1/verifier/verifications`
- verifier result reads
- trust-registry updates to issuer trust state or verification-key references
- service-edge auth token or credential validation

## Abuse and misuse cases

- unauthorized issuance of reusable KYC credentials
- replayed or forged verification submissions intended to obtain repeated decisions or pollute audit state
- verifiers requesting or storing more end-user data than the Phase 1 contract requires
- operator account takeover leading to fraudulent issuance or status changes
- tampering with audit or persistence records to hide misuse
- treating stub-era GET endpoints as if they were authoritative production verification APIs
- misusing credential status to enumerate or correlate credentials across verifiers
- trusting issuer identifiers without consulting the trust-registry boundary

## Privacy harms

- persistence leakage of sensitive KYC claims or opaque Phase 1 credential artifacts
- cross-verifier correlation if `subjectReference` is not kept opaque and bounded
- verifier over-collection beyond the Phase 1 reusable KYC contract
- raw credential or attribution data leaking into logs, audit payloads, or debug tooling
- audit records becoming a shadow copy of sensitive credential data

## Mitigations

- require authenticated issuer operator identity for issuance and status-changing actions
- require authenticated verifier integrator identity for verification writes and sensitive reads
- derive authoritative issuer and verifier organization identity from auth context, not request bodies
- bind issuance and verification writes to request identifiers and idempotency keys where supported
- keep Phase 1 credential claims normalized and bounded; exclude raw KYC evidence from runtime records
- keep deterministic Phase 1 artifacts opaque and non-cryptographic until a later signing ADR lands
- keep verification request persistence to digests and bounded metadata rather than duplicating full credentials by default
- make audit records append-only and reference sensitive artifacts by identifiers or digests
- consult `trust-registry` during verification for issuer trust state and verification-key references
- return verifier decision `deny` for suspended or otherwise non-active issuers in deterministic Phase 1
- keep status handling internal to issuer and verifier flows rather than exposing a broad anonymous status lookup in Phase 1
- keep stub endpoints clearly separated from the real Phase 1 write path and do not reinterpret them as production verification flows

## Residual risks

- operator credentials remain high-value until richer auth, rotation, and step-up controls land
- Phase 1 still has more linkability risk than later holder-controlled or selective-disclosure flows
- opaque Phase 1 artifacts do not provide cryptographic authenticity until a later signing model is approved
- synchronous verifier evaluation can still be abused for denial-of-service without future rate or risk controls
- trust-registry remains an HDIP-controlled dependency rather than a federated trust network in Phase 1
- if implementation cuts corners on idempotency or audit immutability, replay and repudiation risk will remain elevated

## Validation impact

The first real Phase 1 code slice must add:

- schema validation for issuance, credential status, and verification contracts
- TypeScript parity coverage for the new Phase 1 contracts
- service tests for authenticated issuance and verification paths
- tests for malformed or missing auth context
- tests for replay or duplicate write handling
- tests that status and trust-registry lookups affect verifier decisions deterministically, including `deny` for suspended or non-active issuers
- tests that logs and audit records do not contain raw sensitive credential payloads

## Related ADRs, plans, PRs, and issues

- `docs/plans/active/0006-phase1-kyc-credential-and-verifier-api.md`
- `docs/adr/0006-phase1-credential-and-issuance-boundary.md`
- `docs/adr/0007-phase1-state-and-persistence-model.md`
- `docs/adr/0008-phase1-auth-and-attribution-boundary.md`
- `docs/adr/0009-phase1-opaque-artifact-and-suspended-issuer-policy.md`
