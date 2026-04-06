# 0002 HDIP Initial Platform Baseline

- Status: accepted
- Date: 2026-04-06
- Owners: repository maintainer

## Context

HDIP needs a serious baseline architecture for portable identity, verifiable credentials, selective disclosure, trust policy, and global operationalization.
The goal is not a chain-only identity app.
The goal is a hybrid identity and trust infrastructure platform with user-held credentials, standards-based interoperability, and enterprise-ready policy and compliance layers.

## Decision

Adopt the following initial platform baseline:

- hybrid decentralized identity model rather than pure on-chain identity
- standards-led issuance and presentation using W3C VC 2.0, DID Core-compatible identifiers, OpenID4VCI, and OpenID4VP
- default credential profile of SD-JWT VC for mainstream issuance and presentation
- advanced privacy profile of VC Data Integrity plus BBS for high-sensitivity selective disclosure
- Bitstring Status List for credential status and revocation handling
- WebAuthn/passkeys for authentication and device-bound access
- native mobile wallet apps on iOS and Android with a shared Rust crypto core
- web product surfaces in Next.js and TypeScript
- backend services primarily in Go with isolated Rust cryptographic logic
- identity and OAuth/OIDC platform support via Ory Kratos and Ory Hydra
- relationship-based authorization via OpenFGA
- CockroachDB as the primary global transactional store
- Redis/Valkey for cache and ephemeral state
- ClickHouse for analytics and risk telemetry
- NATS JetStream for messaging
- Temporal for long-running workflows
- Vault plus cloud KMS/HSM for platform secret and signing control
- Cloudflare at the edge

## Alternatives considered

### Single blockchain-first identity application

Rejected because it would not fit regulatory, UX, and interoperability demands for mainstream issuers and verifiers.

### Wallet-only product

Rejected because it would underbuild verifier, policy, trust, and compliance infrastructure.

### Pure monolithic backend

Rejected because it would blur critical trust boundaries and make future growth harder.

## Security impact

Positive overall, with increased system complexity.
Security-sensitive logic is intentionally isolated in Rust and protected by stronger governance.

## Privacy impact

Positive if implemented correctly.
The baseline explicitly favors selective disclosure, minimized presentations, and regional data-boundary controls.

## Migration / rollback

This ADR establishes a baseline, not a frozen implementation.
Specific framework or deployment choices may be refined later, but changes that hit the ADR trigger list require follow-up decisions.

## Consequences

- The repository will be polyglot.
- Repo structure must preserve clear ownership boundaries.
- Tooling and CI will need staged rollout by area.
- Some implementation choices remain intentionally deferred until foundation work is in place.

## Open questions

- Exact monorepo/workspace orchestration approach
- Local developer orchestration for Go, Rust, Next.js, and mobile toolchains
- Initial hosted environment topology for the first runnable slice

## Related plans, PRs, and issues

- `docs/plans/active/0001-governance-and-foundation-bootstrap.md`
- `docs/architecture/platform-baseline.md`
