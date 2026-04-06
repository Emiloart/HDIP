# Threat Model Guide

## Purpose

Threat modeling is mandatory for HDIP because identity, trust, privacy, recovery, and operator control are high-risk areas.

## Review tiers

Use two levels:

### Threat delta

Required for:

- moderate workflow changes
- new endpoints or events
- new privileged actions
- sensitive logging or telemetry changes
- external integration changes

### Full threat-model update

Required for:

- trust-boundary changes
- custody or recovery changes
- auth or authorization changes
- issuance, presentation, or verification changes
- key-handling changes
- admin-plane changes
- regional data-flow changes

## File layout

- `docs/threat-model/template-full.md`
- `docs/threat-model/template-delta.md`
- `docs/threat-model/full/`
- `docs/threat-model/delta/`

## Required analysis areas

At minimum, threat review must consider:

- spoofing
- tampering
- repudiation
- information disclosure
- denial of service
- privilege escalation
- insider abuse
- account takeover
- device compromise
- recovery abuse
- verifier over-collection
- issuer compromise
- replay or relay
- correlation and linkability
- dependency compromise

## Baseline

The current repository baseline threat model is:

- `docs/threat-model/full/0001-platform-bootstrap.md`
