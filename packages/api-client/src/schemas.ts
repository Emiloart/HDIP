import { z } from "zod";

export const errorEnvelopeSchema = z.object({
  error: z.object({
    code: z.string().min(1),
    message: z.string().min(1),
    requestId: z.string().min(1).optional(),
  }),
});

export const healthResponseSchema = z.object({
  status: z.enum(["ok", "ready"]),
  service: z.string().min(1),
  version: z.string().min(1),
});

export const credentialTemplateMetadataSchema = z.object({
  templateId: z.string().min(1),
  displayName: z.string().min(1),
  version: z.string().min(1),
  credentialTypes: z.array(z.string().min(1)).min(1),
});

export const issuerProfileSchema = z.object({
  issuerId: z.string().min(1),
  displayName: z.string().min(1),
  endpoint: z.string().url(),
  supportedCredentialTemplates: z.array(z.string().min(1)),
});

export const verifierPolicyRequestSchema = z.object({
  requestId: z.string().min(1),
  purpose: z.string().min(1),
  requiredPredicates: z.array(z.string().min(1)).min(1),
});

export const verifierResultSchema = z.object({
  requestId: z.string().min(1),
  decision: z.enum(["allow", "deny", "review"]),
  reasons: z.array(z.string().min(1)),
});

export type ErrorEnvelope = z.infer<typeof errorEnvelopeSchema>;
export type HealthResponse = z.infer<typeof healthResponseSchema>;
export type CredentialTemplateMetadata = z.infer<typeof credentialTemplateMetadataSchema>;
export type IssuerProfile = z.infer<typeof issuerProfileSchema>;
export type VerifierPolicyRequest = z.infer<typeof verifierPolicyRequestSchema>;
export type VerifierResult = z.infer<typeof verifierResultSchema>;
