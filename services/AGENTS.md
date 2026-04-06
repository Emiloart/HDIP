# AGENTS.md

## Scope

This directory contains HDIP backend services and control-plane services.

## Rules

- Keep trust boundaries explicit.
- Do not mix transport glue with cryptographic authority.
- Treat APIs and events as contracts.
- Use narrow service responsibilities and auditable privileged paths.
- Shared security-sensitive primitives belong in `crates/`, not duplicated ad hoc in services.
