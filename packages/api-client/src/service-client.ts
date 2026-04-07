import type { ZodType } from "zod";

import { errorEnvelopeSchema, healthResponseSchema } from "./schemas";

export class ServiceClientError extends Error {
  readonly code: string;
  readonly requestId?: string;
  readonly status: number;

  constructor(code: string, message: string, status: number, requestId?: string) {
    super(message);
    this.name = "ServiceClientError";
    this.code = code;
    this.status = status;
    if (requestId !== undefined) {
      this.requestId = requestId;
    }
  }
}

export function buildServiceUrl(baseUrl: string, path: string) {
  const normalizedBase = baseUrl.endsWith("/") ? baseUrl.slice(0, -1) : baseUrl;
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `${normalizedBase}${normalizedPath}`;
}

async function fetchWithSchema<T>(url: string, schema: ZodType<T>): Promise<T> {
  const response = await fetch(url, {
    headers: {
      Accept: "application/json",
    },
  });

  const rawBody = await response.text();
  const payload = rawBody === "" ? null : JSON.parse(rawBody);

  if (!response.ok) {
    const parsedError = errorEnvelopeSchema.safeParse(payload);
    if (parsedError.success) {
      throw new ServiceClientError(
        parsedError.data.error.code,
        parsedError.data.error.message,
        response.status,
        parsedError.data.error.requestId,
      );
    }

    throw new ServiceClientError(
      "http_error",
      `request failed with status ${response.status}`,
      response.status,
    );
  }

  return schema.parse(payload);
}

export function createServiceClient(baseUrl: string) {
  return {
    baseUrl,
    health() {
      return fetchWithSchema(buildServiceUrl(baseUrl, "/healthz"), healthResponseSchema);
    },
    readiness() {
      return fetchWithSchema(buildServiceUrl(baseUrl, "/readyz"), healthResponseSchema);
    },
  };
}
