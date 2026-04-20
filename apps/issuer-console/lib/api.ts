import { createIssuerApiClient, createServiceClient } from "@hdip/api-client";

import { getIssuerConsoleEnv } from "./env";

const env = getIssuerConsoleEnv();

export const issuerApi = createIssuerApiClient(env.NEXT_PUBLIC_ISSUER_API_BASE_URL);
export const trustRegistryApi = createServiceClient(env.NEXT_PUBLIC_TRUST_REGISTRY_BASE_URL);
