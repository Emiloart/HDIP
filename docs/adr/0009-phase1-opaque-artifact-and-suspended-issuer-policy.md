# 0009 Phase 1 Opaque Artifact And Suspended Issuer Policy

- Status: accepted
- Date: 2026-04-20
- Owners: repository maintainer

## Context

The accepted Phase 1 design baseline in ADR 0006 and ADR 0007 left two implementation-blocking ambiguities:

- the Phase 1 issuance and verification contracts still used "signed credential artifact" wording even though no accepted ADR has yet locked a real signing or proof-bearing issuance model for the deterministic Phase 1 slices
- issuer suspension handling in verifier decisions remained open between `deny` and `review`

Those ambiguities now block the next real Phase 1 service-logic slice.
HDIP needs an unambiguous repo truth that preserves security posture without pretending that the current deterministic implementation has production cryptographic issuance.

## Decision

### Supersession scope

This ADR partially supersedes the earlier accepted wording in:

- `docs/adr/0006-phase1-credential-and-issuance-boundary.md`
- `docs/adr/0007-phase1-state-and-persistence-model.md`

Specifically, it replaces the ambiguous "signed credential artifact" language for deterministic Phase 1 implementation slices and resolves the previously open suspended-issuer decision policy.

### Phase 1 credential artifact materialization semantics

For the deterministic Phase 1 implementation slices, the issued credential artifact is an HDIP-controlled opaque credential artifact.

It is:

- non-production
- non-cryptographic
- non-proof-bearing
- issuer-produced
- usable only for Phase 1 state continuity and verifier API interoperability

It must not be described in contracts, schemas, examples, or service code as:

- a signed credential
- a signed VC
- a signed proof
- a verifiable proof artifact
- a cryptographically verifiable presentation artifact

The deterministic Phase 1 contract term is `credentialArtifact`.

The Phase 1 artifact may be carried inline or by reference, but when carried inline its shape is limited to an opaque artifact object with bounded metadata and opaque value content.
The verifier treats this artifact as an opaque submitted artifact only.
Verifier logic in Phase 1 may use:

- contract validity
- bounded artifact continuity checks
- artifact digest continuity
- issuer trust state
- template compatibility
- credential status
- expiry

Verifier logic in Phase 1 must not treat the artifact as cryptographically verifiable.

### Deferred signing and proof-bearing semantics

Real signing, key materialization, VC proof-bearing semantics, and cryptographically verifiable artifact semantics are explicitly deferred to a later ADR and a later implementation slice.

No deterministic Phase 1 implementation may imply that opaque artifacts are security-equivalent to real signed credentials.

### Suspended and untrusted issuer verifier policy

In the deterministic Phase 1 verifier flow, issuer trust is security-first and narrow:

- issuer trust state `active` may continue evaluation
- issuer trust state `suspended` must return verifier decision `deny`
- missing, unknown, or otherwise non-active trust state must return verifier decision `deny`

This is the required deterministic Phase 1 policy.
`review` is not used for suspended or untrusted issuers in Phase 1.

The verifier result must include an explicit auditable reason code:

- `issuer_suspended` when the issuer trust state is explicitly suspended
- `issuer_not_trusted` when the issuer trust state is missing, unknown, or otherwise not active

This decision does not broaden trust policy beyond the minimum Phase 1 trust boundary.

## Alternatives considered

### Keeping "signed credential artifact" language and treating it as a harmless placeholder

Rejected because it leaves the repo truth internally contradictory and encourages fake cryptographic semantics.

### Returning `review` for suspended issuers

Rejected because the deterministic Phase 1 verifier flow is meant to be security-first and auditable.
Suspended or non-active trust state is not safe enough to continue evaluation in the first real Phase 1 implementation.

### Defining a minimal signing scheme now

Rejected because no accepted ADR has yet locked the signing, key-materialization, or verification model needed to make such a change defensible.

## Security impact

Positive.
This ADR removes fake-signature ambiguity and forces a clear `deny` policy for suspended or untrusted issuers.

## Privacy impact

Positive.
The deterministic artifact remains bounded and opaque rather than encouraging expanded proof payloads or raw credential duplication.

## Migration / rollback

Later ADRs may replace the opaque Phase 1 artifact with real signed or proof-bearing semantics, but only through explicit architectural decision and contract updates.
Later ADRs may also refine trust-policy outcomes beyond `deny` for non-active issuers, but only explicitly and with matching threat-model updates.

Do not silently reinterpret `credentialArtifact` as a signed or cryptographically verifiable object.

## Consequences

- Phase 1 schemas, examples, typed contracts, and boundary-layer code must use neutral opaque-artifact naming
- the next deterministic service-logic slice can proceed without inventing fake signing
- verifier decision behavior for suspended or non-active issuers is now fixed and auditable
- existing stub metadata endpoints remain unchanged and are still not authoritative Phase 1 issuance or verification APIs

## Open questions

- whether a later signed-artifact ADR should replace inline opaque artifact values with issuer-managed artifact references by default
- whether future post-Phase-1 trust policy should distinguish additional non-active issuer states beyond `active` and `suspended`

## Related plans, PRs, and issues

- `docs/plans/active/0006-phase1-kyc-credential-and-verifier-api.md`
- `docs/adr/0006-phase1-credential-and-issuance-boundary.md`
- `docs/adr/0007-phase1-state-and-persistence-model.md`
- `docs/adr/0008-phase1-auth-and-attribution-boundary.md`
- `docs/threat-model/full/0003-phase1-kyc-issuance-verification-and-auth.md`
