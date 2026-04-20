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
