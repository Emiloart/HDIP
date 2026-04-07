import { z } from "zod";

const verifierConsoleEnvSchema = z.object({
  NEXT_PUBLIC_VERIFIER_API_BASE_URL: z.string().url().default("http://127.0.0.1:8082"),
  NEXT_PUBLIC_TRUST_REGISTRY_BASE_URL: z.string().url().default("http://127.0.0.1:8083"),
});

export function getVerifierConsoleEnv() {
  return verifierConsoleEnvSchema.parse(process.env);
}
