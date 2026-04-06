# 0001 Governance And Foundation Bootstrap

- Status: active
- Date: 2026-04-06
- Owners: repository maintainer

## Objective

Bootstrap HDIP with a durable governance spine, repo-local Codex configuration, architecture baseline documentation, and a workspace skeleton that reflects the accepted technical stack without shipping product logic yet.

## Scope

- root governance files
- deep governance docs
- initial ADRs
- initial threat-model baseline
- standards and privacy registries
- repo-local Codex config and hooks
- bootstrap validation scripts and CI
- architecture baseline docs
- initial repository directory scaffold

## Out of scope

- product implementation
- runtime services
- wallet application logic
- database schemas
- deployment manifests
- production secrets or credentials

## Affected files, services, or packages

- root governance files
- `.codex/`
- `.github/`
- `docs/`
- `scripts/`
- scaffold directories under `apps/`, `mobile/`, `services/`, `crates/`, `packages/`, `infra/`, and `schemas/`

## Assumptions

- the remote repository is effectively greenfield
- the current workspace remains under `/mnt/c/...` during bootstrap only
- the approved stack baseline in ADR `0002` is the current architectural direction

## Risks

- the repo location under `/mnt/c/...` may continue to cause friction until moved into WSL storage
- no production code means validation is still limited
- some workspace tooling choices remain intentionally deferred

## Validation steps

- run `make governance-check`
- run `make security-check`
- review the resulting repository structure and documents for consistency

## Rollback or containment notes

This bootstrap is documentation- and scaffold-heavy.
Rollback is straightforward through normal git revert or selective removal of scaffold directories before runtime code depends on them.

## Open questions

- whether the workspace should standardize on `pnpm` workspaces alone or add a higher-level monorepo tool
- how local orchestration should be structured once the first runnable services land
- when to split data-plane and control-plane infra into separate directories
