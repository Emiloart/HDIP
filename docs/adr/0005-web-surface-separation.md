# 0005 Web Surface Separation

- Status: accepted
- Date: 2026-04-06
- Owners: repository maintainer

## Context

HDIP requires multiple operator-facing web surfaces over time, but the first executable slice should stay narrow.
Frontend code must not become the hidden home of trust logic, credential logic, or privileged backend policy.

## Decision

For the first executable slice, build only:

- `apps/issuer-console`
- `apps/verifier-console`

Use shared frontend packages for:

- `packages/ui`
- `packages/api-client`
- `packages/config-typescript`

Each app starts as a shell only:

- app layout
- auth placeholder boundary
- env/config loading
- typed API client boundary
- loading and error states

Do not build dashboard business logic, real auth, or deep domain workflows in this slice.

## Alternatives considered

### One combined operations console

Rejected because issuer and verifier responsibilities will diverge and combining them too early encourages hidden coupling.

### Building all planned web surfaces immediately

Rejected because it expands scope without increasing foundation quality.

## Security impact

Positive.
Frontends remain thin and do not take on privileged decision logic.

## Privacy impact

Positive.
Typed boundaries and shell-only scope reduce the risk of accidental sensitive-data spread into the UI layer.

## Migration / rollback

If a combined console becomes justified later, it should be an explicit decision rather than accidental drift.

## Consequences

- shared frontend packages must stay narrow
- backend contracts must be consumed through explicit clients
- future new web surfaces should be added deliberately, not opportunistically

## Open questions

- whether the developer portal should join the next slice or remain deferred until SDK assets exist

## Related plans, PRs, and issues

- `docs/plans/active/0002-foundation-workspace-and-service-skeletons.md`
