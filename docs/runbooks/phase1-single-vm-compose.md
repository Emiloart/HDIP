# Phase 1 Single VM Compose Runbook

## Purpose

Run the first controlled HDIP Phase 1 pilot on one Linux VM using Docker Compose and Vault-sourced pilot configuration.

This is a first-pilot deployment profile.
It is not the long-term global production architecture.

## Required Inputs

- Linux VM public IP
- Ubuntu 24.04 LTS or equivalent
- minimum 2 vCPU, 4 GB RAM, 40 GB disk
- Docker Engine and Compose v2
- Cloudflare DNS pointing approved hostnames to the VM
- Vault access for pilot secrets
- checked-out HDIP repo at the approved commit

## VM Firewall

Allow only:

- SSH from operator IPs
- HTTP and HTTPS from Cloudflare source ranges, where the provider firewall supports this

Do not expose:

- PostgreSQL `5432`
- Hydra admin `4445`
- service container ports directly
- Docker daemon
- `trust-registry`

## Deployment Layout

Recommended host paths:

```text
/opt/hdip/HDIP                  # repo checkout
/etc/hdip/phase1.env            # rendered env file, mode 0600, not in repo
/var/log/hdip                   # reverse-proxy or host-level logs
```

The env file must be rendered from Vault values.
Do not copy `infra/phase1/.env.example` into pilot use.

Use `infra/phase1/production.env.template` as the checklist for required values.

## Reverse Proxy

Run a host-level reverse proxy such as Caddy or Nginx in front of the Compose services.

Route only:

- `issuer.<pilot-domain>` to issuer API
- `verifier.<pilot-domain>` to verifier API
- `auth.<pilot-domain>` to Hydra public endpoint
- `issuer-console.<pilot-domain>` to issuer console if deployed on the VM

Do not route:

- Hydra admin
- SQL
- `trust-registry`
- migration jobs
- bootstrap jobs

Use Cloudflare Full Strict with an origin certificate on the reverse proxy.

## Deployment Steps

1. Install Docker and Compose v2.
2. Clone or update the repo:

```bash
cd /opt/hdip
git clone https://github.com/Emiloart/HDIP.git
cd /opt/hdip/HDIP
git checkout <approved-commit-sha>
```

3. Render `/etc/hdip/phase1.env` from Vault with mode `0600`.
4. Confirm the env file does not contain local sandbox values from `.env.example`.
5. Start the stack:

```bash
docker compose --env-file /etc/hdip/phase1.env -f infra/phase1/docker-compose.yml up --build -d
```

6. Confirm service health:

```bash
docker compose --env-file /etc/hdip/phase1.env -f infra/phase1/docker-compose.yml ps
curl -fsS http://127.0.0.1:18081/readyz
curl -fsS http://127.0.0.1:18082/readyz
```

7. Confirm public ingress through Cloudflare:

```bash
curl -fsS https://issuer.<pilot-domain>/readyz
curl -fsS https://verifier.<pilot-domain>/readyz
```

8. Run the public-auth smoke test with Vault-sourced partner values:

```bash
export HYDRA_PUBLIC_URL="https://auth.<pilot-domain>"
export ISSUER_API_BASE_URL="https://issuer.<pilot-domain>"
export VERIFIER_API_BASE_URL="https://verifier.<pilot-domain>"
export ISSUER_CLIENT_ID="<issuer-client-id>"
export ISSUER_CLIENT_SECRET="<issuer-client-secret-from-vault>"
export VERIFIER_CLIENT_ID="<verifier-client-id>"
export VERIFIER_CLIENT_SECRET="<verifier-client-secret-from-vault>"

bash scripts/phase1-public-auth-smoke.sh
```

Expected final line:

```text
final status: PASS
```

## Log Safety

Before pilot access, inspect logs for accidental sensitive values:

```bash
docker compose --env-file /etc/hdip/phase1.env -f infra/phase1/docker-compose.yml logs --tail=500
```

Logs must not contain:

- bearer tokens
- Hydra client secrets
- opaque credential artifacts
- normalized KYC claims
- raw request bodies

## Rollback

Rollback order:

1. block Cloudflare ingress for affected hostnames
2. preserve SQL volume and audit logs
3. delete compromised Hydra clients if needed
4. rotate affected Vault secrets if needed
5. roll all service containers back together to the previous approved commit
6. run public-auth smoke before reopening partner access

Do not delete SQL state as a first response unless the pilot is explicitly disposable and audit preservation is not required.
