# AGENTS.md

## Mission
Build HDIP as secure, privacy-preserving, standards-aligned identity infrastructure with explicit trust boundaries, strong interoperability, and zero undocumented behavior in sensitive paths.

## What this file is
This is the root operational dispatcher for agent work.
It is not the full constitution of the repo.

Agents must use this file together with:
- `docs/repo-structure.md`
- `docs/adr/README.md`
- `docs/threat-model/README.md`
- `docs/standards/README.md`
- `docs/privacy/README.md`
- `docs/plans/README.md`
- `CONTRIBUTING.md`
- `.github/pull_request_template.md`
- relevant accepted ADRs
- relevant active plan artifacts

## Rule precedence
Apply rules in this order:

1. Security and privacy invariants
2. Accepted ADRs
3. Approved standards docs
4. Root governance docs
5. Area-specific docs and local AGENTS guidance
6. Active plan artifacts
7. Task-specific instructions

Plans and task-specific instructions may refine execution, but may not weaken higher-precedence rules.

## Non-negotiables
- No silent architectural drift.
- No undocumented trust-boundary changes.
- No hidden admin or debug bypasses.
- No secrets, raw PII, raw credentials, recovery material, or sensitive tokens in logs, fixtures, screenshots, or examples.
- No production dependency additions without explicit rationale.
- No public contract change without explicit review.
- No data model change without migration notes.
- No completion claim without truthful validation evidence.
- No unrelated edits.
- No improvised requirements in ambiguous security- or privacy-sensitive areas.

## Before any non-trivial change
Read:
1. `docs/repo-structure.md`
2. relevant accepted ADRs in `docs/adr/`
3. relevant threat-model docs in `docs/threat-model/`
4. the relevant plan artifact in `docs/plans/active/`
5. any more-local `AGENTS.md`

If a required governance artifact is missing, create or repair it before implementation.

If the task touches auth, authz, keys, wallet logic, recovery, issuance, presentation, verification, admin powers, logging, storage, privacy, infra, or external integrations, these reads are mandatory.

## Planning gate
Create or update a plan artifact before implementation if the task:
- spans multiple files with behavioral impact
- spans multiple packages or services
- changes trust boundaries
- changes auth, authz, custody, recovery, issuance, presentation, or verification
- changes storage, migrations, APIs, infra, or deployment
- adds a dependency
- is ambiguous or high-risk

Minimum plan contents:
- objective
- scope
- out of scope
- affected files/services
- assumptions
- risks
- validation steps
- rollback/containment notes
- open questions

## ADR gate
An ADR is required for:
- auth model changes
- authorization model changes
- wallet custody or recovery model changes
- credential format or proof model changes
- DID / identifier model changes
- trust registry model changes
- signing or key-management changes
- storage engine changes
- queue or event topology changes
- deployment topology changes
- API versioning strategy changes
- new production dependencies with architectural impact
- privacy architecture changes
- regional data-boundary model changes

## Threat-model gate
Use a threat delta for:
- moderate workflow changes
- new endpoints or events
- new privileged actions
- sensitive logging/telemetry changes
- external integration changes

Use a full threat-model update for:
- trust-boundary changes
- custody or recovery changes
- auth/authz changes
- issuance/presentation/verification changes
- key-handling changes
- admin-plane changes
- regional data-flow changes

## Definition of done
A task is only done when:
- implementation is complete for stated scope
- validation was run and reported truthfully
- required docs were updated
- ADR was added/updated if required
- threat-model artifacts were updated if required
- remaining gaps are stated explicitly

## Required completion report
At the end of meaningful work, report:
1. what changed
2. files changed
3. validation run
4. security impact (`none` or described)
5. privacy impact (`none` or described)
6. remaining gaps
7. follow-up work, if any

## Current repo validation commands
- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `bash scripts/validate.sh`
- `make validate` as an optional wrapper when `make` is available

## Final rule
When uncertain in a high-trust system, stop and surface the ambiguity.
Do not improvise past missing governance.
