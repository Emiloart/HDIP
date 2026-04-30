# Phase 1 Verifier SDK And API Guide

## Audience

This guide is for fintech, exchange, payment, and neobank backend engineers integrating HDIP as a reusable KYC verification check.

The SDK and API are server-side only.
Do not call verifier endpoints directly from browser or mobile clients.

## Integration promise

If a user already has an active HDIP Phase 1 KYC credential from a trusted issuer, an integrated verifier can make one backend call and receive a deterministic onboarding decision.

The Phase 1 promise is:

- no repeated document upload for the verifier
- no wallet dependency
- no proof verification dependency
- deterministic `allow` or `deny` for the narrow reusable KYC case

## Current contract reality

Current Phase 1 verification requires an opaque `credentialArtifact`.

`credentialId` may be supplied for traceability, but it is not the trusted verifier input by itself.
A credential-ID-only verifier flow requires a later governed resolver or token bridge.

## Verification API

Endpoint:

```http
POST /v1/verifier/verifications
```

Required headers:

```http
Content-Type: application/json
Accept: application/json
Authorization: Bearer <verifier-access-token>
Idempotency-Key: <unique-operation-key>
```

Public Phase 1 verifier auth uses Hydra OAuth2 client credentials and bearer tokens.
The OAuth `client_id` is the verifier organization identifier used for attribution.
Local process-run sandbox automation may still use deprecated header attribution, but packaged/pilot flows must use bearer tokens.

Request:

```json
{
  "policyId": "kyc-passport-basic",
  "credentialId": "cred_hdip_passport_basic_001",
  "credentialArtifact": {
    "kind": "phase1_opaque_artifact",
    "mediaType": "application/vnd.hdip.phase1-opaque-artifact",
    "value": "opaque-artifact:v1:..."
  }
}
```

Response:

```json
{
  "verificationId": "ver_000001",
  "credentialId": "cred_hdip_passport_basic_001",
  "issuerId": "did:web:issuer.hdip.dev",
  "decision": "allow",
  "reasonCodes": ["credential_active"],
  "evaluatedAt": "2026-04-28T12:00:00Z",
  "credentialStatus": "active"
}
```

## Decision semantics

`allow` means:

- credential resolved
- credential status is `active`
- credential is not expired
- issuer trust state is active
- template/policy compatibility passed

`deny` means one of the required Phase 1 checks failed.
Examples:

- credential not found
- credential revoked
- credential superseded
- credential expired
- issuer suspended
- issuer not trusted
- template not allowed

`review` is reserved for future product policy.
Suspended or unknown issuers must return `deny`, not `review`.

## Result retrieval

Endpoint:

```http
GET /v1/verifier/verifications/{verificationId}
```

Use this for retry-safe result lookup and operational support.
Do not poll indefinitely; partner systems should persist the `verificationId` returned by the create call.

## TypeScript SDK shape

The first SDK should expose a deliberately small server-side API:

```ts
type VerifyCredentialInput = {
  policyId: string;
  credentialArtifact: {
    kind: "phase1_opaque_artifact";
    mediaType: "application/vnd.hdip.phase1-opaque-artifact";
    value: string;
  };
  credentialId?: string;
  idempotencyKey: string;
};

type HDIPVerifierClient = {
  verifyCredential(input: VerifyCredentialInput): Promise<VerificationResult>;
  getVerification(verificationId: string): Promise<VerificationResult>;
};
```

SDK rules:

- require an explicit base URL
- require caller-provided server-side credentials
- require idempotency key for writes
- validate response payloads against the canonical TypeScript schemas
- wrap non-2xx responses in typed errors
- never log request bodies or opaque artifacts

## cURL example

```bash
curl -sS https://verifier-api.example.com/v1/verifier/verifications \
  -H "Authorization: Bearer $HDIP_VERIFIER_TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -H "Idempotency-Key: onboarding-user-123-20260428" \
  -d '{
    "policyId": "kyc-passport-basic",
    "credentialId": "cred_hdip_passport_basic_001",
    "credentialArtifact": {
      "kind": "phase1_opaque_artifact",
      "mediaType": "application/vnd.hdip.phase1-opaque-artifact",
      "value": "opaque-artifact:v1:..."
    }
  }'
```

## Partner backend flow

1. User chooses `Verify with HDIP`.
2. Partner backend receives the Phase 1 artifact from the user bridge.
3. Partner backend generates an idempotency key scoped to its onboarding attempt.
4. Partner backend calls `POST /v1/verifier/verifications`.
5. Partner stores `verificationId`, `decision`, `reasonCodes`, and `evaluatedAt`.
6. `allow` continues onboarding.
7. `deny` routes to manual KYC or rejection according to partner policy.

## Sandbox runbook

Use `docs/runbooks/phase1-sandbox.md` for the first local integration loop:

- create a credential in issuer console
- copy the verifier transfer payload
- paste it into verifier console
- confirm `allow`
- revoke the credential
- verify again and confirm `deny`

The verifier transfer payload is a console bridge around the existing `credentialArtifact` contract.
It is not a new public API contract and must not be treated as a signed credential.

## Errors

The SDK must expose at least:

- `unauthenticated`
- `insufficient_scope`
- `invalid_request`
- `credential_not_found`
- `idempotency_conflict`
- `persistence_error`
- `trust_runtime_unavailable`
- `invalid_response`
- `network_error`

Error payloads follow `schemas/json/common/error-envelope.schema.json`.

## Security rules for integrators

- call HDIP only from backend services
- keep verifier credentials out of browsers, mobile apps, logs, support tools, and analytics
- use one idempotency key per onboarding attempt
- store only the verification result fields required for onboarding audit
- do not store opaque artifacts unless required for a bounded retry window
- do not treat the Phase 1 artifact as a signed credential or proof

## First developer portal content

The developer portal should publish:

- quickstart
- Hydra client-credentials authentication setup
- cURL verification example
- TypeScript SDK example
- error reference
- idempotency guide
- decision semantics
- privacy and data-minimization guidance
- sandbox artifact fixture

## Deferred

- public webhook delivery
- SDK-managed QR or short-token resolver
- wallet presentation
- selective disclosure
- proof verification
- mobile SDK
- API-key provisioning UI
