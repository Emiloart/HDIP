# 0001 Platform Bootstrap Threat Model

- Status: accepted
- Date: 2026-04-06
- Owners: repository maintainer

## Change summary

This baseline covers the initial HDIP repository bootstrap, governance spine, and the accepted architecture baseline.
No production runtime exists yet, but the repo now reserves future trust boundaries and development constraints.

## Assets

- holder keys and recovery materials
- issuer keys and signing policy
- verifiable credentials and status information
- trust registry contents
- verifier policy configurations
- audit and compliance records
- operator permissions
- regional routing and data-boundary metadata
- standards and governance artifacts that influence security-sensitive implementation

## Trust boundaries

- holder device boundary
- issuer system boundary
- verifier system boundary
- HDIP control-plane boundary
- Rust crypto core boundary
- admin and operator boundary
- regional data-residency boundary
- edge and API ingress boundary

## Attacker classes

- external attackers seeking credential theft or replay
- malicious or compromised issuers
- abusive verifiers over-collecting identity data
- insiders with excessive operational access
- attackers targeting recovery and account takeover flows
- dependency or supply-chain attackers

## Entry points and privileged actions

- future wallet enrollment
- issuance flows
- presentation flows
- verification APIs
- admin and policy consoles
- recovery ceremonies
- operator tooling

## Abuse and misuse cases

- credential replay across relying parties
- proof forwarding or relay
- recovery flow abuse
- verifier over-collection of attributes
- issuer compromise leading to fraudulent credentials
- admin bypasses or undocumented support actions
- linkability leakage across contexts

## Privacy harms

- correlation of holder activity across verifiers
- over-retention of sensitive data
- regional data leakage
- raw identity data appearing in logs, screenshots, or fixtures

## Mitigations

- governance gates for trust-boundary, privacy, and custody changes
- standards registry enforcing selective disclosure and minimized presentation
- isolation of cryptographic logic in Rust
- repo-level rules forbidding secrets and raw PII in non-production artifacts
- explicit ADR and threat-model requirements for architectural shifts
- regional data-boundary model documented in architecture and privacy docs

## Residual risks

- no runtime enforcement exists yet for product behavior
- hook coverage is intentionally light and advisory at this stage
- secret scanning is heuristic until deeper tooling is introduced
- the repo is still located under `/mnt/c/...` during bootstrap

## Validation impact

Bootstrap validation currently covers governance structure and heuristic secret scanning only.

## Related ADRs, plans, PRs, and issues

- `docs/adr/0001-governance-spine-and-bootstrap.md`
- `docs/adr/0002-hdip-initial-platform-baseline.md`
- `docs/plans/active/0001-governance-and-foundation-bootstrap.md`
