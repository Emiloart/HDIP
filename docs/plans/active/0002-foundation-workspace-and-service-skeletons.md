# 0002 Foundation Workspace And Service Skeletons

- Status: active
- Date: 2026-04-06
- Owners: repository maintainer

## Objective

Establish the first executable HDIP foundation slice with locked workspace boundaries, compileable Rust core crates, runnable Go service skeletons, minimal web shells, contract-first schemas, and stronger validation before any business logic lands.

## Scope

- WSL-native working copy under `/home/...`
- root workspace files for Rust, Go, and Node
- Rust foundation crates: `crates/crypto-core`, `crates/identity-core`
- Go services: `services/issuer-api`, `services/verifier-api`, `services/trust-registry`
- shared Go foundation package for service skeleton concerns only if needed
- Next.js shells: `apps/issuer-console`, `apps/verifier-console`
- shared frontend packages for UI and typed API boundaries
- initial JSON Schemas for common contracts
- CI and validation expansion for Rust, Go, TypeScript, and schemas

## Out of scope

- database integration
- auth or passkey flows
- credential issuance logic
- credential verification logic
- wallet logic
- recovery logic
- external integrations
- analytics, AI, or chain anchoring

## Affected files, services, or packages

- `docs/repo-structure.md`
- `docs/architecture/platform-baseline.md`
- `docs/adr/`
- `docs/threat-model/`
- `docs/plans/`
- root Rust, Go, and Node workspace files
- `crates/crypto-core/`
- `crates/identity-core/`
- `services/issuer-api/`
- `services/verifier-api/`
- `services/trust-registry/`
- `packages/ui/`
- `packages/api-client/`
- `packages/config-typescript/`
- `apps/issuer-console/`
- `apps/verifier-console/`
- `schemas/`
- `scripts/`
- `.github/workflows/`

## Exact workspace layout

- Rust: one root `Cargo.toml` workspace with `crates/crypto-core` and `crates/identity-core`
- Go: one root `go.work` referencing `packages/go/foundation`, `services/issuer-api`, `services/verifier-api`, and `services/trust-registry`
- Node: one root `package.json` using npm workspaces for `apps/*` and `packages/*`
- Schemas: contract-first JSON Schema files under `schemas/json/`

## Crate boundaries

- `crypto-core`: key abstractions, signature abstractions, hashing helpers, canonical byte helpers, redacted sensitive wrappers, deterministic crypto-core error types
- `identity-core`: identifiers, credential metadata, presentation request and response models, validation interfaces, status-check interfaces
- `policy-core`: intentionally deferred until policy logic has a justified shared surface

## Service boundaries

- `issuer-api`: issuer-facing API shell, health, readiness, structured logging, middleware, config
- `verifier-api`: verifier-facing API shell, health, readiness, structured logging, middleware, config
- `trust-registry`: trust metadata API shell, health, readiness, structured logging, middleware, config
- `gateway`: intentionally deferred until routing duplication becomes real rather than speculative

## Package ownership

- `packages/ui`: shared React shell components only
- `packages/api-client`: frontend-safe error envelope and typed API request helpers
- `packages/config-typescript`: shared TS compiler config
- `packages/go/foundation`: startup, logging, middleware, and response helpers for Go services only

## Stubbed vs real

Real in this slice:

- workspace wiring
- buildable crate and service entrypoints
- health and readiness endpoints
- config validation
- structured safe logging
- request ID and timeout middleware
- shared error envelope shape
- JSON Schema contract files
- lint, typecheck, and test wiring

Stubbed intentionally:

- auth
- storage
- issuance
- verification
- trust evaluation
- recovery
- mobile code
- external network integrations beyond local tooling install

## Assumptions

- Rust and Go toolchains may need to be installed in the environment during this slice
- npm is available and adequate for the initial JS workspace
- schema-first JSON contracts are sufficient until real inter-service generation needs appear

## Risks

- adding multiple language toolchains in one slice increases bootstrap complexity
- early shared packages can become dumping grounds if boundaries are not kept narrow
- schema files can drift from implementations unless validation is wired immediately

## Validation steps

- Rust: `cargo fmt --check`, `cargo clippy --all-targets --all-features -- -D warnings`, `cargo test`
- Go: `gofmt -w` or `gofmt -l`, `go test ./...`, `go vet ./...`
- Node: `npm run lint`, `npm run typecheck`, `npm run test`
- Schemas: `npm run schema:validate`
- Repo: `bash scripts/check-governance.sh`, `bash scripts/check-no-secrets.sh`

## Rollback or containment notes

All additions in this slice are foundation-only and can be reverted cleanly before business logic depends on them.
If a language workspace choice proves incorrect, revert the scaffolding and replace it before feature code lands.

## Open questions

- whether schema generation should later produce Rust and Go types or remain schema-validation-only for a while
- when a separate gateway becomes justified
- whether frontend shared packages should later split into `ui` and `config-web`
