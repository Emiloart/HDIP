# AGENTS.md

## Scope

This directory contains native holder wallet applications.

## Rules

- Preserve holder control as a product requirement.
- Do not let wallet UX drift into undocumented custodial behavior.
- Keep consent and selective-disclosure flows explicit.
- Route security-critical logic through the Rust core rather than duplicating it in platform code.
