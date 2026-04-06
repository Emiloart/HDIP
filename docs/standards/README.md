# Standards Registry

## Purpose

This document records the standards, profiles, and interoperability constraints that HDIP treats as approved, targeted, deferred, or forbidden.
Approved standards docs have higher precedence than general prose.

## Approved

- W3C Verifiable Credentials Data Model 2.0
- DID Core-compatible identifiers
- OpenID4VCI
- OpenID4VP
- WebAuthn and passkeys
- VC Data Integrity
- Bitstring Status List
- SD-JWT VC as the default mainstream credential profile

## Approved for advanced privacy-sensitive flows

- BBS-based selective disclosure within a VC Data Integrity model

## Targeted

- Digital Credentials API compatibility for browser-mediated wallet flows
- ISO mdoc or related government-grade credential compatibility once implementation scope justifies it

## Deferred

- optional chain anchoring for integrity checkpoints and trust registry notarization
- any ecosystem-specific trust registry federation beyond the initial HDIP control plane

## Forbidden without explicit ADR

- storing raw PII or raw credentials on-chain
- proprietary credential formats when an approved standard profile fits
- undocumented verifier-specific shortcuts that bypass approved issuance or presentation protocols
- undocumented custodial behavior that weakens holder control

## Terminology

- `holder`: the user-controlled wallet side of a credential flow
- `issuer`: the party that signs and issues credentials
- `verifier`: the relying party requesting proof or presentation
- `status`: revocation or current-validity information for a credential
- `selective disclosure`: revealing only the minimum required attributes or predicates
