import type { ZodType } from "zod";

import {
  credentialTemplateMetadataSchema,
  errorEnvelopeSchema,
  healthResponseSchema,
  issuerProfileSchema,
  verifierPolicyRequestSchema,
  verifierResultSchema,
} from "./schemas";

export class ServiceClientError extends Error {
  readonly code: string;
  readonly requestId?: string;
  readonly status: number;

  constructor(
    code: string,
    message: string,
    status: number,
    requestId?: string,
    options?: { cause?: unknown },
  ) {
    super(message, options);
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

function encodePathSegment(value: string) {
  return encodeURIComponent(value);
}

function parseJsonBody(rawBody: string, status: number) {
  if (rawBody.trim() === "") {
    return null;
  }

  try {
    return JSON.parse(rawBody) as unknown;
  } catch (error) {
    throw new ServiceClientError(
      "invalid_json",
      "response body was not valid JSON",
      status,
      undefined,
      { cause: error },
    );
  }
}

function parseTypedPayload<T>(payload: unknown, schema: ZodType<T>, status: number): T {
  const parsedPayload = schema.safeParse(payload);
  if (!parsedPayload.success) {
    throw new ServiceClientError(
      "invalid_response",
      "response payload did not match the expected schema",
      status,
      undefined,
      { cause: parsedPayload.error },
    );
  }

  return parsedPayload.data;
}

async function fetchWithSchema<T>(url: string, schema: ZodType<T>): Promise<T> {
  let response: Response;
  try {
    response = await fetch(url, {
      headers: {
        Accept: "application/json",
      },
    });
  } catch (error) {
    throw new ServiceClientError(
      "network_error",
      "request could not be completed",
      0,
      undefined,
      { cause: error },
    );
  }

  const rawBody = await response.text();
  const payload = parseJsonBody(rawBody, response.status);

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

  return parseTypedPayload(payload, schema, response.status);
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

export function createIssuerApiClient(baseUrl: string) {
  const service = createServiceClient(baseUrl);

  return {
    ...service,
    profile() {
      return fetchWithSchema(buildServiceUrl(baseUrl, "/v1/issuer/profile"), issuerProfileSchema);
    },
    template(templateId: string) {
      return fetchWithSchema(
        buildServiceUrl(baseUrl, `/v1/issuer/templates/${encodePathSegment(templateId)}`),
        credentialTemplateMetadataSchema,
      );
    },
  };
}

export function createVerifierApiClient(baseUrl: string) {
  const service = createServiceClient(baseUrl);

  return {
    ...service,
    policyRequest(policyId: string) {
      return fetchWithSchema(
        buildServiceUrl(baseUrl, `/v1/verifier/policy-requests/${encodePathSegment(policyId)}`),
        verifierPolicyRequestSchema,
      );
    },
    stubResult(requestId: string) {
      return fetchWithSchema(
        buildServiceUrl(baseUrl, `/v1/verifier/results/${encodePathSegment(requestId)}/stub`),
        verifierResultSchema,
      );
    },
  };
}
