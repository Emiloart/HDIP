# AGENTS.md

## Scope

This directory contains Rust security-critical components.

## Rules

- Keep implementations deterministic and reviewable.
- Avoid uncontrolled network I/O.
- Minimize dependencies, especially in cryptographic paths.
- Do not log secrets, credentials, or sensitive payloads.
- Prefer explicit APIs and exhaustive test coverage.
