import { createServiceClient } from "@hdip/api-client";

import { getVerifierConsoleEnv } from "./env";

const env = getVerifierConsoleEnv();

export const verifierApi = createServiceClient(env.NEXT_PUBLIC_VERIFIER_API_BASE_URL);
export const trustRegistryApi = createServiceClient(env.NEXT_PUBLIC_TRUST_REGISTRY_BASE_URL);
