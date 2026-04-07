SHELL := /usr/bin/env bash

.PHONY: governance-check security-check docs-check validate validate-rust validate-go validate-web

governance-check:
	bash scripts/check-governance.sh

security-check:
	bash scripts/check-no-secrets.sh

docs-check: governance-check

validate-rust:
	bash scripts/validate-rust.sh

validate-go:
	bash scripts/validate-go.sh

validate-web:
	bash scripts/validate-web.sh

validate: governance-check security-check
	bash scripts/validate-rust.sh
	bash scripts/validate-go.sh
	bash scripts/validate-web.sh
