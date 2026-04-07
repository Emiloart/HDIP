import { afterEach, describe, expect, it, vi } from "vitest";

import { buildServiceUrl, createServiceClient, ServiceClientError } from "../src/service-client";

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
});
