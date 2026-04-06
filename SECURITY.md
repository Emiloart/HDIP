# Security Policy

## Project status

HDIP is pre-release identity infrastructure.
Assume interfaces, storage models, and trust boundaries are still under active design review.

## Supported versions

At this stage, only the default branch is supported.
There are no versioned releases yet.

## Reporting a vulnerability

- Do not open a public issue for a suspected vulnerability.
- Use GitHub private vulnerability reporting if it is enabled for this repository.
- Otherwise contact the repository maintainer through a private channel before disclosure.

Include:

- affected area
- impact
- reproduction steps
- prerequisites
- proposed mitigation if known

## Handling expectations

- Do not post secrets, raw PII, credentials, recovery material, or exploit data in public threads.
- Prefer minimal proof-of-concept data.
- Coordinate disclosure timing with maintainers.

## Security bar for contributions

- No hidden bypasses.
- No privileged behavior without explicit authn/authz treatment.
- No sensitive data in logs, fixtures, screenshots, or examples.
- No unverifiable security claims in docs or code comments.
