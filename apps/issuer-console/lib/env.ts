import { z } from "zod";

const issuerConsoleEnvSchema = z.object({
  NEXT_PUBLIC_ISSUER_API_BASE_URL: z.string().url().default("http://127.0.0.1:8081"),
  NEXT_PUBLIC_TRUST_REGISTRY_BASE_URL: z.string().url().default("http://127.0.0.1:8083"),
  NEXT_PUBLIC_ISSUER_OPERATOR_PRINCIPAL_ID: z.string().min(1).default("issuer_operator_alex"),
  NEXT_PUBLIC_ISSUER_OPERATOR_ORGANIZATION_ID: z.string().min(1).default("did:web:issuer.hdip.dev"),
  NEXT_PUBLIC_ISSUER_OPERATOR_AUTH_REFERENCE: z.string().min(1).default("local-issuer-console"),
  NEXT_PUBLIC_ISSUER_OPERATOR_SCOPES: z
    .string()
    .min(1)
    .default(
      "issuer.credentials.issue, issuer.credentials.read, issuer.credentials.status.write",
    ),
});

export function getIssuerConsoleEnv() {
  return issuerConsoleEnvSchema.parse(process.env);
}
