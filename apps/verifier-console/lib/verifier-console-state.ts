import type {
  ServiceClientError,
  VerificationSubmissionRequest,
} from "@hdip/api-client";

export type VerifyCredentialFormState = {
  credentialId: string;
  credentialArtifact: string;
};

const credentialArtifactKind = "phase1_opaque_artifact";
const credentialArtifactMediaType = "application/vnd.hdip.phase1-opaque-artifact";

export function defaultVerifyCredentialFormState(): VerifyCredentialFormState {
  return {
    credentialId: "",
    credentialArtifact: "",
  };
}

export function createVerificationRequest(
  form: VerifyCredentialFormState,
): VerificationSubmissionRequest {
  const artifact = parseCredentialArtifact(form.credentialArtifact);
  const credentialId = form.credentialId.trim();

  return {
    policyId: "kyc-passport-basic",
    ...(credentialId === "" ? {} : { credentialId }),
    credentialArtifact: artifact,
  };
}

export function parseCredentialArtifact(raw: string): VerificationSubmissionRequest["credentialArtifact"] {
  const normalized = raw.trim();
  if (normalized === "") {
    throw new Error("credentialArtifact is required by the Phase 1 verifier contract");
  }

  if (normalized.startsWith("{")) {
    const parsed = JSON.parse(normalized) as Partial<VerificationSubmissionRequest["credentialArtifact"]>;
    if (
      parsed.kind === credentialArtifactKind &&
      parsed.mediaType === credentialArtifactMediaType &&
      typeof parsed.value === "string" &&
      parsed.value.trim() !== ""
    ) {
      return {
        kind: parsed.kind,
        mediaType: parsed.mediaType,
        value: parsed.value.trim(),
      };
    }

    throw new Error("credentialArtifact JSON must match the Phase 1 opaque artifact shape");
  }

  return {
    kind: credentialArtifactKind,
    mediaType: credentialArtifactMediaType,
    value: normalized,
  };
}

export function idempotencyKey(prefix: string) {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return `${prefix}-${crypto.randomUUID()}`;
  }

  return `${prefix}-${Date.now()}`;
}

export function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("en", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

export function serviceErrorMessage(error: unknown) {
  const candidate = error as Partial<ServiceClientError>;
  if (
    candidate !== null &&
    typeof candidate === "object" &&
    typeof candidate.message === "string" &&
    typeof candidate.code === "string"
  ) {
    return `${candidate.code}: ${candidate.message}`;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return "Unexpected verifier console error";
}
