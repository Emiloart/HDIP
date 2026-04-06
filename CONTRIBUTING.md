# Contributing

## Intent

HDIP is identity infrastructure.
Contributors must optimize for security, privacy, correctness, interoperability, traceability, and reversibility before speed.

## Recommended local environment

- Use WSL on Windows.
- Keep the repository under your Linux home directory, for example `~/code/HDIP`.
- Prefer Linux-native toolchains even when editing from Windows via VS Code Remote WSL.

## Source of truth

Before non-trivial work, read:

1. `AGENTS.md`
2. `docs/repo-structure.md`
3. relevant accepted ADRs
4. relevant threat-model docs
5. the applicable plan artifact
6. any more-local `AGENTS.md`

## Required process

- Create a plan artifact before non-trivial implementation.
- Use the closed ADR trigger list from `AGENTS.md` and `docs/adr/README.md`.
- Use tiered threat-model review from `docs/threat-model/README.md`.
- Keep changes scoped.
- Update docs in the same change when reality changes.

## Branch naming

Use one of:

- `feat/<topic>`
- `fix/<topic>`
- `docs/<topic>`
- `chore/<topic>`
- `infra/<topic>`

## Pull requests

Every PR must complete the repository template truthfully.
Security impact and privacy impact must be declared as `none` or described.

## Available validation

The repo is still in bootstrap, so validation is currently governance-oriented:

- `bash scripts/check-governance.sh`
- `bash scripts/check-no-secrets.sh`
- `bash scripts/validate.sh`

`make` may be used as a wrapper when available, but it is not the only supported entrypoint.

As code lands, language- and runtime-specific checks will be added here and in CI.

## Bootstrap exception

If a required governance artifact is missing, contributors may create the minimum safe governance files before product implementation.
That exception does not allow bypassing security, privacy, or traceability requirements.
