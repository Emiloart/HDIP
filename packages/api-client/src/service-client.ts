import type { ZodType } from "zod";

import {
  credentialRecordSchema,
  credentialTemplateMetadataSchema,
  credentialStatusSchema,
  credentialStatusUpdateRequestSchema,
  errorEnvelopeSchema,
  healthResponseSchema,
  issuanceRequestSchema,
  issuanceResponseSchema,
  issuerProfileSchema,
  verificationResultSchema,
  verificationSubmissionRequestSchema,
  verifierPolicyRequestSchema,
  verifierResultSchema,
} from "./schemas";
import type {
  CredentialStatusUpdateRequest,
  IssuanceRequest,
  VerificationSubmissionRequest,
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

type ServiceClientOptions = {
  defaultHeaders?: HeadersInit;
};

type RequestOptions = {
  headers?: HeadersInit;
  idempotencyKey?: string;
};

function mergeHeaders(...headerSets: Array<HeadersInit | undefined>) {
  const headers = new Headers();

  for (const headerSet of headerSets) {
    if (headerSet === undefined) {
      continue;
    }

    const nextHeaders = new Headers(headerSet);
    nextHeaders.forEach((value, key) => {
      headers.set(key, value);
    });
  }

  return headers;
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

async function fetchWithSchema<T>(
  url: string,
  schema: ZodType<T>,
  init?: RequestInit,
): Promise<T> {
  let response: Response;
  try {
    response = await fetch(url, init);
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

async function requestWithSchema<T>(
  baseUrl: string,
  path: string,
  schema: ZodType<T>,
  options: ServiceClientOptions = {},
  init: RequestInit = {},
) {
  const headers = mergeHeaders(
    { Accept: "application/json" },
    options.defaultHeaders,
    init.headers,
  );

  return fetchWithSchema(buildServiceUrl(baseUrl, path), schema, {
    ...init,
    headers,
  });
}

function jsonRequestInit(
  method: "POST",
  payload: unknown,
  requestOptions?: RequestOptions,
) {
  const headers = mergeHeaders(
    { "Content-Type": "application/json" },
    requestOptions?.headers,
  );

  if (requestOptions?.idempotencyKey !== undefined) {
    headers.set("Idempotency-Key", requestOptions.idempotencyKey);
  }

  return {
    method,
    headers,
    body: JSON.stringify(payload),
  };
}

export function createServiceClient(baseUrl: string, options: ServiceClientOptions = {}) {
  return {
    baseUrl,
    health() {
      return requestWithSchema(baseUrl, "/healthz", healthResponseSchema, options);
    },
    readiness() {
      return requestWithSchema(baseUrl, "/readyz", healthResponseSchema, options);
    },
  };
}

export function createIssuerApiClient(baseUrl: string, options: ServiceClientOptions = {}) {
  const service = createServiceClient(baseUrl, options);

  return {
    ...service,
    profile() {
      return requestWithSchema(baseUrl, "/v1/issuer/profile", issuerProfileSchema, options);
    },
    template(templateId: string) {
      return requestWithSchema(
        baseUrl,
        `/v1/issuer/templates/${encodePathSegment(templateId)}`,
        credentialTemplateMetadataSchema,
        options,
      );
    },
    issueCredential(request: IssuanceRequest, requestOptions?: RequestOptions) {
      const parsedRequest = issuanceRequestSchema.parse(request);
      return requestWithSchema(
        baseUrl,
        "/v1/issuer/credentials",
        issuanceResponseSchema,
        options,
        jsonRequestInit("POST", parsedRequest, requestOptions),
      );
    },
    credential(credentialId: string) {
      return requestWithSchema(
        baseUrl,
        `/v1/issuer/credentials/${encodePathSegment(credentialId)}`,
        credentialRecordSchema,
        options,
      );
    },
    updateCredentialStatus(
      credentialId: string,
      request: CredentialStatusUpdateRequest,
      requestOptions?: RequestOptions,
    ) {
      const parsedRequest = credentialStatusUpdateRequestSchema.parse(request);
      return requestWithSchema(
        baseUrl,
        `/v1/issuer/credentials/${encodePathSegment(credentialId)}/status`,
        credentialStatusSchema,
        options,
        jsonRequestInit("POST", parsedRequest, requestOptions),
      );
    },
  };
}

export function createVerifierApiClient(baseUrl: string, options: ServiceClientOptions = {}) {
  const service = createServiceClient(baseUrl, options);

  return {
    ...service,
    policyRequest(policyId: string) {
      return requestWithSchema(
        baseUrl,
        `/v1/verifier/policy-requests/${encodePathSegment(policyId)}`,
        verifierPolicyRequestSchema,
        options,
      );
    },
    stubResult(requestId: string) {
      return requestWithSchema(
        baseUrl,
        `/v1/verifier/results/${encodePathSegment(requestId)}/stub`,
        verifierResultSchema,
        options,
      );
    },
    verifyCredential(request: VerificationSubmissionRequest, requestOptions?: RequestOptions) {
      const parsedRequest = verificationSubmissionRequestSchema.parse(request);
      return requestWithSchema(
        baseUrl,
        "/v1/verifier/verifications",
        verificationResultSchema,
        options,
        jsonRequestInit("POST", parsedRequest, requestOptions),
      );
    },
    verification(verificationId: string) {
      return requestWithSchema(
        baseUrl,
        `/v1/verifier/verifications/${encodePathSegment(verificationId)}`,
        verificationResultSchema,
        options,
      );
    },
  };
}
