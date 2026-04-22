# 0003 Phase 1 KYC Issuance Verification And Auth Threat Model

- Status: accepted
- Date: 2026-04-22
- Owners: repository maintainer

## Change summary

This threat model covers the Phase 1 transition from a read-only stub issuer/verifier flow to a real reusable KYC credential and verifier API.
The Phase 1 boundary now adds authenticated issuer writes, authenticated verifier writes, Cockroach-compatible relational persistence for credential and verification state, append-only audit records, issuer-authenticated credential status mutation, reservation-state idempotency for write paths, trust-registry-owned runtime reads through an explicit verifier trust-read adapter, trust-registry-owned runtime trust bootstrap, Hydra client-credentials plus introspection for internal trust runtime identity, and an explicit shared `phase1sql` migration/bootstrap CLI with versioned SQL assets on the primary SQL path.

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
- `issuer-api` to the relational Phase 1 persistence boundary
- `verifier-api` to the relational Phase 1 persistence boundary
- `trust-registry` to the relational issuer-trust persistence boundary it owns for runtime reads
- verifier trust-read client to the `trust-registry` internal runtime-read boundary
- `trust-registry` to Hydra token introspection for internal runtime-read authorization
- operational migration and bootstrap boundary for the primary `phase1sql` lifecycle
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
- `POST /v1/issuer/credentials/{credentialId}/status` for issuer revoke and supersede actions
- `POST /v1/verifier/verifications`
- verifier result reads
- trust-registry updates to issuer trust state or verification-key references
- trust-registry bootstrap or apply flow for deterministic Phase 1 issuer trust state
- service-edge auth token or credential validation
- Hydra client-credentials token acquisition for verifier runtime trust reads
- Hydra token introspection for trust-registry runtime trust reads
- explicit primary SQL migration and trust bootstrap execution for Phase 1 relational state

## Abuse and misuse cases

- unauthorized issuance of reusable KYC credentials
- replayed or forged verification submissions intended to obtain repeated decisions or pollute audit state
- conflicting reuse of an idempotency key to smuggle a second write under the appearance of a retry
- verifiers requesting or storing more end-user data than the Phase 1 contract requires
- operator account takeover leading to fraudulent issuance or status changes
- tampering with audit or persistence records to hide misuse
- treating stub-era GET endpoints as if they were authoritative production verification APIs
- misusing credential status to enumerate or correlate credentials across verifiers
- trusting issuer identifiers without consulting the trust-registry boundary

## Failure modes

- Hydra introspection outage or unreachable dependency causes `trust-registry` internal runtime trust reads to fail closed
- introspection returning an inactive token causes `trust-registry` to reject the runtime trust read
- introspection returning the wrong service client identity causes `trust-registry` to reject the runtime trust read
- missing scope `trust.runtime.read` causes `trust-registry` to reject the runtime trust read
- primary SQL schema not migrated causes the primary SQL path to fail startup or readiness rather than silently mutating schema during normal service boot

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
- persist idempotency records as bounded caller-and-operation-scoped request fingerprints plus bounded response snapshots rather than raw request payload copies
- reject conflicting reuse of a caller-bound idempotency key with a different request fingerprint and audit the conflict
- keep Phase 1 credential claims normalized and bounded; exclude raw KYC evidence from runtime records
- keep deterministic Phase 1 artifacts opaque and non-cryptographic until a later signing ADR lands
- keep verification request persistence to digests and bounded metadata rather than duplicating full credentials by default
- make audit records append-only and reference sensitive artifacts by identifiers or digests
- consult the trust-registry-owned runtime read boundary through an explicit verifier trust-read adapter for issuer trust state and verification-key references rather than a generic issuer-record path or seeded verifier-local placeholders
- make trust-registry the only service that bootstraps or mutates runtime issuer trust state for deterministic Phase 1
- replace the transitional static trust bearer token with Hydra client-credentials access tokens for verifier runtime trust reads
- require trust-registry to validate runtime trust-read access tokens through Hydra introspection and fail closed on inactive tokens, wrong client identity, or missing scope
- move the primary SQL path to an explicit versioned `phase1sql` migration and trust-bootstrap lifecycle instead of relying on startup schema mutation as the primary control point
- reserve caller-bound idempotency keys before write-side effects so overlapping same-key writes fail closed rather than silently duplicating state
- make issuer-authenticated status transitions update persisted credential state before later issuer or verifier reads
- return verifier decision `deny` for suspended or otherwise non-active issuers in deterministic Phase 1
- keep status handling internal to issuer and verifier flows rather than exposing a broad anonymous status lookup in Phase 1
- keep stub endpoints clearly separated from the real Phase 1 write path and do not reinterpret them as production verification flows

## Residual risks

- operator credentials remain high-value until richer auth, rotation, and step-up controls land
- Phase 1 still has more linkability risk than later holder-controlled or selective-disclosure flows
- opaque Phase 1 artifacts do not provide cryptographic authenticity until a later signing model is approved
- synchronous verifier evaluation can still be abused for denial-of-service without future rate or risk controls
- trust-registry remains an HDIP-controlled dependency rather than a federated trust network in Phase 1
- Hydra introspection introduces a new internal dependency whose availability and latency affect trust-runtime-read success
- transitional JSON state fallback still exists for compatibility and tests, so local misuse of fallback configuration could bypass the primary relational path
- if implementation cuts corners on idempotency conflict handling or audit immutability, replay and repudiation risk will remain elevated
- operational mis-ordering of `phase1sql migrate up` and `phase1sql bootstrap trust` can still block startup until deployment automation and fallback retirement are tightened

## Validation impact

Phase 1 code must maintain:

- schema validation for issuance, credential status, and verification contracts
- TypeScript parity coverage for the new Phase 1 contracts
- service tests for authenticated issuance and verification paths
- tests for malformed or missing auth context
- tests for replay or duplicate write handling
- tests that replayed writes return prior stored results and that conflicting idempotency-key reuse fails cleanly
- tests that overlapping same-key writes fail with explicit reservation or in-flight outcomes
- tests that status and trust-registry lookups affect verifier decisions deterministically, including `deny` for suspended or non-active issuers
- tests that issuer status mutation updates the persisted state seen by later issuer and verifier reads
- tests that separate repository or runtime instances observe the same persisted credential, status, trust, and idempotency state
- tests that trust-registry-owned runtime reads determine verifier trust outcomes through the explicit trust adapter
- tests that trust-registry bootstrap or update actions persist issuer trust state on the primary path with bounded append-only audit records
- tests that verifier runtime trust reads fail closed when Hydra introspection reports inactive tokens, wrong client identity, or missing scope
- tests that the explicit `phase1sql` migration and trust bootstrap lifecycle initializes the current approved Phase 1 SQL schema before service startup
- tests that logs and audit records do not contain raw sensitive credential payloads

## Related ADRs, plans, PRs, and issues

- `docs/plans/active/0006-phase1-kyc-credential-and-verifier-api.md`
- `docs/adr/0006-phase1-credential-and-issuance-boundary.md`
- `docs/adr/0007-phase1-state-and-persistence-model.md`
- `docs/adr/0008-phase1-auth-and-attribution-boundary.md`
- `docs/adr/0009-phase1-opaque-artifact-and-suspended-issuer-policy.md`
- `docs/adr/0010-phase1-internal-trust-service-identity-and-sql-lifecycle.md`
- `docs/plans/archive/0013-phase1-hydra-internal-trust-auth-and-phase1sql-lifecycle.md`
- `docs/plans/active/0014-phase1-sql-primary-hardening-and-fallback-retirement.md`
