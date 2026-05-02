# Phase 1 Vault Secrets Runbook

## Purpose

Store Phase 1 pilot deployment and partner secrets in Vault KV v2.

Real secrets must not be committed to the repo, pasted into normal chat, included in screenshots, or stored in shell history.

## Required Inputs

- Vault address
- Vault auth method for deployment operators
- KV v2 mount available at `kv/`
- operator policy with least-privilege access to Phase 1 pilot paths
- secure partner credential delivery channel

## Required Paths

Use these KV v2 paths for the first pilot:

| Path | Purpose |
| --- | --- |
| `kv/hdip/phase1/pilot/database` | SQL password and Phase 1 database DSN inputs |
| `kv/hdip/phase1/pilot/hydra` | Hydra system secret, public URL, admin URL, and OAuth bootstrap inputs |
| `kv/hdip/phase1/pilot/services` | issuer/verifier/trust-registry introspection and runtime client secrets |
| `kv/hdip/phase1/pilot/cloudflare` | Cloudflare zone ID and deployment automation token if used |
| `kv/hdip/phase1/pilot/vm` | VM hostname, operator SSH notes, and non-secret deployment metadata |
| `kv/hdip/phase1/pilot/partners/<partner-id>` | public issuer or verifier client ID, generated secret, type, scopes, and approval reference |

Do not store KYC claims, opaque credential artifacts, bearer tokens, or raw request bodies in Vault.

## Minimum Secret Set

Database:

- `postgres_password`
- `phase1_database_url` if using an external managed database

Hydra:

- `hydra_system_secret`
- `hydra_public_url`
- `hydra_admin_url`
- public issuer/verifier client IDs and generated secrets when seeded through Compose

Services:

- issuer API introspection client ID and secret
- verifier API introspection client ID and secret
- verifier runtime trust client ID and secret
- trust-registry introspection client ID and secret

Partners:

- `client_type`: `issuer` or `verifier`
- `client_id`
- `client_secret`
- canonical scopes
- approval reference
- delivery reference
- created timestamp
- rotation or revocation notes

## Secret Write Procedure

Prefer the Vault UI, an approved secrets broker, or a local file with restrictive permissions.
Avoid putting secrets directly on shell command lines.

If a temporary file is used:

```bash
umask 077
cat > /tmp/hdip-phase1-services.json <<'JSON'
{
  "issuer_api_introspection_client_id": "<issuer-api-client-id>",
  "issuer_api_introspection_client_secret": "<issuer-api-client-secret>",
  "verifier_api_introspection_client_id": "<verifier-api-client-id>",
  "verifier_api_introspection_client_secret": "<verifier-api-client-secret>",
  "verifier_runtime_client_id": "<verifier-runtime-client-id>",
  "verifier_runtime_client_secret": "<verifier-runtime-client-secret>",
  "trust_registry_introspection_client_id": "<trust-registry-client-id>",
  "trust_registry_introspection_client_secret": "<trust-registry-client-secret>"
}
JSON

vault kv put kv/hdip/phase1/pilot/services @/tmp/hdip-phase1-services.json
shred -u /tmp/hdip-phase1-services.json 2>/dev/null || rm -f /tmp/hdip-phase1-services.json
```

## Partner Client Storage

After provisioning a partner client:

```bash
HYDRA_ADMIN_URL="<private-hydra-admin-url>" \
  bash scripts/phase1-provision-client.sh create verifier \
  --client-id "<verifier-org-id>" > /tmp/hdip-partner-client.json
```

Immediately move the generated secret into Vault:

```bash
vault kv put kv/hdip/phase1/pilot/partners/<partner-id> @/tmp/hdip-partner-client.json
shred -u /tmp/hdip-partner-client.json 2>/dev/null || rm -f /tmp/hdip-partner-client.json
```

Do not leave the generated JSON in `/tmp`, terminal scrollback, tickets, chat, or screenshots.

## Access Policy

Minimum policy shape:

- deployment operators can read deployment/service paths
- partner onboarding operators can create and read only partner paths they manage
- incident operators can read partner paths and delete Hydra clients
- application services do not need broad Vault access unless a later deployment automation explicitly injects secrets at runtime

Do not grant wildcard read access to all partners unless it is required for an incident role.

## Rotation And Emergency Revocation

Rotation:

1. create a replacement Hydra client
2. write the replacement client secret to Vault
3. deliver the new secret through the approved secure channel
4. confirm token acquisition
5. delete the old Hydra client
6. run `scripts/phase1-public-auth-smoke.sh` when the rotated client participates in the pilot loop

Emergency revocation:

```bash
HYDRA_ADMIN_URL="<private-hydra-admin-url>" \
  bash scripts/phase1-provision-client.sh delete \
  --client-id "<compromised-client-id>"
```

Then stop public ingress if active bearer tokens may still be usable before expiry.

## Validation

Before pilot access:

- Vault paths exist
- Vault audit logging is enabled
- operator access is role-bounded
- no real secret appears in repo files
- no real secret appears in Cloudflare, reverse-proxy, service, or support logs
- public-auth smoke passes using secrets loaded from Vault
