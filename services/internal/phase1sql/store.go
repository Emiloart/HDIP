package phase1sql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrRecordNotFound = errors.New("phase1 sql record not found")

type Store struct {
	db      *sql.DB
	dialect string
}

func Open(driverName string, dsn string) (*Store, error) {
	normalizedDriver := strings.TrimSpace(driverName)
	if normalizedDriver == "" {
		normalizedDriver = "pgx"
	}

	db, err := sql.Open(normalizedDriver, strings.TrimSpace(dsn))
	if err != nil {
		return nil, fmt.Errorf("open phase1 sql store: %w", err)
	}

	store := &Store{
		db:      db,
		dialect: dialectForDriver(normalizedDriver),
	}

	if err := store.db.Ping(); err != nil {
		_ = store.db.Close()
		return nil, fmt.Errorf("ping phase1 sql store: %w", err)
	}

	if err := store.ensureSchema(context.Background()); err != nil {
		_ = store.db.Close()
		return nil, err
	}

	return store, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}

	return s.db.Close()
}

func (s *Store) NextCredentialID(ctx context.Context, templateID string) (string, error) {
	sequence, err := s.nextSequence(ctx, "credential")
	if err != nil {
		return "", err
	}

	return formatCredentialID(templateID, sequence), nil
}

func (s *Store) NextVerificationID(ctx context.Context) (string, error) {
	sequence, err := s.nextSequence(ctx, "verification")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("verification_hdip_%03d", sequence), nil
}

func (s *Store) UpsertIssuerRecord(ctx context.Context, record IssuerRecord) error {
	allowedTemplateIDs, err := encodeJSON(record.AllowedTemplateIDs)
	if err != nil {
		return err
	}
	verificationKeyReferences, err := encodeJSON(record.VerificationKeyReferences)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(
		ctx,
		s.bind(`
INSERT INTO trust_registry_issuer_records (
  issuer_id,
  display_name,
  trust_state,
  allowed_template_ids,
  verification_key_references,
  created_at,
  updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (issuer_id) DO UPDATE SET
  display_name = excluded.display_name,
  trust_state = excluded.trust_state,
  allowed_template_ids = excluded.allowed_template_ids,
  verification_key_references = excluded.verification_key_references,
  created_at = excluded.created_at,
  updated_at = excluded.updated_at
`),
		strings.TrimSpace(record.IssuerID),
		record.DisplayName,
		record.TrustState,
		allowedTemplateIDs,
		verificationKeyReferences,
		formatTime(record.CreatedAt),
		formatTime(record.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("upsert issuer record: %w", err)
	}

	return nil
}

func (s *Store) DeleteIssuerRecord(ctx context.Context, issuerID string) error {
	result, err := s.db.ExecContext(
		ctx,
		s.bind(`DELETE FROM trust_registry_issuer_records WHERE issuer_id = ?`),
		strings.TrimSpace(issuerID),
	)
	if err != nil {
		return fmt.Errorf("delete issuer record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (s *Store) GetIssuerRecord(ctx context.Context, issuerID string) (IssuerRecord, error) {
	row := s.db.QueryRowContext(
		ctx,
		s.bind(`
SELECT
  issuer_id,
  display_name,
  trust_state,
  allowed_template_ids,
  verification_key_references,
  created_at,
  updated_at
FROM trust_registry_issuer_records
WHERE issuer_id = ?
`),
		strings.TrimSpace(issuerID),
	)

	var (
		record                       IssuerRecord
		allowedTemplateIDsRaw        string
		verificationKeyReferencesRaw string
		createdAtRaw                 string
		updatedAtRaw                 string
	)
	if err := row.Scan(
		&record.IssuerID,
		&record.DisplayName,
		&record.TrustState,
		&allowedTemplateIDsRaw,
		&verificationKeyReferencesRaw,
		&createdAtRaw,
		&updatedAtRaw,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return IssuerRecord{}, ErrRecordNotFound
		}
		return IssuerRecord{}, fmt.Errorf("load issuer record: %w", err)
	}

	if err := decodeJSON(allowedTemplateIDsRaw, &record.AllowedTemplateIDs); err != nil {
		return IssuerRecord{}, err
	}
	if err := decodeJSON(verificationKeyReferencesRaw, &record.VerificationKeyReferences); err != nil {
		return IssuerRecord{}, err
	}
	record.CreatedAt = parseTime(createdAtRaw)
	record.UpdatedAt = parseTime(updatedAtRaw)

	return record, nil
}

func (s *Store) CreateCredentialRecord(ctx context.Context, record CredentialRecord) error {
	claimsRaw, err := encodeJSON(record.Claims)
	if err != nil {
		return err
	}
	artifactRaw, err := encodeJSON(record.CredentialArtifact)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(
		ctx,
		s.bind(`
INSERT INTO issuer_api_credential_records (
  credential_id,
  issuer_id,
  template_id,
  subject_reference,
  claims_json,
  artifact_digest,
  credential_artifact_json,
  artifact_reference,
  status,
  status_reference,
  issued_at,
  expires_at,
  status_updated_at,
  superseded_by_credential_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`),
		strings.TrimSpace(record.CredentialID),
		strings.TrimSpace(record.IssuerID),
		strings.TrimSpace(record.TemplateID),
		strings.TrimSpace(record.SubjectReference),
		claimsRaw,
		strings.TrimSpace(record.ArtifactDigest),
		artifactRaw,
		strings.TrimSpace(record.ArtifactReference),
		strings.TrimSpace(record.Status),
		strings.TrimSpace(record.StatusReference),
		formatTime(record.IssuedAt),
		formatTime(record.ExpiresAt),
		formatTime(record.StatusUpdatedAt),
		strings.TrimSpace(record.SupersededByCredentialID),
	)
	if err != nil {
		return fmt.Errorf("create credential record: %w", err)
	}

	return nil
}

func (s *Store) UpsertCredentialRecord(ctx context.Context, record CredentialRecord) error {
	claimsRaw, err := encodeJSON(record.Claims)
	if err != nil {
		return err
	}
	artifactRaw, err := encodeJSON(record.CredentialArtifact)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(
		ctx,
		s.bind(`
INSERT INTO issuer_api_credential_records (
  credential_id,
  issuer_id,
  template_id,
  subject_reference,
  claims_json,
  artifact_digest,
  credential_artifact_json,
  artifact_reference,
  status,
  status_reference,
  issued_at,
  expires_at,
  status_updated_at,
  superseded_by_credential_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (credential_id) DO UPDATE SET
  issuer_id = excluded.issuer_id,
  template_id = excluded.template_id,
  subject_reference = excluded.subject_reference,
  claims_json = excluded.claims_json,
  artifact_digest = excluded.artifact_digest,
  credential_artifact_json = excluded.credential_artifact_json,
  artifact_reference = excluded.artifact_reference,
  status = excluded.status,
  status_reference = excluded.status_reference,
  issued_at = excluded.issued_at,
  expires_at = excluded.expires_at,
  status_updated_at = excluded.status_updated_at,
  superseded_by_credential_id = excluded.superseded_by_credential_id
`),
		strings.TrimSpace(record.CredentialID),
		strings.TrimSpace(record.IssuerID),
		strings.TrimSpace(record.TemplateID),
		strings.TrimSpace(record.SubjectReference),
		claimsRaw,
		strings.TrimSpace(record.ArtifactDigest),
		artifactRaw,
		strings.TrimSpace(record.ArtifactReference),
		strings.TrimSpace(record.Status),
		strings.TrimSpace(record.StatusReference),
		formatTime(record.IssuedAt),
		formatTime(record.ExpiresAt),
		formatTime(record.StatusUpdatedAt),
		strings.TrimSpace(record.SupersededByCredentialID),
	)
	if err != nil {
		return fmt.Errorf("upsert credential record: %w", err)
	}

	return nil
}

func (s *Store) GetCredentialRecord(ctx context.Context, credentialID string) (CredentialRecord, error) {
	row := s.db.QueryRowContext(
		ctx,
		s.bind(`
SELECT
  credential_id,
  issuer_id,
  template_id,
  subject_reference,
  claims_json,
  artifact_digest,
  credential_artifact_json,
  artifact_reference,
  status,
  status_reference,
  issued_at,
  expires_at,
  status_updated_at,
  superseded_by_credential_id
FROM issuer_api_credential_records
WHERE credential_id = ?
`),
		strings.TrimSpace(credentialID),
	)

	return scanCredentialRecord(row)
}

func (s *Store) GetCredentialRecordByArtifactDigest(ctx context.Context, artifactDigest string) (CredentialRecord, error) {
	row := s.db.QueryRowContext(
		ctx,
		s.bind(`
SELECT
  credential_id,
  issuer_id,
  template_id,
  subject_reference,
  claims_json,
  artifact_digest,
  credential_artifact_json,
  artifact_reference,
  status,
  status_reference,
  issued_at,
  expires_at,
  status_updated_at,
  superseded_by_credential_id
FROM issuer_api_credential_records
WHERE artifact_digest = ?
`),
		strings.TrimSpace(artifactDigest),
	)

	return scanCredentialRecord(row)
}

func (s *Store) UpdateCredentialStatus(
	ctx context.Context,
	credentialID string,
	status string,
	statusUpdatedAt time.Time,
	supersededByCredentialID string,
) error {
	result, err := s.db.ExecContext(
		ctx,
		s.bind(`
UPDATE issuer_api_credential_records
SET status = ?, status_updated_at = ?, superseded_by_credential_id = ?
WHERE credential_id = ?
`),
		strings.TrimSpace(status),
		formatTime(statusUpdatedAt),
		strings.TrimSpace(supersededByCredentialID),
		strings.TrimSpace(credentialID),
	)
	if err != nil {
		return fmt.Errorf("update credential status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (s *Store) CreateVerificationRequestRecord(ctx context.Context, record VerificationRequestRecord) error {
	actorRaw, err := encodeJSON(record.Actor)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(
		ctx,
		s.bind(`
INSERT INTO verifier_api_verification_request_records (
  verification_id,
  verifier_id,
  submitted_credential_digest,
  credential_id,
  policy_id,
  requested_at,
  actor_json,
  idempotency_key
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`),
		strings.TrimSpace(record.VerificationID),
		strings.TrimSpace(record.VerifierID),
		strings.TrimSpace(record.SubmittedCredentialDigest),
		strings.TrimSpace(record.CredentialID),
		strings.TrimSpace(record.PolicyID),
		formatTime(record.RequestedAt),
		actorRaw,
		strings.TrimSpace(record.IdempotencyKey),
	)
	if err != nil {
		return fmt.Errorf("create verification request record: %w", err)
	}

	return nil
}

func (s *Store) GetVerificationRequestRecord(ctx context.Context, verificationID string) (VerificationRequestRecord, error) {
	row := s.db.QueryRowContext(
		ctx,
		s.bind(`
SELECT
  verification_id,
  verifier_id,
  submitted_credential_digest,
  credential_id,
  policy_id,
  requested_at,
  actor_json,
  idempotency_key
FROM verifier_api_verification_request_records
WHERE verification_id = ?
`),
		strings.TrimSpace(verificationID),
	)

	var (
		record         VerificationRequestRecord
		requestedAtRaw string
		actorRaw       string
	)
	if err := row.Scan(
		&record.VerificationID,
		&record.VerifierID,
		&record.SubmittedCredentialDigest,
		&record.CredentialID,
		&record.PolicyID,
		&requestedAtRaw,
		&actorRaw,
		&record.IdempotencyKey,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return VerificationRequestRecord{}, ErrRecordNotFound
		}
		return VerificationRequestRecord{}, fmt.Errorf("load verification request record: %w", err)
	}

	record.RequestedAt = parseTime(requestedAtRaw)
	if err := decodeJSON(actorRaw, &record.Actor); err != nil {
		return VerificationRequestRecord{}, err
	}

	return record, nil
}

func (s *Store) CreateVerificationResultRecord(ctx context.Context, record VerificationResultRecord) error {
	reasonCodesRaw, err := encodeJSON(record.ReasonCodes)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(
		ctx,
		s.bind(`
INSERT INTO verifier_api_verification_result_records (
  verification_id,
  issuer_id,
  decision,
  reason_codes_json,
  issuer_trust_state,
  credential_status,
  evaluated_at,
  response_version
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`),
		strings.TrimSpace(record.VerificationID),
		strings.TrimSpace(record.IssuerID),
		strings.TrimSpace(record.Decision),
		reasonCodesRaw,
		strings.TrimSpace(record.IssuerTrustState),
		strings.TrimSpace(record.CredentialStatus),
		formatTime(record.EvaluatedAt),
		strings.TrimSpace(record.ResponseVersion),
	)
	if err != nil {
		return fmt.Errorf("create verification result record: %w", err)
	}

	return nil
}

func (s *Store) GetVerificationResultRecord(ctx context.Context, verificationID string) (VerificationResultRecord, error) {
	row := s.db.QueryRowContext(
		ctx,
		s.bind(`
SELECT
  verification_id,
  issuer_id,
  decision,
  reason_codes_json,
  issuer_trust_state,
  credential_status,
  evaluated_at,
  response_version
FROM verifier_api_verification_result_records
WHERE verification_id = ?
`),
		strings.TrimSpace(verificationID),
	)

	var (
		record         VerificationResultRecord
		reasonCodesRaw string
		evaluatedAtRaw string
	)
	if err := row.Scan(
		&record.VerificationID,
		&record.IssuerID,
		&record.Decision,
		&reasonCodesRaw,
		&record.IssuerTrustState,
		&record.CredentialStatus,
		&evaluatedAtRaw,
		&record.ResponseVersion,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return VerificationResultRecord{}, ErrRecordNotFound
		}
		return VerificationResultRecord{}, fmt.Errorf("load verification result record: %w", err)
	}

	if err := decodeJSON(reasonCodesRaw, &record.ReasonCodes); err != nil {
		return VerificationResultRecord{}, err
	}
	record.EvaluatedAt = parseTime(evaluatedAtRaw)

	return record, nil
}

func (s *Store) AppendAuditRecord(ctx context.Context, record AuditRecord) error {
	actorRaw, err := encodeJSON(record.Actor)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(
		ctx,
		s.bind(`
INSERT INTO phase1_audit_records (
  audit_id,
  actor_json,
  action_name,
  resource_type,
  resource_id,
  request_id,
  idempotency_key,
  outcome,
  occurred_at,
  service_name
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`),
		strings.TrimSpace(record.AuditID),
		actorRaw,
		strings.TrimSpace(record.Action),
		strings.TrimSpace(record.ResourceType),
		strings.TrimSpace(record.ResourceID),
		strings.TrimSpace(record.RequestID),
		strings.TrimSpace(record.IdempotencyKey),
		strings.TrimSpace(record.Outcome),
		formatTime(record.OccurredAt),
		strings.TrimSpace(record.ServiceName),
	)
	if err != nil {
		return fmt.Errorf("append audit record: %w", err)
	}

	return nil
}

func (s *Store) ListAuditRecords(ctx context.Context) ([]AuditRecord, error) {
	rows, err := s.db.QueryContext(
		ctx,
		s.bind(`
SELECT
  audit_id,
  actor_json,
  action_name,
  resource_type,
  resource_id,
  request_id,
  idempotency_key,
  outcome,
  occurred_at,
  service_name
FROM phase1_audit_records
ORDER BY occurred_at, audit_id
`),
	)
	if err != nil {
		return nil, fmt.Errorf("list audit records: %w", err)
	}
	defer rows.Close()

	records := make([]AuditRecord, 0)
	for rows.Next() {
		var (
			record        AuditRecord
			actorRaw      string
			occurredAtRaw string
		)
		if err := rows.Scan(
			&record.AuditID,
			&actorRaw,
			&record.Action,
			&record.ResourceType,
			&record.ResourceID,
			&record.RequestID,
			&record.IdempotencyKey,
			&record.Outcome,
			&occurredAtRaw,
			&record.ServiceName,
		); err != nil {
			return nil, fmt.Errorf("scan audit record: %w", err)
		}

		record.OccurredAt = parseTime(occurredAtRaw)
		if err := decodeJSON(actorRaw, &record.Actor); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (s *Store) CreateIdempotencyRecord(ctx context.Context, record IdempotencyRecord) error {
	return s.upsertIdempotencyRecord(ctx, record, false)
}

func (s *Store) GetIdempotencyRecord(
	ctx context.Context,
	operation string,
	callerOrganizationID string,
	callerPrincipalID string,
	callerActorType string,
	idempotencyKey string,
) (IdempotencyRecord, error) {
	row := s.db.QueryRowContext(
		ctx,
		s.bind(`
SELECT
  operation,
  caller_principal_id,
  caller_organization_id,
  caller_actor_type,
  idempotency_key,
  request_fingerprint,
  reservation_state,
  response_status_code,
  resource_type,
  resource_id,
  location,
  response_body,
  created_at,
  updated_at
FROM phase1_idempotency_records
WHERE operation = ?
  AND caller_organization_id = ?
  AND caller_principal_id = ?
  AND caller_actor_type = ?
  AND idempotency_key = ?
`),
		strings.TrimSpace(operation),
		strings.TrimSpace(callerOrganizationID),
		strings.TrimSpace(callerPrincipalID),
		strings.TrimSpace(callerActorType),
		strings.TrimSpace(idempotencyKey),
	)

	return scanIdempotencyRecord(row)
}

func (s *Store) ListIdempotencyRecords(ctx context.Context) ([]IdempotencyRecord, error) {
	rows, err := s.db.QueryContext(
		ctx,
		s.bind(`
SELECT
  operation,
  caller_principal_id,
  caller_organization_id,
  caller_actor_type,
  idempotency_key,
  request_fingerprint,
  reservation_state,
  response_status_code,
  resource_type,
  resource_id,
  location,
  response_body,
  created_at,
  updated_at
FROM phase1_idempotency_records
ORDER BY created_at, operation, idempotency_key
`),
	)
	if err != nil {
		return nil, fmt.Errorf("list idempotency records: %w", err)
	}
	defer rows.Close()

	records := make([]IdempotencyRecord, 0)
	for rows.Next() {
		record, err := scanIdempotencyRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (s *Store) ReserveIdempotencyRecord(ctx context.Context, record IdempotencyRecord) (IdempotencyReservationResult, error) {
	reservationRecord := record
	now := record.CreatedAt.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	reservationRecord.CreatedAt = now
	reservationRecord.UpdatedAt = now
	reservationRecord.State = IdempotencyStateReserved

	result, err := s.db.ExecContext(
		ctx,
		s.bind(`
INSERT INTO phase1_idempotency_records (
  operation,
  caller_principal_id,
  caller_organization_id,
  caller_actor_type,
  idempotency_key,
  request_fingerprint,
  reservation_state,
  response_status_code,
  resource_type,
  resource_id,
  location,
  response_body,
  created_at,
  updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (operation, caller_organization_id, caller_principal_id, caller_actor_type, idempotency_key) DO NOTHING
`),
		strings.TrimSpace(reservationRecord.Operation),
		strings.TrimSpace(reservationRecord.CallerPrincipalID),
		strings.TrimSpace(reservationRecord.CallerOrganizationID),
		strings.TrimSpace(reservationRecord.CallerActorType),
		strings.TrimSpace(reservationRecord.IdempotencyKey),
		strings.TrimSpace(reservationRecord.RequestFingerprint),
		IdempotencyStateReserved,
		0,
		strings.TrimSpace(reservationRecord.ResourceType),
		strings.TrimSpace(reservationRecord.ResourceID),
		strings.TrimSpace(reservationRecord.Location),
		string(reservationRecord.ResponseBody),
		formatTime(reservationRecord.CreatedAt),
		formatTime(reservationRecord.UpdatedAt),
	)
	if err != nil {
		return IdempotencyReservationResult{}, fmt.Errorf("reserve idempotency record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected > 0 {
		return IdempotencyReservationResult{
			Outcome: IdempotencyReservationReserved,
			Record:  reservationRecord,
		}, nil
	}

	storedRecord, err := s.GetIdempotencyRecord(
		ctx,
		reservationRecord.Operation,
		reservationRecord.CallerOrganizationID,
		reservationRecord.CallerPrincipalID,
		reservationRecord.CallerActorType,
		reservationRecord.IdempotencyKey,
	)
	if err != nil {
		return IdempotencyReservationResult{}, err
	}

	if strings.TrimSpace(storedRecord.RequestFingerprint) != strings.TrimSpace(reservationRecord.RequestFingerprint) {
		return IdempotencyReservationResult{
			Outcome: IdempotencyReservationConflict,
			Record:  storedRecord,
		}, nil
	}

	if strings.TrimSpace(storedRecord.State) == IdempotencyStateCompleted {
		return IdempotencyReservationResult{
			Outcome: IdempotencyReservationReplay,
			Record:  storedRecord,
		}, nil
	}

	return IdempotencyReservationResult{
		Outcome: IdempotencyReservationInProgress,
		Record:  storedRecord,
	}, nil
}

func (s *Store) CompleteIdempotencyRecord(ctx context.Context, record IdempotencyRecord) error {
	record.State = IdempotencyStateCompleted
	record.UpdatedAt = record.UpdatedAt.UTC()
	if record.UpdatedAt.IsZero() {
		record.UpdatedAt = time.Now().UTC()
	}

	result, err := s.db.ExecContext(
		ctx,
		s.bind(`
UPDATE phase1_idempotency_records
SET reservation_state = ?, response_status_code = ?, resource_type = ?, resource_id = ?, location = ?, response_body = ?, updated_at = ?
WHERE operation = ?
  AND caller_organization_id = ?
  AND caller_principal_id = ?
  AND caller_actor_type = ?
  AND idempotency_key = ?
`),
		IdempotencyStateCompleted,
		record.ResponseStatusCode,
		strings.TrimSpace(record.ResourceType),
		strings.TrimSpace(record.ResourceID),
		strings.TrimSpace(record.Location),
		string(record.ResponseBody),
		formatTime(record.UpdatedAt),
		strings.TrimSpace(record.Operation),
		strings.TrimSpace(record.CallerOrganizationID),
		strings.TrimSpace(record.CallerPrincipalID),
		strings.TrimSpace(record.CallerActorType),
		strings.TrimSpace(record.IdempotencyKey),
	)
	if err != nil {
		return fmt.Errorf("complete idempotency record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (s *Store) ReleaseIdempotencyRecord(
	ctx context.Context,
	operation string,
	callerOrganizationID string,
	callerPrincipalID string,
	callerActorType string,
	idempotencyKey string,
) error {
	_, err := s.db.ExecContext(
		ctx,
		s.bind(`
DELETE FROM phase1_idempotency_records
WHERE operation = ?
  AND caller_organization_id = ?
  AND caller_principal_id = ?
  AND caller_actor_type = ?
  AND idempotency_key = ?
  AND reservation_state = ?
`),
		strings.TrimSpace(operation),
		strings.TrimSpace(callerOrganizationID),
		strings.TrimSpace(callerPrincipalID),
		strings.TrimSpace(callerActorType),
		strings.TrimSpace(idempotencyKey),
		IdempotencyStateReserved,
	)
	if err != nil {
		return fmt.Errorf("release idempotency record: %w", err)
	}

	return nil
}

func (s *Store) ensureSchema(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS phase1_sequences (
  name TEXT PRIMARY KEY,
  value INTEGER NOT NULL
)`,
		`CREATE TABLE IF NOT EXISTS trust_registry_issuer_records (
  issuer_id TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  trust_state TEXT NOT NULL,
  allowed_template_ids TEXT NOT NULL,
  verification_key_references TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
)`,
		`CREATE TABLE IF NOT EXISTS issuer_api_credential_records (
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
)`,
		`CREATE INDEX IF NOT EXISTS issuer_api_credential_records_artifact_digest_idx
ON issuer_api_credential_records (artifact_digest)`,
		`CREATE TABLE IF NOT EXISTS verifier_api_verification_request_records (
  verification_id TEXT PRIMARY KEY,
  verifier_id TEXT NOT NULL,
  submitted_credential_digest TEXT NOT NULL,
  credential_id TEXT NOT NULL,
  policy_id TEXT NOT NULL,
  requested_at TEXT NOT NULL,
  actor_json TEXT NOT NULL,
  idempotency_key TEXT NOT NULL
)`,
		`CREATE TABLE IF NOT EXISTS verifier_api_verification_result_records (
  verification_id TEXT PRIMARY KEY,
  issuer_id TEXT NOT NULL,
  decision TEXT NOT NULL,
  reason_codes_json TEXT NOT NULL,
  issuer_trust_state TEXT NOT NULL,
  credential_status TEXT NOT NULL,
  evaluated_at TEXT NOT NULL,
  response_version TEXT NOT NULL
)`,
		`CREATE TABLE IF NOT EXISTS phase1_audit_records (
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
)`,
		`CREATE TABLE IF NOT EXISTS phase1_idempotency_records (
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
)`,
	}

	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("initialize phase1 sql schema: %w", err)
		}
	}

	return nil
}

func (s *Store) nextSequence(ctx context.Context, name string) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin sequence transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(
		ctx,
		s.bind(`INSERT INTO phase1_sequences (name, value) VALUES (?, 0) ON CONFLICT (name) DO NOTHING`),
		strings.TrimSpace(name),
	); err != nil {
		return 0, fmt.Errorf("ensure sequence row: %w", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		s.bind(`UPDATE phase1_sequences SET value = value + 1 WHERE name = ?`),
		strings.TrimSpace(name),
	); err != nil {
		return 0, fmt.Errorf("increment sequence: %w", err)
	}

	var value int
	if err := tx.QueryRowContext(
		ctx,
		s.bind(`SELECT value FROM phase1_sequences WHERE name = ?`),
		strings.TrimSpace(name),
	).Scan(&value); err != nil {
		return 0, fmt.Errorf("load sequence value: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit sequence transaction: %w", err)
	}

	return value, nil
}

func (s *Store) upsertIdempotencyRecord(ctx context.Context, record IdempotencyRecord, allowUpdate bool) error {
	if record.State == "" {
		record.State = IdempotencyStateCompleted
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	}
	if record.UpdatedAt.IsZero() {
		record.UpdatedAt = record.CreatedAt
	}

	query := `
INSERT INTO phase1_idempotency_records (
  operation,
  caller_principal_id,
  caller_organization_id,
  caller_actor_type,
  idempotency_key,
  request_fingerprint,
  reservation_state,
  response_status_code,
  resource_type,
  resource_id,
  location,
  response_body,
  created_at,
  updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`
	if allowUpdate {
		query += `
ON CONFLICT (operation, caller_organization_id, caller_principal_id, caller_actor_type, idempotency_key) DO UPDATE SET
  request_fingerprint = excluded.request_fingerprint,
  reservation_state = excluded.reservation_state,
  response_status_code = excluded.response_status_code,
  resource_type = excluded.resource_type,
  resource_id = excluded.resource_id,
  location = excluded.location,
  response_body = excluded.response_body,
  updated_at = excluded.updated_at
`
	}

	_, err := s.db.ExecContext(
		ctx,
		s.bind(query),
		strings.TrimSpace(record.Operation),
		strings.TrimSpace(record.CallerPrincipalID),
		strings.TrimSpace(record.CallerOrganizationID),
		strings.TrimSpace(record.CallerActorType),
		strings.TrimSpace(record.IdempotencyKey),
		strings.TrimSpace(record.RequestFingerprint),
		strings.TrimSpace(record.State),
		record.ResponseStatusCode,
		strings.TrimSpace(record.ResourceType),
		strings.TrimSpace(record.ResourceID),
		strings.TrimSpace(record.Location),
		string(record.ResponseBody),
		formatTime(record.CreatedAt),
		formatTime(record.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("upsert idempotency record: %w", err)
	}

	return nil
}

func (s *Store) bind(query string) string {
	if s.dialect != "postgres" {
		return query
	}

	var (
		builder strings.Builder
		index   int
	)
	for _, r := range query {
		if r != '?' {
			builder.WriteRune(r)
			continue
		}

		index++
		builder.WriteString(fmt.Sprintf("$%d", index))
	}

	return builder.String()
}

func scanCredentialRecord(scanner interface{ Scan(dest ...any) error }) (CredentialRecord, error) {
	var (
		record             CredentialRecord
		claimsRaw          string
		artifactRaw        string
		issuedAtRaw        string
		expiresAtRaw       string
		statusUpdatedAtRaw string
	)
	if err := scanner.Scan(
		&record.CredentialID,
		&record.IssuerID,
		&record.TemplateID,
		&record.SubjectReference,
		&claimsRaw,
		&record.ArtifactDigest,
		&artifactRaw,
		&record.ArtifactReference,
		&record.Status,
		&record.StatusReference,
		&issuedAtRaw,
		&expiresAtRaw,
		&statusUpdatedAtRaw,
		&record.SupersededByCredentialID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CredentialRecord{}, ErrRecordNotFound
		}
		return CredentialRecord{}, fmt.Errorf("scan credential record: %w", err)
	}

	if err := decodeJSON(claimsRaw, &record.Claims); err != nil {
		return CredentialRecord{}, err
	}
	if err := decodeJSON(artifactRaw, &record.CredentialArtifact); err != nil {
		return CredentialRecord{}, err
	}
	record.IssuedAt = parseTime(issuedAtRaw)
	record.ExpiresAt = parseTime(expiresAtRaw)
	record.StatusUpdatedAt = parseTime(statusUpdatedAtRaw)

	return record, nil
}

func scanIdempotencyRecord(scanner interface{ Scan(dest ...any) error }) (IdempotencyRecord, error) {
	var (
		record       IdempotencyRecord
		responseBody string
		createdAtRaw string
		updatedAtRaw string
	)
	if err := scanner.Scan(
		&record.Operation,
		&record.CallerPrincipalID,
		&record.CallerOrganizationID,
		&record.CallerActorType,
		&record.IdempotencyKey,
		&record.RequestFingerprint,
		&record.State,
		&record.ResponseStatusCode,
		&record.ResourceType,
		&record.ResourceID,
		&record.Location,
		&responseBody,
		&createdAtRaw,
		&updatedAtRaw,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return IdempotencyRecord{}, ErrRecordNotFound
		}
		return IdempotencyRecord{}, fmt.Errorf("scan idempotency record: %w", err)
	}

	record.ResponseBody = json.RawMessage(responseBody)
	record.CreatedAt = parseTime(createdAtRaw)
	record.UpdatedAt = parseTime(updatedAtRaw)

	return record, nil
}

func encodeJSON(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("encode phase1 sql json value: %w", err)
	}

	return string(raw), nil
}

func decodeJSON(raw string, destination any) error {
	if strings.TrimSpace(raw) == "" {
		raw = `null`
	}

	if err := json.Unmarshal([]byte(raw), destination); err != nil {
		return fmt.Errorf("decode phase1 sql json value: %w", err)
	}

	return nil
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return time.Time{}.UTC().Format(time.RFC3339Nano)
	}

	return value.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) time.Time {
	if strings.TrimSpace(value) == "" {
		return time.Time{}
	}

	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}
	}

	return parsed.UTC()
}

func dialectForDriver(driverName string) string {
	switch strings.TrimSpace(driverName) {
	case "pgx", "postgres", "postgresql":
		return "postgres"
	default:
		return "sqlite"
	}
}

func formatCredentialID(templateID string, sequence int) string {
	sanitizedTemplateID := strings.NewReplacer("-", "_", ".", "_").Replace(strings.TrimSpace(templateID))
	if sanitizedTemplateID == "" {
		sanitizedTemplateID = "hdip_passport_basic"
	}

	return fmt.Sprintf("cred_%s_%03d", sanitizedTemplateID, sequence)
}
