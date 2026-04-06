# Repo Structure

## Purpose

This document is the stable map of repository boundaries, intended ownership zones, and forbidden dependency directions.
If the repository structure changes, this file must be updated in the same change.

## Current state

The repository is in foundation bootstrap.
Most directories are structural placeholders that reserve boundaries before product code lands.

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

## Planned web surfaces

- `apps/public-site/`
- `apps/issuer-console/`
- `apps/verifier-console/`
- `apps/org-admin-console/`
- `apps/developer-portal/`

These are expected to use Next.js and TypeScript.
They must not become the system of record for wallet custody, cryptographic verification, or backend policy decisions.

## Planned mobile surfaces

- `mobile/wallet-ios/`
- `mobile/wallet-android/`

These are expected to be native applications with a shared Rust credential and crypto core.

## Planned backend service boundaries

- `services/identity/`
- `services/credential-issuance/`
- `services/presentation/`
- `services/verification/`
- `services/trust-registry/`
- `services/identifier-resolution/`
- `services/credential-status/`
- `services/consent-delegation/`
- `services/recovery/`
- `services/audit-compliance/`
- `services/notification/`
- `services/developer-platform/`
- `services/risk-fraud/`

Services are expected to be implemented primarily in Go with explicit contracts and isolated trust boundaries.

## Planned security-critical core

- `crates/hdip-crypto-core/`

This area exists for Rust implementations of key handling, proof generation, credential parsing, verification primitives, and other security-sensitive logic that should remain isolated from transport and UI glue.

## Dependency directions

Allowed high-level directions:

- `apps/` -> `packages/`, external service APIs
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

## Local AGENTS guidance

Area-specific `AGENTS.md` files may refine rules in `apps/`, `mobile/`, `services/`, and `crates/`.
Those local files may add stricter constraints but may not weaken root security, privacy, traceability, or validation requirements.

## Bootstrap note

If a directory exists only as a placeholder, contributors must still preserve its intended boundary.
Do not collapse reserved service or product areas together for convenience without an ADR.
