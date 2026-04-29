import { describe, expect, it, vi } from "vitest";

import {
  availableStatusActions,
  createIssuanceRequest,
  credentialStatusUpdateRequest,
  defaultCreateCredentialFormState,
  idempotencyKey,
  mergeRecentCredentials,
  serviceErrorMessage,
  stringifyVerifierTransferPayload,
  verifierTransferPayload,
} from "./issuer-console-state";

describe("issuer console state helpers", () => {
  it("builds the Phase 1 issuance request from form input", () => {
    const request = createIssuanceRequest({
      ...defaultCreateCredentialFormState(),
      subjectReference: " subject_123 ",
      fullLegalName: " Ada Lovelace ",
      dateOfBirth: "1990-01-02",
      countryOfResidence: "ng",
      documentCountry: "gb",
      verifiedAt: "2026-04-28T10:15",
      expiresAt: "2027-04-28T10:15",
    });

    expect(request).toMatchObject({
      templateId: "hdip-passport-basic",
      subjectReference: "subject_123",
      claims: {
        fullLegalName: "Ada Lovelace",
        dateOfBirth: "1990-01-02",
        countryOfResidence: "NG",
        documentCountry: "GB",
        kycLevel: "basic",
      },
    });
    expect(request.claims.verifiedAt).toContain("2026-04-28T");
    expect(request.claims.expiresAt).toContain("2027-04-28T");
  });

  it("only exposes revoke and supersede for active credentials", () => {
    expect(availableStatusActions({
      credentialId: "cred_1",
      issuerId: "did:web:issuer.hdip.dev",
      templateId: "hdip-passport-basic",
      subjectReference: "subject_1",
      claims: {
        fullLegalName: "Ada Lovelace",
        dateOfBirth: "1990-01-02",
        countryOfResidence: "NG",
        documentCountry: "NG",
        kycLevel: "basic",
        verifiedAt: "2026-04-28T10:15:00Z",
        expiresAt: "2027-04-28T10:15:00Z",
      },
      artifactDigest: "a".repeat(64),
      status: "active",
      statusReference: "status:cred_1",
      issuedAt: "2026-04-28T10:15:00Z",
      expiresAt: "2027-04-28T10:15:00Z",
      statusUpdatedAt: "2026-04-28T10:15:00Z",
      credentialArtifact: {
        kind: "phase1_opaque_artifact",
        mediaType: "application/vnd.hdip.phase1-opaque-artifact",
        value: "opaque-artifact:v1:cred_1",
      },
    })).toEqual(["revoked", "superseded"]);
  });

  it("does not expose status actions for terminal credentials", () => {
    const record = {
      credentialId: "cred_1",
      issuerId: "did:web:issuer.hdip.dev",
      templateId: "hdip-passport-basic",
      subjectReference: "subject_1",
      claims: {
        fullLegalName: "Ada Lovelace",
        dateOfBirth: "1990-01-02",
        countryOfResidence: "NG",
        documentCountry: "NG",
        kycLevel: "basic",
        verifiedAt: "2026-04-28T10:15:00Z",
        expiresAt: "2027-04-28T10:15:00Z",
      },
      artifactDigest: "a".repeat(64),
      status: "revoked" as const,
      statusReference: "status:cred_1",
      issuedAt: "2026-04-28T10:15:00Z",
      expiresAt: "2027-04-28T10:15:00Z",
      statusUpdatedAt: "2026-04-28T10:15:00Z",
      credentialArtifact: {
        kind: "phase1_opaque_artifact" as const,
        mediaType: "application/vnd.hdip.phase1-opaque-artifact" as const,
        value: "opaque-artifact:v1:cred_1",
      },
    };

    expect(availableStatusActions(record)).toEqual([]);
  });

  it("builds the allowed Phase 1 status update payloads", () => {
    expect(credentialStatusUpdateRequest("revoked", "ignored")).toEqual({
      status: "revoked",
    });
    expect(credentialStatusUpdateRequest("superseded", " cred_2 ")).toEqual({
      status: "superseded",
      supersededByCredentialId: "cred_2",
    });
  });

  it("keeps recent credentials unique and bounded", () => {
    const existing = Array.from({ length: 8 }, (_, index) => ({
      credentialId: `cred_${index}`,
      status: "active",
      expiresAt: "2027-04-28T10:15:00Z",
    }));

    const merged = mergeRecentCredentials(existing, {
      credentialId: "cred_3",
      status: "revoked",
      expiresAt: "2027-04-28T10:15:00Z",
    });

    expect(merged).toHaveLength(8);
    expect(merged[0]).toMatchObject({ credentialId: "cred_3", status: "revoked" });
    expect(merged.filter((credential) => credential.credentialId === "cred_3")).toHaveLength(1);
  });

  it("builds the temporary verifier transfer payload without changing contracts", () => {
    const credential = {
      credentialId: "cred_hdip_passport_basic_001",
      credentialArtifact: {
        kind: "phase1_opaque_artifact" as const,
        mediaType: "application/vnd.hdip.phase1-opaque-artifact" as const,
        value: "opaque-artifact:v1:credential",
      },
    };

    expect(verifierTransferPayload(credential)).toEqual({
      kind: "hdip_phase1_verifier_transfer",
      credentialId: "cred_hdip_passport_basic_001",
      credentialArtifact: credential.credentialArtifact,
    });
    expect(JSON.parse(stringifyVerifierTransferPayload(credential))).toMatchObject({
      kind: "hdip_phase1_verifier_transfer",
      credentialId: "cred_hdip_passport_basic_001",
    });
  });

  it("formats typed service errors and generates deterministic-key prefixes", () => {
    vi.spyOn(globalThis.crypto, "randomUUID").mockReturnValue("uuid-1");

    expect(idempotencyKey("issuer-create")).toBe("issuer-create-uuid-1");
    expect(serviceErrorMessage({ code: "invalid_request", message: "bad payload" })).toBe(
      "invalid_request: bad payload",
    );
  });
});
