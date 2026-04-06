# 0001 Governance Spine And Bootstrap

- Status: accepted
- Date: 2026-04-06
- Owners: repository maintainer

## Context

HDIP is starting as a greenfield repository for identity infrastructure.
This category requires stronger controls than a generic application because architecture, trust boundaries, privacy promises, and auditability are product requirements.
The repository also needs to work well with Codex and human contributors without relying on ad hoc prompt memory.

## Decision

Adopt a layered governance spine with:

- a short root `AGENTS.md` dispatcher
- deeper governance documents in `docs/`
- plan artifacts in `docs/plans/`
- a closed ADR trigger list
- tiered threat-model review
- repo-local Codex configuration under `.codex/`
- repo-local validation and CI checks for bootstrap governance

Root rule precedence is:

1. Security and privacy invariants
2. Accepted ADRs
3. Approved standards docs
4. Root governance docs
5. Area-specific docs and local AGENTS guidance
6. Active plan artifacts
7. Task-specific instructions

## Alternatives considered

### Single large root AGENTS file

Rejected because it becomes brittle and mixes operational dispatch with deep policy detail.

### Prompt-only governance

Rejected because it is not durable, reviewable, or enforceable.

### No plan artifacts

Rejected because it weakens traceability for architectural and security-sensitive work.

## Security impact

Positive.
This decision improves reviewability, documents trust-boundary changes, and reduces accidental drift.

## Privacy impact

Positive.
Privacy review becomes an explicit gate rather than an implicit assumption.

## Migration / rollback

The repository can evolve these documents incrementally.
If this governance model proves too heavy, relaxation must be explicit through follow-up ADRs rather than ad hoc bypass.

## Consequences

- Contributors must accept more process for sensitive work.
- The repo will carry governance artifacts from the start.
- Machine enforcement starts simple and will expand as code arrives.

## Open questions

- Whether plan artifacts should eventually be mirrored into issues or project boards
- Whether protected-path hooks should later become stronger than the current session-start reminder

## Related plans, PRs, and issues

- `docs/plans/active/0001-governance-and-foundation-bootstrap.md`
