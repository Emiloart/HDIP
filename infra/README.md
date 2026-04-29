# Infra

Infrastructure topology, environment definitions, and deployment assets for HDIP.
Operational convenience must not silently redefine application trust logic.

## Phase 1

`infra/phase1/docker-compose.yml` is the local packaging path for the first fintech/exchange reusable-KYC loop.
It runs one PostgreSQL instance, one Hydra instance, explicit `phase1sql` migration/bootstrap jobs, and the three Phase 1 services.

Use `docs/integration/quickstart.md` for the integrator-facing walkthrough.
