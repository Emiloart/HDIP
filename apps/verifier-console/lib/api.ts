import { createServiceClient, createVerifierApiClient } from "@hdip/api-client";

import { getVerifierConsoleEnv } from "./env";

const env = getVerifierConsoleEnv();

const verifierIntegratorHeaders = {
  "X-HDIP-Principal-ID": env.NEXT_PUBLIC_VERIFIER_INTEGRATOR_PRINCIPAL_ID,
  "X-HDIP-Organization-ID": env.NEXT_PUBLIC_VERIFIER_INTEGRATOR_ORGANIZATION_ID,
  "X-HDIP-Auth-Reference": env.NEXT_PUBLIC_VERIFIER_INTEGRATOR_AUTH_REFERENCE,
  "X-HDIP-Scopes": env.NEXT_PUBLIC_VERIFIER_INTEGRATOR_SCOPES,
};

export const verifierApi = createVerifierApiClient(env.NEXT_PUBLIC_VERIFIER_API_BASE_URL, {
  defaultHeaders: verifierIntegratorHeaders,
});
export const trustRegistryApi = createServiceClient(env.NEXT_PUBLIC_TRUST_REGISTRY_BASE_URL);
