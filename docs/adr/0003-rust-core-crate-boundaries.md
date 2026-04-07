# 0003 Rust Core Crate Boundaries

- Status: accepted
- Date: 2026-04-06
- Owners: repository maintainer

## Context

HDIP needs a trustworthy Rust foundation for security-sensitive logic before wallet code, service integrations, or proof systems are implemented.
The initial crate boundaries must stay narrow so cryptographic and identity-domain logic do not get mixed with transport or storage concerns.

## Decision

Adopt two Rust foundation crates for the first executable slice:

- `crates/crypto-core`
- `crates/identity-core`

`crypto-core` owns:

- key abstraction interfaces
- signature abstraction interfaces
- hashing helpers
- canonical byte helpers
- sensitive wrapper types with redacted debug output
- deterministic error types for crypto-core concerns

`identity-core` owns:

- identifier models
- credential metadata models
- presentation request and response models
- validation interfaces
- status-check interfaces

Both crates must remain transport-agnostic, storage-agnostic, and framework-agnostic.
No placeholder cryptography that could accidentally be treated as production-ready is allowed.

`policy-core` is deferred until shared policy logic exists and is justified by a later ADR.

## Alternatives considered

### Single `core` crate

Rejected because it would blur boundaries too early and encourage transport and policy drift into the same package.

### Adding `policy-core` immediately

Rejected because there is not yet enough real shared policy logic to justify it.

## Security impact

Positive.
This isolates security-sensitive abstractions and reduces the chance of hidden runtime coupling.

## Privacy impact

Positive.
Identity models remain explicit and inspectable, which helps preserve data-minimization discipline.

## Migration / rollback

If these boundaries prove too coarse or too fine, adjust them before real product logic accumulates.
Any later split or merge requires follow-up documentation if it changes architectural expectations.

## Consequences

- Rust code will remain small and deliberate early on.
- Services and apps must consume Rust core through clear boundaries rather than reimplementing domain models ad hoc.

## Open questions

- whether key-material zeroization support should be added in the next Rust-focused slice
- whether canonical serialization should remain byte-oriented or gain structured helpers later

## Related plans, PRs, and issues

- `docs/plans/active/0002-foundation-workspace-and-service-skeletons.md`
