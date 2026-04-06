# HDIP

HDIP is a hybrid decentralized identity and trust infrastructure platform.
Users hold portable credentials and selective disclosure proofs, while issuers, verifiers, and platform operators rely on standards-based trust rails, policy controls, and auditability.

## Current phase

This repository is in foundation bootstrap.
Governance, architecture, repo structure, and security/privacy constraints are being established before product implementation.

## Working agreements

- Root operational rules live in [`AGENTS.md`](AGENTS.md).
- Architecture and process decisions are tracked in `docs/`.
- Non-trivial changes require a plan artifact before implementation.
- Architecture, trust-boundary, privacy, and custody changes require stronger review gates.

## Validation

Current bootstrap validation commands:

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `bash scripts/validate.sh`

## Environment note

For Windows users, prefer running Codex and development tools from WSL with the repository stored under `/home/...` rather than `/mnt/c/...` to avoid performance and permission issues.
