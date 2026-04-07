import { z } from "zod";

const issuerConsoleEnvSchema = z.object({
  NEXT_PUBLIC_ISSUER_API_BASE_URL: z.string().url().default("http://127.0.0.1:8081"),
  NEXT_PUBLIC_TRUST_REGISTRY_BASE_URL: z.string().url().default("http://127.0.0.1:8083"),
});

export function getIssuerConsoleEnv() {
  return issuerConsoleEnvSchema.parse(process.env);
}
