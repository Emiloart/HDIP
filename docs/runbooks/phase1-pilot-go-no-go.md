# Phase 1 Pilot Go/No-Go Checklist

## Purpose

This checklist decides whether HDIP Phase 1 is ready for a controlled fintech/exchange pilot.

Approval means ready for 1-2 controlled partners.
It does not mean self-service onboarding or global production scale.

## Required Decisions

- pilot domain selected
- VM provider and public IP selected
- Cloudflare zone active
- Vault address and auth method selected
- first issuer organization ID selected
- first verifier organization ID selected
- secure partner credential delivery channel selected

## Go Criteria

### Edge

- Cloudflare DNS records exist only for approved public/protected hostnames.
- Cloudflare SSL/TLS mode is Full Strict.
- HTTPS redirects are enabled.
- WAF managed rules are enabled.
- Request body limits are configured.
- Rate limits exist for Hydra token, issuer write, issuer status, verifier create, and verifier read paths.
- Hydra admin is not publicly reachable.
- SQL is not publicly reachable.
- `trust-registry` is not publicly reachable.

### Secrets

- Vault KV v2 is enabled.
- Required Vault paths are populated.
- Real secrets are absent from repo files, screenshots, docs, tickets, and chat.
- Partner client secrets are stored under partner-specific Vault paths.
- Vault audit logging is enabled.
- Operator access is role-bounded.

### Runtime

- VM firewall exposes only SSH from operators and HTTP/HTTPS through the approved edge path.
- Compose stack starts from `/etc/hdip/phase1.env`, not `infra/phase1/.env.example`.
- SQL migration has run.
- Trust bootstrap has run.
- Hydra issuer and verifier clients are provisioned with canonical scopes only.
- `issuer-api /readyz` passes.
- `verifier-api /readyz` passes.
- `trust-registry /readyz` passes privately.

### Product Loop

- Public-auth smoke passes.
- Verifier token cannot issue credentials.
- Invalid token fails closed.
- Issued active credential verifies `allow`.
- Revoked credential verifies `deny`.
- Suspended issuer verifies `deny` if issuer suspension is part of the pilot scenario.
- Audit records exist for issuance, status mutation, and verification.

### Privacy And Logging

- Service logs contain request IDs but not bearer tokens.
- Service logs do not contain Hydra client secrets.
- Service logs do not contain opaque credential artifacts.
- Service logs do not contain normalized KYC claims.
- Cloudflare and reverse-proxy logs do not store raw request bodies.
- Partner docs instruct server-side verifier use only.

## No-Go Conditions

Do not start the pilot if any of these are true:

- any private surface is publicly reachable
- header auth is enabled in production
- public-auth smoke fails
- Vault paths contain real KYC data or opaque artifacts
- secrets are stored in repo files or normal chat
- Cloudflare rate limits are missing
- partner clients have scopes outside the canonical issuer or verifier scope sets
- rollback owner and procedure are unclear

## Sign-Off Record

Record sign-off outside this repo with:

- pilot domain
- deployed commit SHA
- VM provider and region
- Vault path prefix
- issuer client ID
- verifier client ID
- Cloudflare zone
- public-auth smoke timestamp
- operator sign-off
- known residual risks
