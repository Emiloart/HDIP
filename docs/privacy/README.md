# Privacy Rules

## Purpose

HDIP makes privacy a product requirement, not a legal afterthought.
This document records baseline privacy constraints for architecture and implementation.

## Core principles

- minimum collection
- minimum storage
- minimum disclosure
- purpose limitation
- role-bounded access
- auditable access to sensitive data
- regional data-boundary awareness
- verifiable alignment between product claims and actual behavior

## Data classes

- public metadata
- low-sensitivity operational metadata
- sensitive personal data
- credential payloads
- recovery material
- audit evidence

## Baseline rules

- New data fields require purpose, access, retention, and deletion rationale.
- Sensitive identity flows must not gain analytics casually.
- Raw credentials and raw PII must not appear in logs, fixtures, screenshots, or examples.
- Prefer proofs and predicates over full disclosure.
- Cross-context correlation must be intentional, documented, and minimized.

## Regional handling

- keep sensitive personal data region-local wherever possible
- replicate only the minimum metadata required for federation and routing
- avoid region-specific forks of core trust logic

## Retention posture

Retention must be explicit by data class and product feature.
Until runtime systems exist, contributors should avoid adding sample data or fixtures that imply unstated retention.
