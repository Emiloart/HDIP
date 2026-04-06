# ADR Guide

## Purpose

Architecture Decision Records capture decisions that materially affect HDIP architecture, trust posture, privacy posture, portability, or operational shape.

## Status lifecycle

- `proposed`
- `accepted`
- `superseded`
- `rejected`

Only `accepted` ADRs have the higher rule precedence referenced by `AGENTS.md`.

## File naming

Use `NNNN-short-kebab-title.md`.

Examples:

- `0001-governance-spine-and-bootstrap.md`
- `0002-hdip-initial-platform-baseline.md`

## When an ADR is required

Use the closed trigger list below.
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

If a change does not match this list, do not require an ADR unless explicitly escalated.

## Review expectations

- The ADR must exist before or alongside implementation.
- Security impact and privacy impact are mandatory sections.
- Alternatives must be documented.
- Superseding a prior ADR must explicitly link both records.

## Template

Start from `docs/adr/0000-template.md`.

## Required sections

- Title
- Status
- Date
- Owners
- Context
- Decision
- Alternatives considered
- Security impact
- Privacy impact
- Migration / rollback
- Consequences
- Open questions
- Related plans, PRs, and issues
