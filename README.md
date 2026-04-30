# HDIP

HDIP is a hybrid decentralized identity passport and trust infrastructure platform.
Users hold portable credentials and selective disclosure proofs, while issuers, verifiers, and platform operators rely on standards-based trust rails, policy controls, and auditability.

## Current phase

This repository is in Phase 1 product hardening for the reusable KYC credential and verifier API loop.
Governance, foundation scaffolding, contract parity, deterministic issuer/verifier application logic, SQL-primary persistence, Hydra-backed internal trust reads, Hydra-backed public issuer/verifier auth, console shells, transfer bridge, sandbox automation, and local Docker Compose packaging are in place.
Wallet flows, selective disclosure, proof verification, self-service partner provisioning, and multi-region production infrastructure remain intentionally deferred.

## Working agreements

- Root operational rules live in [`AGENTS.md`](AGENTS.md).
- Architecture and process decisions are tracked in `docs/`.
- Non-trivial changes require a plan artifact before implementation.
- Architecture, trust-boundary, privacy, and custody changes require stronger review gates.

## Validation

Current foundation validation commands:

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `bash scripts/validate.sh`

## Phase 1 local quickstart

Use [`docs/integration/quickstart.md`](docs/integration/quickstart.md) to run the local Docker Compose stack and prove issue -> allow -> revoke -> deny.

## Environment note

For Windows users, prefer running Codex and development tools from WSL with the repository stored under `/home/...` rather than `/mnt/c/...` to avoid performance and permission issues.
