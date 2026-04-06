SHELL := /usr/bin/env bash

.PHONY: governance-check security-check docs-check validate

governance-check:
	bash scripts/check-governance.sh

security-check:
	bash scripts/check-no-secrets.sh

docs-check: governance-check

validate: governance-check security-check
