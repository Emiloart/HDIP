# Phase 1 Partner Provisioning Runbook

## Purpose

Provision controlled Phase 1 issuer and verifier access for a fintech/exchange pilot using Ory Hydra client credentials.

This is an operator workflow, not self-service onboarding.
Do not create API keys or trusted-header partner access.

## Prerequisites

- Hydra admin endpoint reachable from the operator environment.
- `hydra` CLI installed and authenticated for the target Hydra admin endpoint.
- `python3` available for generated client secrets and JSON output.
- Partner approval recorded outside this repo.
- Secret manager destination ready before generating credentials.

Set:

```bash
export HYDRA_ADMIN_URL="<private-hydra-admin-url>"
```

## Provision An Issuer Client

Issuer `client_id` must be the issuer organization identifier used by HDIP trust state.
For DID-style issuer identifiers, use `client_secret_post`.

```bash
bash scripts/phase1-provision-client.sh create issuer \
  --client-id "did:web:issuer.example.com"
```

The script prints JSON containing:

- `clientId`
- `clientSecret`
- `grantType`
- `tokenEndpointAuthMethod`
- canonical issuer scopes

Store the `clientSecret` immediately in the deployment secret manager.
Do not paste it into docs, tickets, chat, screenshots, shell history, or repository files.

## Provision A Verifier Client

Verifier `client_id` must identify the verifier organization.

```bash
bash scripts/phase1-provision-client.sh create verifier \
  --client-id "verifier_org_example"
```

Verifier clients receive only:

- `verifier.requests.create`
- `verifier.results.read`

## Emergency Revocation

Delete the Hydra client if partner credentials are compromised or access must be revoked immediately:

```bash
bash scripts/phase1-provision-client.sh delete \
  --client-id "verifier_org_example"
```

Deletion invalidates future token issuance for that client.
Previously issued access tokens remain governed by Hydra token lifetime and issuer/verifier introspection behavior.
Stop public ingress if immediate containment is required.

## Rotation

Phase 1 rotation is delete plus recreate:

1. create a replacement client
2. deliver the new secret through the approved secret channel
3. verify the partner can obtain a token
4. delete the old client
5. run the public-auth smoke test if the client is part of the pilot path

## Validation

After provisioning, run:

```bash
bash scripts/phase1-public-auth-smoke.sh
```

Use environment overrides when testing non-local clients:

```bash
export ISSUER_CLIENT_ID="<issuer-client-id>"
export ISSUER_CLIENT_SECRET="<issuer-client-secret>"
export VERIFIER_CLIENT_ID="<verifier-client-id>"
export VERIFIER_CLIENT_SECRET="<verifier-client-secret>"
export HYDRA_PUBLIC_URL="<hydra-public-url>"
export ISSUER_API_BASE_URL="<issuer-api-url>"
export VERIFIER_API_BASE_URL="<verifier-api-url>"

bash scripts/phase1-public-auth-smoke.sh
```

## Safety Rules

- Use Hydra client credentials only.
- Use canonical Phase 1 scopes only.
- Never grant issuer scopes to verifier clients.
- Never grant verifier scopes to issuer clients unless a later ADR explicitly changes this.
- Never use `X-HDIP-*` headers for partner access.
- Never log client secrets, bearer tokens, opaque artifacts, normalized KYC claims, or raw request bodies.
