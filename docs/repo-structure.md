# Repo Structure

## Purpose

This document is the stable map of repository boundaries, intended ownership zones, and forbidden dependency directions.
If the repository structure changes, this file must be updated in the same change.

## Current state

The repository is in foundation scaffolding.
Governance is in place, the first executable baseline is landed, and the issuer/verifier surfaces now include deterministic stub flow endpoints built on the foundation contracts.

## Ownership model

Until formal teams exist, the repository maintainer owns all areas.
The intended long-term responsibility split is:

- `apps/` for web product surfaces
- `mobile/` for holder wallet applications
- `services/` for backend services
- `crates/` for security-critical Rust components
- `infra/` for deployment, network, and environment definitions
- `schemas/` for contract artifacts
- `docs/` for governance and architecture records
- `scripts/` for repo automation and quality gates

## Top-level directories

- `apps/`: public site and operator-facing web applications
- `mobile/`: native wallet applications
- `services/`: backend services and control-plane services
- `crates/`: Rust security- and crypto-critical components
- `packages/`: future shared libraries and SDK packages
- `infra/`: infrastructure topology, environment definitions, and deployment assets
- `schemas/`: credential schemas, API schemas, and event contract artifacts
- `docs/`: governance, plans, ADRs, threat models, standards, privacy, and architecture docs
- `scripts/`: repo-local automation used by contributors, CI, and Codex hooks

## Documentation subdirectories

- `docs/adr/`: accepted, proposed, superseded, and rejected architecture decisions
- `docs/architecture/`: broad platform architecture notes
- `docs/deployment/`: environment and deployment topology designs
- `docs/integration/`: external developer and SDK integration designs
- `docs/plans/`: active and archived execution plans
- `docs/privacy/`: privacy constraints and data-handling rules
- `docs/product/`: product-surface flows and UX requirements
- `docs/standards/`: standards registry and interoperability constraints
- `docs/threat-model/`: full threat models and threat deltas

## Web surfaces

Foundation slice:

- `apps/issuer-console/`
- `apps/verifier-console/`

Reserved for later slices:

- `apps/public-site/`
- `apps/org-admin-console/`
- `apps/developer-portal/`

These are expected to use Next.js and TypeScript.
They must not become the system of record for wallet custody, cryptographic verification, or backend policy decisions.

## Planned mobile surfaces

- `mobile/wallet-ios/`
- `mobile/wallet-android/`

These are expected to be native applications with a shared Rust credential and crypto core.

## Backend service boundaries

Foundation slice:

- `services/issuer-api/`
- `services/verifier-api/`
- `services/trust-registry/`

Reserved for later slices:

- `services/identity/`
- `services/credential-issuance/`
- `services/presentation/`
- `services/verification/`
- `services/identifier-resolution/`
- `services/credential-status/`
- `services/consent-delegation/`
- `services/recovery/`
- `services/audit-compliance/`
- `services/notification/`
- `services/developer-platform/`
- `services/risk-fraud/`

Services are expected to be implemented primarily in Go with explicit contracts and isolated trust boundaries.

## Security-critical core

Foundation slice:

- `crates/crypto-core/`
- `crates/identity-core/`

Deferred unless justified by a later ADR:

- `crates/policy-core/`

This area exists for Rust implementations of key handling, proof generation, credential parsing, verification primitives, and other security-sensitive logic that should remain isolated from transport and UI glue.

## Dependency directions

Allowed high-level directions:

- `apps/` -> `packages/`, `schemas/`, external service APIs
- `mobile/` -> `crates/` via native bindings, backend APIs
- `services/` -> `crates/`, `schemas/`, selected `packages/` only where language/tooling allows
- `infra/` -> may reference service names and deployment artifacts, but not own product behavior
- `docs/` -> may describe all areas, but does not define runtime behavior by itself

Forbidden directions:

- `apps/` must not own or duplicate cryptographic verification logic that belongs in `crates/` or backend services.
- `apps/` must not embed privileged backend business rules.
- `mobile/` must not become a hidden custodial backend.
- `crates/` must not depend on web or mobile UI layers.
- `crates/` must not perform uncontrolled network I/O.
- `services/` must not depend on app directories.
- `infra/` must not silently redefine trust logic that belongs in services, policies, or standards docs.
- `schemas/` must not contain service-specific hidden behavior.

## Shared package ownership

Foundation slice shared packages:

- `packages/ui/` for React UI primitives and layout shell pieces
- `packages/api-client/` for typed frontend API boundaries and error helpers
- `packages/config-typescript/` for shared TypeScript configuration

Do not put Go or Rust business logic in `packages/`.

## Local AGENTS guidance

Area-specific `AGENTS.md` files may refine rules in `apps/`, `mobile/`, `services/`, and `crates/`.
Those local files may add stricter constraints but may not weaken root security, privacy, traceability, or validation requirements.

## Bootstrap note

If a directory exists only as a placeholder, contributors must still preserve its intended boundary.
Do not collapse reserved service or product areas together for convenience without an ADR.
