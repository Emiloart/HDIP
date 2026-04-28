# 0016 Validation Deterministic Cross Environment

- Status: completed
- Date: 2026-04-28
- Owners: repository maintainer

## Objective

Make repository validation deterministic and environment-agnostic across WSL home-directory checkouts, WSL mounted-path checkouts, and Linux CI without changing product contracts, runtime behavior, or business logic.

## Scope

- remove Windows PowerShell branching from validation flow
- standardize web validation on POSIX shell execution under bash
- make Node and npm resolution consistent through the shared toolchain environment
- make workspace lint, typecheck, and test execution explicit and predictable
- keep validation path behavior consistent under `/home/...`, `/mnt/c/...`, and CI Linux checkouts
- run frontend validation from a Linux-local mirrored worktree so mounted-path execution does not depend on host filesystem quirks

## Out of scope

- service, API, schema, auth, SQL lifecycle, or runtime behavior changes
- product feature work
- new business logic or contract changes
- CI topology redesign beyond using the existing bash validation path

## Affected files, services, or packages

- `docs/plans/active/0016-validation-deterministic-cross-environment.md`
- `scripts/validate.sh`
- `scripts/validate-web.sh`
- `scripts/toolchain-env.sh`
- root `package.json`
- workspace `package.json` files only if required for deterministic validation

## Assumptions

- bash is the canonical validation shell for local and CI execution
- Node 22 remains the required validation baseline
- the repo's recommended WSL workflow remains valid, but validation must still behave predictably when the checkout is under `/mnt/c`
- existing unrelated dirty worktree changes outside this validation slice will remain uncommitted and must not be modified by this task

## Risks

- overcorrecting path handling could break CI or standard WSL home-directory validation
- changing root or workspace npm scripts carelessly could create recursion or script-selection drift
- mounted-path performance may still be slower than `/home/...` even after behavior is made deterministic

## Validation steps

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `npm run schema:validate`
- `bash scripts/validate.sh`

## Rollback or containment notes

If the deterministic bash-only validation path is incorrect, roll back validation-script and package-metadata changes together.
Do not restore PowerShell-specific branching as a hidden fallback.

## Open questions

- none for this slice
