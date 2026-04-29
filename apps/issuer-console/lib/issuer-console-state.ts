import type {
  CredentialRecord,
  CredentialStatusUpdateRequest,
  IssuanceResponse,
  IssuanceRequest,
  ServiceClientError,
} from "@hdip/api-client";

export type CreateCredentialFormState = {
  templateId: string;
  subjectReference: string;
  fullLegalName: string;
  dateOfBirth: string;
  countryOfResidence: string;
  documentCountry: string;
  kycLevel: string;
  verifiedAt: string;
  expiresAt: string;
};

export type RecentCredential = {
  credentialId: string;
  status: string;
  expiresAt: string;
};

export type StatusAction = "revoked" | "superseded";

export const recentCredentialStorageKey = "hdip.issuer-console.recent-credentials";
export const verifierTransferPayloadKind = "hdip_phase1_verifier_transfer";

export type VerifierTransferPayload = {
  kind: typeof verifierTransferPayloadKind;
  credentialId: string;
  credentialArtifact: IssuanceResponse["credentialArtifact"];
};

export function defaultCreateCredentialFormState(): CreateCredentialFormState {
  const now = new Date();
  const nextYear = new Date(now);
  nextYear.setUTCFullYear(now.getUTCFullYear() + 1);

  return {
    templateId: "hdip-passport-basic",
    subjectReference: "",
    fullLegalName: "",
    dateOfBirth: "",
    countryOfResidence: "",
    documentCountry: "",
    kycLevel: "basic",
    verifiedAt: toDateTimeLocalValue(now),
    expiresAt: toDateTimeLocalValue(nextYear),
  };
}

export function createIssuanceRequest(form: CreateCredentialFormState): IssuanceRequest {
  return {
    templateId: form.templateId.trim(),
    subjectReference: form.subjectReference.trim(),
    claims: {
      fullLegalName: form.fullLegalName.trim(),
      dateOfBirth: form.dateOfBirth,
      countryOfResidence: form.countryOfResidence.trim().toUpperCase(),
      documentCountry: form.documentCountry.trim().toUpperCase(),
      kycLevel: form.kycLevel.trim(),
      verifiedAt: toIsoDateTime(form.verifiedAt),
      expiresAt: toIsoDateTime(form.expiresAt),
    },
  };
}

export function credentialStatusUpdateRequest(
  action: StatusAction,
  supersededByCredentialId: string,
): CredentialStatusUpdateRequest {
  if (action === "revoked") {
    return { status: "revoked" };
  }

  return {
    status: "superseded",
    supersededByCredentialId: supersededByCredentialId.trim(),
  };
}

export function availableStatusActions(record: CredentialRecord): StatusAction[] {
  if (record.status !== "active") {
    return [];
  }

  return ["revoked", "superseded"];
}

export function toRecentCredential(record: Pick<CredentialRecord, "credentialId" | "status" | "expiresAt">): RecentCredential {
  return {
    credentialId: record.credentialId,
    status: record.status,
    expiresAt: record.expiresAt,
  };
}

export function mergeRecentCredentials(
  existing: RecentCredential[],
  next: RecentCredential,
): RecentCredential[] {
  return [
    next,
    ...existing.filter((credential) => credential.credentialId !== next.credentialId),
  ].slice(0, 8);
}

export function verifierTransferPayload(
  record: Pick<IssuanceResponseLike, "credentialId" | "credentialArtifact">,
): VerifierTransferPayload {
  return {
    kind: verifierTransferPayloadKind,
    credentialId: record.credentialId,
    credentialArtifact: record.credentialArtifact,
  };
}

export function stringifyVerifierTransferPayload(
  record: Pick<IssuanceResponseLike, "credentialId" | "credentialArtifact">,
) {
  return JSON.stringify(verifierTransferPayload(record), null, 2);
}

export function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("en", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

export function idempotencyKey(prefix: string) {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return `${prefix}-${crypto.randomUUID()}`;
  }

  return `${prefix}-${Date.now()}`;
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

  return "Unexpected issuer console error";
}

function toIsoDateTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toISOString();
}

function toDateTimeLocalValue(value: Date) {
  const offsetMs = value.getTimezoneOffset() * 60_000;
  return new Date(value.getTime() - offsetMs).toISOString().slice(0, 16);
}

type IssuanceResponseLike = {
  credentialId: string;
  credentialArtifact: IssuanceResponse["credentialArtifact"];
};
