# Schemas

Credential schemas, API contracts, and event contracts belong here.
Schemas are contracts and must be versioned and reviewed accordingly.

## Layout

- `schemas/json/` contains the canonical JSON Schemas.
- `schemas/examples/` contains canonical example payloads and validity expectations used by validation and parity tests.

## Current parity scope

The foundation slice currently enforces example-based parity for:

- common health and error envelopes
- issuer profile
- verifier policy request
- verifier result
- credential template metadata
- Phase 1 issuance request and response
- Phase 1 issuer credential status mutation request
- Phase 1 credential record and credential status
- Phase 1 verification submission request and verification result
- Phase 1 audit record

Stub metadata contracts remain separate from the future real Phase 1 write and read contracts.
Deterministic Phase 1 contracts use an opaque `credentialArtifact` model, not a signed or proof-bearing artifact.
