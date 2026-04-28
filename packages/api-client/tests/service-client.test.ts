import { afterEach, describe, expect, it, vi } from "vitest";

import {
  buildServiceUrl,
  createIssuerApiClient,
  createServiceClient,
  createVerifierApiClient,
  ServiceClientError,
} from "../src/service-client";

describe("buildServiceUrl", () => {
  it("joins urls without double slashes", () => {
    expect(buildServiceUrl("http://127.0.0.1:8081/", "/healthz")).toBe(
      "http://127.0.0.1:8081/healthz",
    );
  });
});

describe("createServiceClient", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("surfaces structured error envelopes", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: {
              code: "route_not_found",
              message: "route not found",
              requestId: "req-1",
            },
          }),
          {
            status: 404,
            headers: { "Content-Type": "application/json" },
          },
        ),
      ),
    );

    const client = createServiceClient("http://127.0.0.1:8082");

    await expect(client.health()).rejects.toMatchObject({
      name: "ServiceClientError",
      code: "route_not_found",
      message: "route not found",
      status: 404,
      requestId: "req-1",
    });
  });

  it("wraps malformed json responses in a typed client error", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response("<html>gateway failure</html>", {
          status: 502,
          headers: { "Content-Type": "text/html" },
        }),
      ),
    );

    const client = createServiceClient("http://127.0.0.1:8082");

    await expect(client.health()).rejects.toMatchObject({
      name: "ServiceClientError",
      code: "invalid_json",
      message: "response body was not valid JSON",
      status: 502,
    });
  });

  it("wraps schema mismatches in a typed client error", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            status: "ok",
            service: "verifier-api",
          }),
          {
            status: 200,
            headers: { "Content-Type": "application/json" },
          },
        ),
      ),
    );

    const client = createServiceClient("http://127.0.0.1:8082");

    await expect(client.health()).rejects.toMatchObject({
      name: "ServiceClientError",
      code: "invalid_response",
      message: "response payload did not match the expected schema",
      status: 200,
    });
  });

  it("wraps network failures in a typed client error", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new TypeError("network down")));

    const client = createServiceClient("http://127.0.0.1:8082");

    await expect(client.health()).rejects.toMatchObject({
      name: "ServiceClientError",
      code: "network_error",
      message: "request could not be completed",
      status: 0,
    });
  });

  it("fetches issuer profile through the typed issuer client", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            issuerId: "did:web:issuer.hdip.dev",
            displayName: "HDIP Passport Issuer",
            endpoint: "http://127.0.0.1:8081",
            supportedCredentialTemplates: ["hdip-passport-basic"],
          }),
          {
            status: 200,
            headers: { "Content-Type": "application/json" },
          },
        ),
      ),
    );

    const client = createIssuerApiClient("http://127.0.0.1:8081");

    await expect(client.profile()).resolves.toMatchObject({
      issuerId: "did:web:issuer.hdip.dev",
      supportedCredentialTemplates: ["hdip-passport-basic"],
    });
  });

  it("creates issuer credentials with attribution and idempotency headers", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          credentialId: "cred_hdip_passport_basic_001",
          issuerId: "did:web:issuer.hdip.dev",
          templateId: "hdip-passport-basic",
          status: "active",
          issuedAt: "2026-04-28T10:15:00Z",
          expiresAt: "2027-04-28T10:15:00Z",
          statusReference: "status:cred_hdip_passport_basic_001",
          credentialArtifact: {
            kind: "phase1_opaque_artifact",
            mediaType: "application/vnd.hdip.phase1-opaque-artifact",
            value: "opaque-artifact:v1:cred_hdip_passport_basic_001",
          },
        }),
        {
          status: 201,
          headers: { "Content-Type": "application/json" },
        },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    const client = createIssuerApiClient("http://127.0.0.1:8081", {
      defaultHeaders: {
        "X-HDIP-Principal-ID": "issuer_operator_alex",
      },
    });

    await expect(client.issueCredential(
      {
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
      },
      { idempotencyKey: "issue-1" },
    )).resolves.toMatchObject({
      credentialId: "cred_hdip_passport_basic_001",
      status: "active",
    });

    const [, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    const headers = new Headers(init.headers);
    expect(init.method).toBe("POST");
    expect(headers.get("Idempotency-Key")).toBe("issue-1");
    expect(headers.get("X-HDIP-Principal-ID")).toBe("issuer_operator_alex");
  });

  it("fetches credentials and updates credential status through the issuer client", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            credentialId: "cred_hdip_passport_basic_001",
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
            statusReference: "status:cred_hdip_passport_basic_001",
            issuedAt: "2026-04-28T10:15:00Z",
            expiresAt: "2027-04-28T10:15:00Z",
            statusUpdatedAt: "2026-04-28T10:15:00Z",
            credentialArtifact: {
              kind: "phase1_opaque_artifact",
              mediaType: "application/vnd.hdip.phase1-opaque-artifact",
              value: "opaque-artifact:v1:cred_hdip_passport_basic_001",
            },
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            credentialId: "cred_hdip_passport_basic_001",
            status: "revoked",
            statusReference: "status:cred_hdip_passport_basic_001",
            statusUpdatedAt: "2026-04-28T10:20:00Z",
            expiresAt: "2027-04-28T10:15:00Z",
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        ),
      );
    vi.stubGlobal("fetch", fetchMock);

    const client = createIssuerApiClient("http://127.0.0.1:8081");

    await expect(client.credential("cred_hdip_passport_basic_001")).resolves.toMatchObject({
      status: "active",
    });
    await expect(client.updateCredentialStatus(
      "cred_hdip_passport_basic_001",
      { status: "revoked" },
      { idempotencyKey: "status-1" },
    )).resolves.toMatchObject({
      status: "revoked",
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://127.0.0.1:8081/v1/issuer/credentials/cred_hdip_passport_basic_001",
      expect.objectContaining({ headers: expect.any(Headers) }),
    );
    expect(fetchMock).toHaveBeenLastCalledWith(
      "http://127.0.0.1:8081/v1/issuer/credentials/cred_hdip_passport_basic_001/status",
      expect.objectContaining({ method: "POST" }),
    );
  });

  it("fetches verifier stub results through the typed verifier client", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            requestId: "kyc-passport-basic-review",
            decision: "allow",
            reasons: [
              "stub flow matched the expected issuer profile",
              "stub flow matched the HDIP passport template contract",
            ],
          }),
          {
            status: 200,
            headers: { "Content-Type": "application/json" },
          },
        ),
      ),
    );

    const client = createVerifierApiClient("http://127.0.0.1:8082");

    await expect(client.stubResult("kyc-passport-basic-review")).resolves.toMatchObject({
      requestId: "kyc-passport-basic-review",
      decision: "allow",
    });
  });
});
