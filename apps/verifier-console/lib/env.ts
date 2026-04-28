import { z } from "zod";

const verifierConsoleEnvSchema = z.object({
  NEXT_PUBLIC_VERIFIER_API_BASE_URL: z.string().url().default("http://127.0.0.1:8082"),
  NEXT_PUBLIC_TRUST_REGISTRY_BASE_URL: z.string().url().default("http://127.0.0.1:8083"),
  NEXT_PUBLIC_VERIFIER_INTEGRATOR_PRINCIPAL_ID: z
    .string()
    .min(1)
    .default("verifier_integrator_exchange"),
  NEXT_PUBLIC_VERIFIER_INTEGRATOR_ORGANIZATION_ID: z
    .string()
    .min(1)
    .default("verifier_org_exchange"),
  NEXT_PUBLIC_VERIFIER_INTEGRATOR_AUTH_REFERENCE: z
    .string()
    .min(1)
    .default("local-verifier-console"),
  NEXT_PUBLIC_VERIFIER_INTEGRATOR_SCOPES: z
    .string()
    .min(1)
    .default("verifier.requests.create, verifier.results.read"),
});

export function getVerifierConsoleEnv() {
  return verifierConsoleEnvSchema.parse(process.env);
}
