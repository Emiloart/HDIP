# Phase 1 Issuer Console

## Purpose

The Phase 1 issuer console is an internal KYC operations tool.
It is not a wallet, proof engine, or customer-facing identity app.

The console exists to let an approved issuer operator:

- create a reusable KYC credential after the user has already passed an internal KYC process
- retrieve a credential by operational identifier
- inspect status and key timestamps
- revoke or supersede a credential
- hand the temporary Phase 1 credential artifact to the user for verifier onboarding

## User roles

### Issuer operator

The issuer operator is the primary user.
They are bound to exactly one issuer organization for the request, consistent with ADR 0008.

Required capabilities:

- create credential
- read own issuer credentials
- update credential status

Required scopes:

- `issuer.credentials.issue`
- `issuer.credentials.read`
- `issuer.credentials.status.write`

### Issuer reviewer

The reviewer role is a later product refinement.
Phase 1 may show reviewer-oriented status information, but it must not invent a separate authorization model before the auth boundary is governed.

## Core screens

### 1. Credential creation

Primary route:

- `Create credential`

Inputs:

- `templateId`
- `subjectReference`
- `fullLegalName`
- `dateOfBirth`
- `countryOfResidence`
- `documentCountry`
- `kycLevel`
- `verifiedAt`
- `expiresAt`
- optional `Idempotency-Key`

Rules:

- Do not collect raw document numbers, document scans, selfies, liveness media, or free-form KYC evidence in this screen.
- Do not expose `issuerId` as an editable field.
- The issuer organization comes from authenticated service-edge context.
- Show validation errors from the API without leaking raw request bodies into logs or UI telemetry.

Success state:

- show `credentialId`
- show status `active`
- show `expiresAt`
- show `statusReference`
- show the opaque `credentialArtifact`
- provide a copy action for the artifact
- reserve QR rendering for the temporary bridge slice

### 2. Credential lookup

Primary route:

- `Credentials`

Initial lookup modes:

- `credentialId`
- `subjectReference` only after the backend exposes bounded search with audit coverage

Rules:

- A credential detail read is auditable.
- Do not show records outside the authenticated issuer organization.
- Do not render raw opaque artifact values in list views.

Detail view:

- `credentialId`
- `templateId`
- `subjectReference`
- normalized claims
- `artifactDigest`
- current status
- `issuedAt`
- `expiresAt`
- `statusUpdatedAt`
- `supersededByCredentialId` when present

### 3. Status management

Allowed actions:

- revoke active credential
- supersede active credential with `supersededByCredentialId`

Forbidden Phase 1 actions:

- reactivate revoked credential
- mutate normalized claims after issuance
- edit issuer trust state from issuer console
- delete audit records

Confirmation requirements:

- status changes require a confirmation dialog
- confirmation copy must name the credential and resulting terminal state
- the UI must show API failure states clearly

### 4. Operational audit view

Phase 1 does not require a full audit browser.
The console should still surface enough local context after actions:

- request ID
- resource ID
- action outcome
- timestamp

Do not display broad audit logs or raw audit payloads until a dedicated audit access model exists.

## Temporary user bridge

Phase 1 has no wallet.
The issuer console therefore supports a temporary bridge:

- copy the opaque `credentialArtifact`
- later QR-encode the exact same artifact payload

The bridge must not claim:

- signed credential
- wallet credential
- proof
- selective disclosure
- reusable public identity token

The current trusted verifier input is the opaque `credentialArtifact`.
`credentialId` is useful for support and traceability, but it is not enough by itself under the accepted Phase 1 verifier contract.

## Flow summary

1. Operator completes KYC outside HDIP runtime scope.
2. Operator opens `Create credential`.
3. Operator enters normalized claims only.
4. Console calls `POST /v1/issuer/credentials`.
5. Console displays credential details and artifact transfer controls.
6. User carries artifact to the fintech or exchange.
7. Operator can later lookup the credential and revoke or supersede status.

## Error states

The console must handle:

- unauthenticated
- insufficient scope
- invalid request
- unsupported template
- credential not found
- invalid status transition
- idempotency conflict
- persistence unavailable
- service readiness failure

## UX posture

The console should feel like an operations tool:

- dense forms
- clear status indicators
- restrained visual design
- predictable navigation
- no marketing copy
- no decorative identity claims

## Metrics for pilot readiness

Track operationally, without adding sensitive analytics to identity flows:

- credential creation success rate
- status mutation success rate
- verifier `allow` after issued artifact
- verifier `deny` after revoked credential
- verifier `deny` after suspended issuer
- median time from credential creation to verifier decision

Any analytics beyond bounded operational counters requires privacy review.
