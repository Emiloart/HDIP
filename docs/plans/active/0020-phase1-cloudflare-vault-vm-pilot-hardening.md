# 0020 Phase 1 Cloudflare Vault VM Pilot Hardening

- Status: active
- Date: 2026-05-02
- Owners: repository maintainer

## Objective

Close the remaining Phase 1 pilot-operations gap by standardizing the first controlled pilot on Cloudflare edge controls, Vault-backed secret handling, and a single Linux VM running the existing Phase 1 Docker Compose stack.

This plan does not add product behavior.
It makes the existing reusable-KYC issuer/verifier loop safe enough for 1-2 controlled fintech/exchange pilots.

## Scope

- Add Cloudflare edge requirements for DNS, TLS, WAF, request-size limits, and rate limits.
- Add Vault KV v2 secret-management requirements and paths for Phase 1 pilot secrets.
- Add a single-VM Docker Compose deployment runbook.
- Add a pilot go/no-go checklist.
- Add a threat delta for Cloudflare, Vault, VM, and operator-run provisioning.
- Keep rate limiting, WAF, and ingress policy outside issuer/verifier business logic.
- Keep partner onboarding operator-driven.

## Out Of Scope

- Wallet flows.
- Selective disclosure, ZK, proof verification, or signed credential artifacts.
- New API contracts, schemas, trust rules, or credential semantics.
- API keys or a second public auth model.
- Self-service partner onboarding.
- Kubernetes, multi-region production, or global-scale deployment.
- Implementing Cloudflare, Vault, or VM provisioning automation in repo code.

## Affected Files, Services, Or Packages

- `docs/runbooks/phase1-cloudflare-edge.md`
- `docs/runbooks/phase1-vault-secrets.md`
- `docs/runbooks/phase1-single-vm-compose.md`
- `docs/runbooks/phase1-pilot-go-no-go.md`
- `docs/threat-model/delta/0008-phase1-cloudflare-vault-vm-pilot-ops.md`
- `docs/deployment/phase1-fintech-exchange-cloud-infra.md`
- `docs/runbooks/phase1-pilot-readiness.md`
- `infra/phase1/production.env.template`
- `infra/phase1/docker-compose.yml`

## Assumptions

- Cloudflare is the first pilot edge provider.
- Vault KV v2 is the first pilot secret manager.
- The first pilot uses a single Linux VM with Docker Compose.
- Hydra client credentials remain the only public partner auth model.
- Partner client provisioning remains operator-driven through `scripts/phase1-provision-client.sh`.
- Cloudflare public ingress reaches a VM-local reverse proxy that forwards only approved public surfaces.
- Hydra admin, SQL, and `trust-registry` remain private.

## Risks

- Cloudflare rules can be misconfigured and expose internal routes or fail to enforce rate limits.
- Vault access can be over-broad, allowing operators or automation to read partner secrets unnecessarily.
- Single-VM deployment concentrates availability risk and must not be mistaken for global production scale.
- Compose defaults can accidentally leak into a pilot if the operator uses `.env.example`.
- Partner credentials can leak during manual handoff.
- Edge logs, reverse-proxy logs, or support captures can accidentally include bearer tokens or opaque artifacts.

## Validation Steps

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `bash scripts/validate.sh`
- Confirm `infra/phase1/docker-compose.yml` still defaults to local development when no pilot env is supplied.
- Confirm pilot env values are placeholders only in `infra/phase1/production.env.template`.
- Manual pilot readiness checks from `docs/runbooks/phase1-pilot-go-no-go.md`.

## Rollback Or Containment Notes

- Stop Cloudflare public ingress first.
- Delete compromised Hydra clients with `scripts/phase1-provision-client.sh delete`.
- Preserve SQL and audit state.
- Revoke or rotate Vault secrets before re-enabling partner access.
- Roll back all three services together to the previous tested image or commit.
- Do not use trusted-header auth as a production rollback path.

## Open Questions

- Exact pilot domain name.
- Exact VM provider and public IP.
- Vault address, auth method, and operator policy names.
- First real issuer organization ID and verifier organization ID.
- Secure partner credential delivery channel.

## Implementation Sequence

1. Add the Cloudflare, Vault, single-VM, and go/no-go runbooks.
2. Add this plan and the matching threat delta.
3. Update pilot readiness and deployment docs to point operators at the new runbooks.
4. Update the production env template with Cloudflare/Vault/Compose pilot placeholders.
5. Make Compose API environment selection configurable while preserving local defaults.
6. Run governance, secret, and full validation.
