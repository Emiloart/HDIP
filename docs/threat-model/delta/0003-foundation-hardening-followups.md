# 0003 Foundation Hardening Follow-ups

- Status: accepted
- Date: 2026-04-18
- Owners: repository maintainer

## Change summary

Harden client and service foundation behavior by wrapping malformed transport payload failures in the TypeScript API client, rejecting malformed Go service environment values at startup, switching Go formatting validation to non-mutating checks, and refreshing stale root status docs.

## New or changed entry points

- none

## New or changed privileged actions

- none

## Threat delta

- Reduces the chance that malformed upstream responses escape typed client boundaries as raw exceptions.
- Reduces the chance that operators start services with invalid runtime configuration while believing defaults were not in effect.
- Reduces CI ambiguity by ensuring Go validation reports formatting drift instead of silently rewriting source during validation.

## Privacy delta

- none

## Mitigations

- convert malformed response parsing into explicit typed client failures
- fail startup on invalid integer or duration environment values
- make validation fail on formatting drift without mutating tracked files
- refresh root docs so contributor expectations match the actual foundation state

## Residual risks

- client and schema parity is still manually maintained
- service config parsing remains duplicated across services for now

## Validation impact

- full repo validation remains required
- add focused TS and Go tests for the new failure paths

## Related ADRs, plans, PRs, and issues

- `docs/adr/0004-foundation-service-and-schema-baseline.md`
- `docs/plans/active/0003-foundation-hardening-followups.md`
