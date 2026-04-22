# 0013 Phase 1 Hydra Internal Trust Auth And Phase1SQL Lifecycle

- Status: active
- Date: 2026-04-22
- Owners: repository maintainer

## Objective

Replace the transitional static bearer-token trust runtime auth path with the governed Hydra client-credentials plus introspection model, and replace implicit startup SQL initialization with the explicit `phase1sql` migration and bootstrap lifecycle, without changing current public Phase 1 issuer or verifier contracts.

## Scope

- add Hydra client-credentials token acquisition in `verifier-api` for trust runtime reads
- add Hydra token introspection validation in `trust-registry` for trust runtime reads
- enforce the governed internal scope `trust.runtime.read`
- keep verifier evaluation dependent only on the existing narrow trust-read client boundary
- add a shared `phase1sql` CLI with versioned SQL assets for:
  - `migrate up`
  - `bootstrap trust --file <path>`
- remove startup-time schema mutation as the primary SQL lifecycle on the primary path
- keep JSON runtime clearly transitional and non-primary
- add the minimum validation-path hardening needed only if required to make `bash scripts/validate.sh` truthful and reproducible

## Out of scope

- public Phase 1 issuer or verifier contract changes
- broader production auth rollout for issuer or verifier public APIs
- wallet flows, proof verification, selective disclosure, chain anchoring, or cross-vertical behavior
- broader trust-registry product APIs or trust write administration beyond the already-landed bounded bootstrap model
- platform-wide migration framework rollout beyond the Phase 1 shared SQL lifecycle
- JWT/JWKS local verification, mTLS-only identity, or forwarded edge-auth context in place of Hydra introspection

## Affected files, services, or packages

- `docs/plans/active/0013-phase1-hydra-internal-trust-auth-and-phase1sql-lifecycle.md`
- `docs/plans/archive/0012-phase1-trust-registry-writes-bootstrap-and-internal-auth.md`
- `docs/adr/0010-phase1-internal-trust-service-identity-and-sql-lifecycle.md`
- `docs/threat-model/full/0003-phase1-kyc-issuance-verification-and-auth.md`
- `services/internal/phase1sql/`
- `services/trust-registry/internal/config/`
- `services/trust-registry/internal/httpapi/`
- `services/trust-registry/internal/phase1/`
- `services/verifier-api/internal/config/`
- `services/verifier-api/internal/httpapi/`
- `services/verifier-api/internal/phase1/`
- `scripts/validate-web.sh` only if the validation hardening remains strictly scoped and low-risk

## Assumptions

- ADR 0010 is the governing source for the internal trust runtime service-identity and SQL lifecycle model
- Hydra client-credentials is the only internal trust runtime auth mechanism targeted in this slice
- trust-registry runtime reads remain bounded to issuer trust state, allowed template identifiers, and verification-key references
- `services/internal/phase1sql` remains the schema owner for the current approved Phase 1 relational table set
- public Phase 1 contracts remain unchanged

## Risks

- Hydra introspection outages will fail trust runtime reads closed until later resilience work is governed
- startup removal of implicit schema mutation can create operational failures if migration ordering is not explicit and tested
- validation hardening can accidentally hide real source-level casing errors if it edits more than ignored generated artifacts or validation orchestration
- leaving both static bearer tokens and Hydra client-credentials active beyond a short migration window would blur the trust boundary
- `bash scripts/validate.sh` is currently vulnerable to TS1149 due workspace-fanout typecheck plus ignored `*.tsbuildinfo` path-casing contamination between `/mnt/c/dev/HDIP` and `/mnt/c/dev/hdip`

## Validation steps

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `npm run schema:validate`
- `bash scripts/validate.sh`

## Rollback or containment notes

If the Hydra trust auth or explicit SQL lifecycle is incorrect, revert both together.
Do not keep a half-migrated state where runtime trust reads use the new Hydra path but the primary SQL lifecycle still depends on implicit startup schema mutation, or vice versa.

## Open questions

- whether a later broader internal service-identity rollout should reuse the same Hydra introspection pattern for other resource-server boundaries
- whether the transitional JSON fallback should be removed immediately after the explicit `phase1sql` lifecycle stabilizes or in a later persistence-hardening slice
