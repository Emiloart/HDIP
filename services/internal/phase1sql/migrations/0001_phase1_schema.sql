CREATE TABLE IF NOT EXISTS phase1_sequences (
  name TEXT PRIMARY KEY,
  value INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS trust_registry_issuer_records (
  issuer_id TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  trust_state TEXT NOT NULL,
  allowed_template_ids TEXT NOT NULL,
  verification_key_references TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS issuer_api_credential_records (
  credential_id TEXT PRIMARY KEY,
  issuer_id TEXT NOT NULL,
  template_id TEXT NOT NULL,
  subject_reference TEXT NOT NULL,
  claims_json TEXT NOT NULL,
  artifact_digest TEXT NOT NULL,
  credential_artifact_json TEXT NOT NULL,
  artifact_reference TEXT NOT NULL,
  status TEXT NOT NULL,
  status_reference TEXT NOT NULL,
  issued_at TEXT NOT NULL,
  expires_at TEXT NOT NULL,
  status_updated_at TEXT NOT NULL,
  superseded_by_credential_id TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS issuer_api_credential_records_artifact_digest_idx
ON issuer_api_credential_records (artifact_digest);

CREATE TABLE IF NOT EXISTS verifier_api_verification_request_records (
  verification_id TEXT PRIMARY KEY,
  verifier_id TEXT NOT NULL,
  submitted_credential_digest TEXT NOT NULL,
  credential_id TEXT NOT NULL,
  policy_id TEXT NOT NULL,
  requested_at TEXT NOT NULL,
  actor_json TEXT NOT NULL,
  idempotency_key TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS verifier_api_verification_result_records (
  verification_id TEXT PRIMARY KEY,
  issuer_id TEXT NOT NULL,
  decision TEXT NOT NULL,
  reason_codes_json TEXT NOT NULL,
  issuer_trust_state TEXT NOT NULL,
  credential_status TEXT NOT NULL,
  evaluated_at TEXT NOT NULL,
  response_version TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS phase1_audit_records (
  audit_id TEXT PRIMARY KEY,
  actor_json TEXT NOT NULL,
  action_name TEXT NOT NULL,
  resource_type TEXT NOT NULL,
  resource_id TEXT NOT NULL,
  request_id TEXT NOT NULL,
  idempotency_key TEXT NOT NULL,
  outcome TEXT NOT NULL,
  occurred_at TEXT NOT NULL,
  service_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS phase1_idempotency_records (
  operation TEXT NOT NULL,
  caller_principal_id TEXT NOT NULL,
  caller_organization_id TEXT NOT NULL,
  caller_actor_type TEXT NOT NULL,
  idempotency_key TEXT NOT NULL,
  request_fingerprint TEXT NOT NULL,
  reservation_state TEXT NOT NULL,
  response_status_code INTEGER NOT NULL,
  resource_type TEXT NOT NULL,
  resource_id TEXT NOT NULL,
  location TEXT NOT NULL,
  response_body TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  PRIMARY KEY (operation, caller_organization_id, caller_principal_id, caller_actor_type, idempotency_key)
);
