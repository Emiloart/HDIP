import { z } from "zod";

const isoCountryCodeSchema = z.string().regex(/^[A-Z]{2}$/);
const sha256DigestSchema = z.string().regex(/^[a-f0-9]{64}$/);
const credentialStatusValueSchema = z.enum(["active", "revoked", "superseded"]);
const verificationCredentialStatusValueSchema = z.enum([
  "active",
  "revoked",
  "superseded",
  "expired",
]);
const signedCredentialSchema = z.object({
  format: z.literal("sd_jwt_vc"),
  mediaType: z.literal("application/vc+sd-jwt"),
  value: z.string().min(1),
}).strict();
const kycClaimsSchema = z.object({
  fullLegalName: z.string().min(1),
  dateOfBirth: z.iso.date(),
  countryOfResidence: isoCountryCodeSchema,
  documentCountry: isoCountryCodeSchema,
  kycLevel: z.string().min(1),
  verifiedAt: z.iso.datetime({ offset: true }),
  expiresAt: z.iso.datetime({ offset: true }),
}).strict();

export const errorEnvelopeSchema = z.object({
  error: z.object({
    code: z.string().min(1),
    message: z.string().min(1),
    requestId: z.string().min(1).optional(),
  }).strict(),
}).strict();

export const healthResponseSchema = z.object({
  status: z.enum(["ok", "ready"]),
  service: z.string().min(1),
  version: z.string().min(1),
}).strict();

export const credentialTemplateMetadataSchema = z.object({
  templateId: z.string().min(1),
  displayName: z.string().min(1),
  version: z.string().min(1),
  credentialTypes: z.array(z.string().min(1)).min(1),
}).strict();

export const issuerProfileSchema = z.object({
  issuerId: z.string().min(1),
  displayName: z.string().min(1),
  endpoint: z.string().url(),
  supportedCredentialTemplates: z.array(z.string().min(1)),
}).strict();

export const verifierPolicyRequestSchema = z.object({
  requestId: z.string().min(1),
  purpose: z.string().min(1),
  requiredPredicates: z.array(z.string().min(1)).min(1),
}).strict();

export const verifierResultSchema = z.object({
  requestId: z.string().min(1),
  decision: z.enum(["allow", "deny", "review"]),
  reasons: z.array(z.string().min(1)),
}).strict();

export const issuanceRequestSchema = z.object({
  templateId: z.string().min(1),
  subjectReference: z.string().min(1),
  claims: kycClaimsSchema,
}).strict();

export const issuanceResponseSchema = z.object({
  credentialId: z.string().min(1),
  issuerId: z.string().min(1),
  templateId: z.string().min(1),
  status: credentialStatusValueSchema,
  issuedAt: z.iso.datetime({ offset: true }),
  expiresAt: z.iso.datetime({ offset: true }),
  statusReference: z.string().min(1),
  signedCredential: signedCredentialSchema,
}).strict();

export const credentialStatusSchema = z.object({
  credentialId: z.string().min(1),
  status: credentialStatusValueSchema,
  statusReference: z.string().min(1),
  statusUpdatedAt: z.iso.datetime({ offset: true }),
  expiresAt: z.iso.datetime({ offset: true }),
  supersededByCredentialId: z.string().min(1).optional(),
}).strict();

export const credentialRecordSchema = z
  .object({
    credentialId: z.string().min(1),
    issuerId: z.string().min(1),
    templateId: z.string().min(1),
    subjectReference: z.string().min(1),
    claims: kycClaimsSchema,
    artifactDigest: sha256DigestSchema,
    status: credentialStatusValueSchema,
    statusReference: z.string().min(1),
    issuedAt: z.iso.datetime({ offset: true }),
    expiresAt: z.iso.datetime({ offset: true }),
    statusUpdatedAt: z.iso.datetime({ offset: true }),
    supersededByCredentialId: z.string().min(1).optional(),
    signedCredential: signedCredentialSchema.optional(),
    artifactReference: z.string().min(1).optional(),
  })
  .strict()
  .refine(
    (value) =>
      (value.signedCredential !== undefined && value.artifactReference === undefined) ||
      (value.signedCredential === undefined && value.artifactReference !== undefined),
    {
      message: "credential record must include either signedCredential or artifactReference",
      path: ["signedCredential"],
    },
  );

export const verificationSubmissionRequestSchema = z.object({
  policyId: z.string().min(1),
  credentialId: z.string().min(1).optional(),
  signedCredential: signedCredentialSchema,
}).strict();

export const verificationResultSchema = z.object({
  verificationId: z.string().min(1),
  credentialId: z.string().min(1).optional(),
  issuerId: z.string().min(1),
  decision: z.enum(["allow", "deny", "review"]),
  reasonCodes: z.array(z.string().min(1)).min(1),
  evaluatedAt: z.iso.datetime({ offset: true }),
  credentialStatus: verificationCredentialStatusValueSchema,
}).strict();

export const auditRecordSchema = z.object({
  auditId: z.string().min(1),
  actor: z.object({
    principalId: z.string().min(1),
    organizationId: z.string().min(1),
    actorType: z.enum(["issuer_operator", "verifier_integrator"]),
    authenticationReference: z.string().min(1),
  }).strict(),
  action: z.string().min(1),
  resourceType: z.string().min(1),
  resourceId: z.string().min(1),
  requestId: z.string().min(1),
  idempotencyKey: z.string().min(1).optional(),
  outcome: z.enum(["succeeded", "denied", "failed"]),
  occurredAt: z.iso.datetime({ offset: true }),
  serviceName: z.string().min(1),
}).strict();

export type ErrorEnvelope = z.infer<typeof errorEnvelopeSchema>;
export type HealthResponse = z.infer<typeof healthResponseSchema>;
export type CredentialTemplateMetadata = z.infer<typeof credentialTemplateMetadataSchema>;
export type IssuerProfile = z.infer<typeof issuerProfileSchema>;
export type VerifierPolicyRequest = z.infer<typeof verifierPolicyRequestSchema>;
export type VerifierResult = z.infer<typeof verifierResultSchema>;
export type IssuanceRequest = z.infer<typeof issuanceRequestSchema>;
export type IssuanceResponse = z.infer<typeof issuanceResponseSchema>;
export type CredentialStatus = z.infer<typeof credentialStatusSchema>;
export type CredentialRecord = z.infer<typeof credentialRecordSchema>;
export type VerificationSubmissionRequest = z.infer<typeof verificationSubmissionRequestSchema>;
export type VerificationResult = z.infer<typeof verificationResultSchema>;
export type AuditRecord = z.infer<typeof auditRecordSchema>;
