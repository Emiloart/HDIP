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
