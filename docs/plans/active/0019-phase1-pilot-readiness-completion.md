# 0019 Phase 1 Pilot Readiness Completion

- Status: active
- Date: 2026-04-30
- Owners: repository maintainer

## Objective

Close the remaining Phase 1 pilot-readiness gap by adding operational tooling and runbooks for controlled partner provisioning, public-auth smoke validation, production configuration separation, and edge/observability readiness.

This slice does not add product behavior.
It operationalizes the already-governed Phase 1 system for 1-2 controlled fintech/exchange pilots.

## Scope

- Add an operator-driven Hydra client provisioning script for issuer and verifier partners.
- Add an emergency client deletion path for revocation or containment.
- Add a public-auth smoke script that proves the packaged/pilot public Hydra path.
- Add a production environment template with placeholders only.
- Add partner provisioning and pilot readiness runbooks.
- Add a threat delta for pilot operations.
- Update stale docs that still describe public auth as future work.

## Out Of Scope

- API keys.
- Self-service partner onboarding.
- New issuer, verifier, trust, schema, or credential contracts.
- Wallet flows.
- Selective disclosure or proof verification.
- New trust decision rules.
- Core-service rate limiting.
- Cloud-provider-specific infrastructure.
- Full production global scale.

## Affected Files, Services, Or Packages

- `scripts/phase1-provision-client.sh`
- `scripts/phase1-public-auth-smoke.sh`
- `scripts/validate.sh`
- `infra/phase1/production.env.template`
- `docs/runbooks/phase1-partner-provisioning.md`
- `docs/runbooks/phase1-pilot-readiness.md`
- `docs/threat-model/delta/0007-phase1-pilot-operations.md`
- `docs/deployment/phase1-fintech-exchange-cloud-infra.md`
- `docs/integration/quickstart.md`
- `docs/repo-structure.md`

## Assumptions

- Ory Hydra remains the public Phase 1 OAuth2 server.
- Public partner access uses client credentials only.
- Issuer and verifier scopes remain the canonical scopes from ADR 0008 and ADR 0012.
- Partner provisioning is operator-driven for Phase 1.
- Generated client secrets are printed once to the operator and are not stored by repo scripts.
- Edge controls are deployment responsibilities and must not be embedded into issuer/verifier business logic.

## Risks

- Operators may mishandle generated client secrets after provisioning.
- A partner may be over-scoped if provisioning is allowed to accept arbitrary scopes.
- A smoke script can leak tokens or artifacts if it prints raw responses.
- Production templates can accidentally normalize local defaults if placeholders are not explicit.
- Leaving historical docs stale can cause operators to use header auth or skip public-auth validation.

## Validation Steps

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `bash scripts/validate.sh`
- `bash -n scripts/phase1-provision-client.sh`
- `bash -n scripts/phase1-public-auth-smoke.sh`
- Compose public-auth smoke:
  - `docker compose --env-file infra/phase1/.env.example -f infra/phase1/docker-compose.yml up --build -d`
  - `bash scripts/phase1-public-auth-smoke.sh`
  - `docker compose --env-file infra/phase1/.env.example -f infra/phase1/docker-compose.yml down -v`

## Rollback Or Containment Notes

Rollback by removing this slice's scripts, runbooks, validation hook, and production template.
If a provisioned partner client is compromised, delete the Hydra client immediately and stop public ingress if token misuse is suspected.
Do not use header attribution as a production rollback.

## Open Questions

- Which edge provider or gateway will enforce pilot rate limits.
- Whether the first pilot uses host-installed Hydra CLI or an operator container with Hydra CLI.
- Whether later partner onboarding should become a governed service or admin console workflow.

## Implementation Sequence

1. Archive stale product-layer plan artifacts that still describe public auth as future work.
2. Add the pilot-readiness plan and threat delta.
3. Add the Hydra client provisioning script.
4. Add the public-auth smoke script.
5. Add the production env template and runbooks.
6. Update deployment, quickstart, and repo status docs.
7. Add an opt-in validation gate for public-auth smoke.
8. Run validation and packaged smoke.
