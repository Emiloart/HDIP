# Phase 1 Cloudflare Edge Runbook

## Purpose

Configure the first Phase 1 pilot edge using Cloudflare without changing issuer/verifier service logic.

Cloudflare owns public TLS, WAF, request-size limits, and rate limits.
HDIP services still own auth, attribution, trust decisions, audit, and credential state.

## Required Inputs

- pilot domain or subdomain, for example `<pilot-domain>`
- VM public IP
- Cloudflare account with DNS, SSL/TLS, WAF, Rules, and Rate Limiting permissions
- origin reverse proxy on the VM

## Public Hostnames

Use proxied Cloudflare DNS records for only these public or protected surfaces:

| Hostname | Origin | Access posture |
| --- | --- | --- |
| `issuer.<pilot-domain>` | VM reverse proxy to `issuer-api` | partner-restricted or protected |
| `verifier.<pilot-domain>` | VM reverse proxy to `verifier-api` | partner-facing |
| `auth.<pilot-domain>` | VM reverse proxy to Hydra public port | partner-facing token endpoint only |
| `issuer-console.<pilot-domain>` | VM reverse proxy to issuer console | protected by Cloudflare Access, VPN, or equivalent |
| `docs.<pilot-domain>` | static docs or developer portal | public or protected, no secrets |

Do not create public DNS records for:

- Hydra admin
- SQL/PostgreSQL
- `trust-registry`
- Docker daemon
- VM internal service ports
- migration or bootstrap jobs

## TLS And Origin Rules

Required Cloudflare settings:

- SSL/TLS mode: Full Strict
- Always Use HTTPS: enabled
- automatic HTTP to HTTPS redirect: enabled
- HSTS: enabled only after the pilot hostname is stable
- origin certificate: Cloudflare Origin Certificate or public CA certificate on the VM reverse proxy

The VM reverse proxy must route by hostname and must not expose internal paths by default.

## Allowed Routes

For `issuer.<pilot-domain>` allow only:

- `GET /healthz`
- `GET /readyz`
- `POST /v1/issuer/credentials`
- `GET /v1/issuer/credentials/*`
- `POST /v1/issuer/credentials/*/status`

For `verifier.<pilot-domain>` allow only:

- `GET /healthz`
- `GET /readyz`
- `POST /v1/verifier/verifications`
- `GET /v1/verifier/verifications/*`

For `auth.<pilot-domain>` allow only:

- `POST /oauth2/token`
- Hydra public discovery endpoints only if a partner integration explicitly needs them

Block all other public paths at the reverse proxy or Cloudflare rule layer.

## WAF And Request Limits

Required controls:

- Cloudflare managed WAF rules enabled
- request body size limit set to the smallest plan-supported value that still allows Phase 1 requests
- bot and abuse filtering enabled where available
- block requests with missing or unsupported HTTP methods
- log only metadata required for operations

Do not log:

- `Authorization` headers
- bearer tokens
- Hydra client secrets
- opaque credential artifacts
- normalized KYC claims
- raw request bodies

## Starting Rate Limits

These are first-pilot limits and should be tightened after observing real partner traffic.

| Surface | Starting limit | Action |
| --- | --- | --- |
| `POST /oauth2/token` | 30 requests per minute per IP | block for 10 minutes |
| `POST /v1/issuer/credentials` | 30 requests per minute per IP | block for 10 minutes |
| `POST /v1/issuer/credentials/*/status` | 30 requests per minute per IP | block for 10 minutes |
| `POST /v1/verifier/verifications` | 120 requests per minute per IP | block for 10 minutes |
| `GET /v1/verifier/verifications/*` | 240 requests per minute per IP | block for 10 minutes |

Use per-client limits only if the edge can enforce a stable partner identity without reading, logging, or keying on bearer token values.

## Private Surface Check

Before pilot access, confirm these fail from a public network:

```bash
curl -fsS "https://auth.<pilot-domain>/admin/oauth2/introspect" && exit 1 || true
curl -fsS "https://trust-registry.<pilot-domain>/readyz" && exit 1 || true
nc -vz "<vm-public-ip>" 5432 && exit 1 || true
```

Confirm public health and readiness routes work only for approved public hostnames:

```bash
curl -fsS "https://issuer.<pilot-domain>/readyz"
curl -fsS "https://verifier.<pilot-domain>/readyz"
```

## Rollback

If abuse, leakage, or misrouting is suspected:

1. disable Cloudflare DNS proxy or block the affected route
2. delete compromised Hydra clients if credentials may be exposed
3. preserve SQL, Vault, Cloudflare, and service logs
4. run the public-auth smoke test only after root cause is understood
