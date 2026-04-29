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
const verifierTransferPayloadKind = "hdip_phase1_verifier_transfer";

export function defaultVerifyCredentialFormState(): VerifyCredentialFormState {
  return {
    credentialId: "",
    credentialArtifact: "",
  };
}

export function createVerificationRequest(
  form: VerifyCredentialFormState,
): VerificationSubmissionRequest {
  const parsedInput = parseVerificationInput(form.credentialArtifact);
  const explicitCredentialId = form.credentialId.trim();
  const payloadCredentialId = parsedInput.credentialId?.trim() ?? "";

  if (
    explicitCredentialId !== "" &&
    payloadCredentialId !== "" &&
    explicitCredentialId !== payloadCredentialId
  ) {
    throw new Error("credentialId does not match the pasted verifier transfer payload");
  }

  const credentialId = explicitCredentialId === "" ? payloadCredentialId : explicitCredentialId;

  return {
    policyId: "kyc-passport-basic",
    ...(credentialId === "" ? {} : { credentialId }),
    credentialArtifact: parsedInput.credentialArtifact,
  };
}

export function parseCredentialArtifact(raw: string): VerificationSubmissionRequest["credentialArtifact"] {
  return parseVerificationInput(raw).credentialArtifact;
}

export function parseVerificationInput(raw: string): {
  credentialArtifact: VerificationSubmissionRequest["credentialArtifact"];
  credentialId?: string;
} {
  const normalized = raw.trim();
  if (normalized === "") {
    throw new Error("credentialArtifact is required by the Phase 1 verifier contract");
  }

  if (normalized.startsWith("{")) {
    const parsed = JSON.parse(normalized) as Record<string, unknown>;
    const directArtifact = parseArtifactObject(parsed);
    if (directArtifact !== null) {
      return {
        credentialArtifact: directArtifact,
      };
    }

    if (parsed.kind === verifierTransferPayloadKind) {
      const credentialArtifact = parseArtifactObject(parsed.credentialArtifact);
      const credentialId = typeof parsed.credentialId === "string" ? parsed.credentialId.trim() : "";
      if (credentialId !== "" && credentialArtifact !== null) {
        return {
          credentialId,
          credentialArtifact,
        };
      }

      throw new Error("verifier transfer payload must include credentialId and credentialArtifact");
    }

    throw new Error(
      "credentialArtifact JSON must match the Phase 1 opaque artifact shape or verifier transfer payload",
    );
  }

  return {
    credentialArtifact: {
      kind: credentialArtifactKind,
      mediaType: credentialArtifactMediaType,
      value: normalized,
    },
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

function parseArtifactObject(
  value: unknown,
): VerificationSubmissionRequest["credentialArtifact"] | null {
  const parsed = value as Partial<VerificationSubmissionRequest["credentialArtifact"]>;
  if (
    parsed !== null &&
    typeof parsed === "object" &&
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

  return null;
}
