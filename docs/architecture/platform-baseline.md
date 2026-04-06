# Platform Baseline

## Goal

HDIP is a hybrid identity and trust infrastructure platform.
It is not a chain-only identity app and not just a wallet.
The platform combines user-held credentials, standards-based issuance and presentation, trust and policy controls, and region-aware operations.

## Architecture shape

The initial reference shape is:

- user plane: wallets, passkeys, consent UX, device binding
- credential plane: issuance, schemas, status, revocation, selective disclosure support
- proof plane: presentation requests, proof generation, proof verification
- trust plane: trust registry, issuer trust, verifier policy, reputation and risk inputs
- compliance plane: audit, policy, retention, jurisdiction controls
- developer plane: SDKs, verifier APIs, issuer APIs, docs, sandboxing
- operations plane: edge routing, telemetry, workflows, messaging, secrets, service identity

## Technology baseline

### Standards and identity

- W3C VC 2.0
- DID Core-compatible identifiers
- OpenID4VCI
- OpenID4VP
- WebAuthn and passkeys
- VC Data Integrity
- Bitstring Status List
- SD-JWT VC as the mainstream default
- BBS/Data Integrity for advanced privacy-sensitive credentials

### Client surfaces

- native wallet on iOS with SwiftUI
- native wallet on Android with Kotlin and Jetpack Compose
- shared Rust core for keys, credentials, proofs, and local vault operations
- web surfaces in Next.js and TypeScript

### Backend

- Go services for APIs and orchestration
- Rust crypto core isolated from web and transport glue
- gRPC internally and HTTP/JSON externally where appropriate

### Platform services

- Ory Kratos for identity flows
- Ory Hydra for OAuth2/OIDC server capabilities
- OpenFGA for authorization
- CockroachDB for global transactional state
- Redis/Valkey for cache and ephemeral state
- ClickHouse for analytics and risk telemetry
- NATS JetStream for messaging
- Temporal for durable workflows
- Vault plus cloud KMS/HSM for secret and signing control
- Cloudflare for edge routing and filtering

## Initial repository mapping

- `apps/` for web surfaces
- `mobile/` for wallet applications
- `services/` for Go services
- `crates/` for Rust core
- `packages/` for future shared libraries and SDK surfaces
- `infra/` for topology and deployment assets
- `schemas/` for contracts and credential definitions

## Planned build sequence

### Phase 0

Governance spine, architecture baseline, repo structure, and validation bootstrap.

### Phase 1

Foundation implementation for:

- repo workspaces and toolchains
- Rust crypto core skeleton
- one or two core Go services
- public site and developer portal skeleton
- initial schemas and contract directories

### Phase 2

First runnable verification slice:

- passkey-backed account entry
- issuer and verifier control surfaces
- issuance and verification service skeletons
- status service
- trust registry service

### Phase 3

Wallet and proof flows:

- mobile wallet scaffolds
- selective disclosure support
- presentation flow orchestration
- recovery model implementation

## Explicit deferrals

- final monorepo orchestration choice
- local devcontainer or Nix setup
- production cluster manifests
- optional chain anchoring
- advanced reputation graph implementation
