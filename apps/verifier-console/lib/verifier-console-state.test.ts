import { describe, expect, it, vi } from "vitest";

import {
  createVerificationRequest,
  defaultVerifyCredentialFormState,
  idempotencyKey,
  parseCredentialArtifact,
  serviceErrorMessage,
} from "./verifier-console-state";

describe("verifier console state helpers", () => {
  it("creates a Phase 1 verification request from artifact JSON", () => {
    const request = createVerificationRequest({
      ...defaultVerifyCredentialFormState(),
      credentialId: " cred_hdip_passport_basic_001 ",
      credentialArtifact: JSON.stringify({
        kind: "phase1_opaque_artifact",
        mediaType: "application/vnd.hdip.phase1-opaque-artifact",
        value: "opaque-artifact:v1:credential",
      }),
    });

    expect(request).toEqual({
      policyId: "kyc-passport-basic",
      credentialId: "cred_hdip_passport_basic_001",
      credentialArtifact: {
        kind: "phase1_opaque_artifact",
        mediaType: "application/vnd.hdip.phase1-opaque-artifact",
        value: "opaque-artifact:v1:credential",
      },
    });
  });

  it("accepts a raw opaque artifact value", () => {
    expect(parseCredentialArtifact(" opaque-artifact:v1:credential ")).toEqual({
      kind: "phase1_opaque_artifact",
      mediaType: "application/vnd.hdip.phase1-opaque-artifact",
      value: "opaque-artifact:v1:credential",
    });
  });

  it("rejects credential-id-only verification attempts", () => {
    expect(() =>
      createVerificationRequest({
        credentialId: "cred_hdip_passport_basic_001",
        credentialArtifact: "",
      }),
    ).toThrow("credentialArtifact is required");
  });

  it("rejects malformed artifact JSON", () => {
    expect(() => parseCredentialArtifact('{"kind":"wrong"}')).toThrow(
      "credentialArtifact JSON must match",
    );
  });

  it("formats typed service errors and generates idempotency keys", () => {
    vi.spyOn(globalThis.crypto, "randomUUID").mockReturnValue("uuid-1");

    expect(idempotencyKey("verifier-create")).toBe("verifier-create-uuid-1");
    expect(serviceErrorMessage({ code: "credential_not_found", message: "missing" })).toBe(
      "credential_not_found: missing",
    );
  });
});
