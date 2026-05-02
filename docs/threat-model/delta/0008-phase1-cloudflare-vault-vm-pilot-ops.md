# 0008 Phase 1 Cloudflare Vault VM Pilot Ops

- Status: accepted
- Date: 2026-05-02
- Owners: repository maintainer

## Change Summary

This delta covers the first controlled Phase 1 pilot operations profile:

- Cloudflare edge for public ingress, WAF, TLS, request-size limits, and rate limiting
- Vault KV v2 for pilot deployment and partner secrets
- one Linux VM running the Phase 1 Docker Compose stack
- operator-run Hydra partner provisioning and public-auth smoke validation

No service contracts, credential semantics, trust decision rules, wallet flows, proof flows, or self-service onboarding change.

## New Or Changed Entry Points

- Cloudflare proxied public hostnames for approved public surfaces
- VM-local reverse proxy routing public hostnames to the Compose services
- Vault KV v2 paths for deployment, service, Hydra, database, and partner secrets
- pilot go/no-go checklist and runbooks

## New Or Changed Privileged Actions

- operator configures Cloudflare DNS, WAF, TLS, and rate-limit rules
- operator writes and reads Vault pilot secrets
- operator renders a VM-local env file from Vault values
- operator provisions and deletes Hydra partner clients
- operator runs public-auth smoke validation against pilot URLs

## Threat Delta

- spoofing: public DNS or reverse-proxy misrouting could direct callers to the wrong origin
- tampering: edge or VM proxy rules could be changed to expose private paths
- repudiation: manual Vault and Cloudflare changes could be unaudited if operator accounts are shared
- information disclosure: logs may capture bearer tokens, client secrets, opaque artifacts, or KYC claims
- denial of service: the single VM can be exhausted if Cloudflare limits are missing or too permissive
- privilege escalation: over-broad Vault policies can let one operator read unrelated partner secrets
- insider abuse: operators can create or delete partner clients without an external approval trail
- replay or relay: stolen bearer tokens remain usable until expiry unless ingress is stopped or the client is deleted
- dependency compromise: Cloudflare, Vault, Hydra, Docker, or VM host compromise affects pilot operation

## Privacy Delta

- no new KYC data is collected
- Vault stores operational secrets, not KYC claims
- Cloudflare and reverse-proxy logs must not store raw request bodies, bearer tokens, opaque artifacts, or normalized KYC claims
- public-auth smoke uses synthetic claims only
- partner onboarding remains manual, so partner identity metadata must be minimized to client ID, scopes, contact channel, and approval reference

## Mitigations

- expose only approved public hostnames through Cloudflare
- keep Hydra admin, SQL, and `trust-registry` private
- require TLS Full Strict and HTTPS redirects
- enforce WAF managed rules, request body limits, and conservative per-IP path limits at Cloudflare
- do not key Cloudflare rate limits on bearer token values
- store real secrets only in Vault or the chosen operator secret channel
- use Vault least-privilege policies by path
- provision Hydra clients with canonical scopes only
- delete compromised Hydra clients for emergency revocation
- run `scripts/phase1-public-auth-smoke.sh` before pilot access and after secret rotation
- preserve SQL and audit state during rollback

## Residual Risks

- single-VM deployment has limited availability and is not global production scale
- Cloudflare per-client limits require additional edge identity controls and may not be available in the first pilot
- Vault operation and unseal/recovery processes remain external to this repo
- partner secret handoff still depends on a secure human process
- self-service approval, onboarding, and credential rotation workflows remain deferred

## Validation Impact

This slice adds runbook-level validation:

- Cloudflare public hostnames and private-surface reachability checks
- Vault path and access-policy checks
- VM firewall checks
- public-auth smoke validation against pilot URLs
- go/no-go checklist before partner access

Repo validation remains:

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `bash scripts/validate.sh`

## Related ADRs, Plans, PRs, And Issues

- `docs/plans/active/0020-phase1-cloudflare-vault-vm-pilot-hardening.md`
- `docs/plans/active/0019-phase1-pilot-readiness-completion.md`
- `docs/adr/0011-phase1-fintech-exchange-deployment-topology.md`
- `docs/adr/0012-phase1-public-hydra-auth.md`
- `docs/threat-model/delta/0007-phase1-pilot-operations.md`
- `docs/runbooks/phase1-cloudflare-edge.md`
- `docs/runbooks/phase1-vault-secrets.md`
- `docs/runbooks/phase1-single-vm-compose.md`
- `docs/runbooks/phase1-pilot-go-no-go.md`
