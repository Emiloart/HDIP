# 0007 Phase 1 Pilot Operations

- Status: accepted
- Date: 2026-04-30
- Owners: repository maintainer

## Change Summary

This delta covers Phase 1 pilot operationalization:

- operator-driven Hydra issuer/verifier client provisioning
- emergency public client deletion
- public-auth smoke automation against a running packaged or pilot stack
- production configuration template separation
- edge and observability readiness runbooks

No service contracts, credential semantics, trust decision rules, storage schemas, wallet behavior, or proof flows change in this slice.

## New Or Changed Entry Points

- `scripts/phase1-provision-client.sh`
- `scripts/phase1-public-auth-smoke.sh`
- optional `HDIP_VALIDATE_PHASE1_PUBLIC_AUTH_SMOKE=1` validation gate
- operator-facing runbooks for provisioning and pilot readiness

## New Or Changed Privileged Actions

- deployment operator creates a Hydra issuer client
- deployment operator creates a Hydra verifier client
- deployment operator deletes a Hydra client for emergency revocation
- deployment operator runs a public-auth smoke test with partner credentials

## Threat Delta

- over-scoped clients could allow unauthorized issuance or verification
- generated client secrets could leak through terminal history, support tickets, logs, screenshots, or pasted output
- public-auth smoke automation could print bearer tokens, opaque artifacts, or KYC claims if not bounded
- production config templates could accidentally include local defaults
- partner access could be enabled before edge limits, private network boundaries, and readiness checks are confirmed
- client deletion may be mistaken for full credential rotation if downstream partners keep stale secrets

## Privacy Delta

- no new KYC data is collected
- smoke automation creates synthetic KYC claims only
- partner runbooks must instruct integrators to store decisions and audit identifiers, not opaque artifacts or raw claims
- operational logs must not include bearer tokens, client secrets, opaque artifacts, normalized claims, or raw request bodies

## Mitigations

- provisioning uses fixed canonical issuer/verifier scopes only
- provisioning outputs generated secrets once and never writes them to repo files
- emergency revocation deletes the Hydra client by explicit client ID
- public-auth smoke prints only credential ID and decisions
- production template uses placeholders only
- pilot readiness runbook requires private Hydra admin, private trust-registry, private SQL, TLS, request-size limits, and rate limiting at the edge
- header attribution remains local/test only and is not a production rollback path

## Residual Risks

- client credential storage after provisioning is outside repo automation and depends on operator secret handling
- edge rate limiting remains provider-specific and is documented rather than enforced by repo code
- full client rotation remains delete plus recreate until a later governed lifecycle exists
- no self-service onboarding or approval workflow exists in Phase 1

## Validation Impact

This slice adds:

- shell syntax checks for provisioning and smoke scripts
- optional public-auth smoke validation for running stacks
- updated governance and secret scans
- full repo validation remains required

## Related ADRs, Plans, PRs, And Issues

- `docs/plans/active/0019-phase1-pilot-readiness-completion.md`
- `docs/adr/0012-phase1-public-hydra-auth.md`
- `docs/adr/0011-phase1-fintech-exchange-deployment-topology.md`
- `docs/threat-model/full/0004-phase1-public-hydra-auth.md`
- `docs/deployment/phase1-fintech-exchange-cloud-infra.md`
