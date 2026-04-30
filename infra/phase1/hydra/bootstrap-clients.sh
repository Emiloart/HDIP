#!/bin/sh
set -eu

endpoint="${HYDRA_ADMIN_ENDPOINT:-http://hydra:4445}"

hydra delete oauth2-client "$VERIFIER_TRUST_CLIENT_ID" --endpoint "$endpoint" >/dev/null 2>&1 || true
hydra delete oauth2-client "$TRUST_REGISTRY_INTROSPECTION_CLIENT_ID" --endpoint "$endpoint" >/dev/null 2>&1 || true
hydra delete oauth2-client "$ISSUER_API_INTROSPECTION_CLIENT_ID" --endpoint "$endpoint" >/dev/null 2>&1 || true
hydra delete oauth2-client "$VERIFIER_API_INTROSPECTION_CLIENT_ID" --endpoint "$endpoint" >/dev/null 2>&1 || true
hydra delete oauth2-client "$ISSUER_PUBLIC_CLIENT_ID" --endpoint "$endpoint" >/dev/null 2>&1 || true
hydra delete oauth2-client "$VERIFIER_PUBLIC_CLIENT_ID" --endpoint "$endpoint" >/dev/null 2>&1 || true

hydra create oauth2-client \
  --endpoint "$endpoint" \
  --id "$VERIFIER_TRUST_CLIENT_ID" \
  --secret "$VERIFIER_TRUST_CLIENT_SECRET" \
  --grant-type client_credentials \
  --response-type token \
  --scope "$TRUST_RUNTIME_SCOPE" \
  --token-endpoint-auth-method client_secret_basic \
  --format json >/dev/null

hydra create oauth2-client \
  --endpoint "$endpoint" \
  --id "$TRUST_REGISTRY_INTROSPECTION_CLIENT_ID" \
  --secret "$TRUST_REGISTRY_INTROSPECTION_CLIENT_SECRET" \
  --grant-type client_credentials \
  --response-type token \
  --scope "$TRUST_RUNTIME_SCOPE" \
  --token-endpoint-auth-method client_secret_basic \
  --format json >/dev/null

hydra create oauth2-client \
  --endpoint "$endpoint" \
  --id "$ISSUER_API_INTROSPECTION_CLIENT_ID" \
  --secret "$ISSUER_API_INTROSPECTION_CLIENT_SECRET" \
  --grant-type client_credentials \
  --response-type token \
  --scope "hydra.introspect" \
  --token-endpoint-auth-method client_secret_basic \
  --format json >/dev/null

hydra create oauth2-client \
  --endpoint "$endpoint" \
  --id "$VERIFIER_API_INTROSPECTION_CLIENT_ID" \
  --secret "$VERIFIER_API_INTROSPECTION_CLIENT_SECRET" \
  --grant-type client_credentials \
  --response-type token \
  --scope "hydra.introspect" \
  --token-endpoint-auth-method client_secret_basic \
  --format json >/dev/null

hydra create oauth2-client \
  --endpoint "$endpoint" \
  --id "$ISSUER_PUBLIC_CLIENT_ID" \
  --secret "$ISSUER_PUBLIC_CLIENT_SECRET" \
  --grant-type client_credentials \
  --response-type token \
  --scope "$ISSUER_PUBLIC_SCOPE" \
  --token-endpoint-auth-method client_secret_post \
  --format json >/dev/null

hydra create oauth2-client \
  --endpoint "$endpoint" \
  --id "$VERIFIER_PUBLIC_CLIENT_ID" \
  --secret "$VERIFIER_PUBLIC_CLIENT_SECRET" \
  --grant-type client_credentials \
  --response-type token \
  --scope "$VERIFIER_PUBLIC_SCOPE" \
  --token-endpoint-auth-method client_secret_post \
  --format json >/dev/null

echo "Hydra Phase 1 clients bootstrapped."
