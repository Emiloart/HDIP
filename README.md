# HDIP

HDIP is a hybrid decentralized identity passport and trust infrastructure platform.
Users hold portable credentials and selective disclosure proofs, while issuers, verifiers, and platform operators rely on standards-based trust rails, policy controls, and auditability.

## Current phase

This repository has moved past governance-only bootstrap into the executable foundation slice.
The Rust core crates, Go service skeletons, web shells, shared packages, schemas, validation wiring, and the first deterministic stub issuer/verifier flow are in place, while real issuance and verification logic remain intentionally deferred.

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

## Environment note

For Windows users, prefer running Codex and development tools from WSL with the repository stored under `/home/...` rather than `/mnt/c/...` to avoid performance and permission issues.
