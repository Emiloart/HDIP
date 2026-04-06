# Plan Artifacts

## Purpose

Plan artifacts are required for non-trivial work.
The repository does not use one shared `PLANS.md`.
Each meaningful effort gets its own plan document.

## Layout

- `docs/plans/active/` for current work
- `docs/plans/archive/` for completed or superseded work

## Naming

Use `NNNN-short-kebab-title.md`.

## Minimum sections

- objective
- scope
- out of scope
- affected files, services, or packages
- assumptions
- risks
- validation steps
- rollback or containment notes
- open questions

## Rules

- Plans may refine execution but may not weaken higher-precedence governance.
- If implementation diverges from the plan, update the plan in the same change.
- A stale plan is a defect.
