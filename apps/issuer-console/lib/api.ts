import { createIssuerApiClient, createServiceClient } from "@hdip/api-client";

import { getIssuerConsoleEnv } from "./env";

const env = getIssuerConsoleEnv();

const issuerOperatorHeaders = {
  "X-HDIP-Principal-ID": env.NEXT_PUBLIC_ISSUER_OPERATOR_PRINCIPAL_ID,
  "X-HDIP-Organization-ID": env.NEXT_PUBLIC_ISSUER_OPERATOR_ORGANIZATION_ID,
  "X-HDIP-Auth-Reference": env.NEXT_PUBLIC_ISSUER_OPERATOR_AUTH_REFERENCE,
  "X-HDIP-Scopes": env.NEXT_PUBLIC_ISSUER_OPERATOR_SCOPES,
};

export const issuerApi = createIssuerApiClient(env.NEXT_PUBLIC_ISSUER_API_BASE_URL, {
  defaultHeaders: issuerOperatorHeaders,
});
export const trustRegistryApi = createServiceClient(env.NEXT_PUBLIC_TRUST_REGISTRY_BASE_URL);
