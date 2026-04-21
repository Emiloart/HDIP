package phase1

import (
	"context"
	"errors"
	"time"

	"github.com/Emiloart/HDIP/packages/go/foundation/authctx"
	phase1runtime "github.com/Emiloart/HDIP/services/internal/phase1runtime"
	phase1sql "github.com/Emiloart/HDIP/services/internal/phase1sql"
)

type StoreOptions struct {
	DatabaseDriver  string
	DatabaseURL     string
	LegacyStatePath string
}

type RuntimeStore struct {
	legacy *phase1runtime.Store
	sql    *phase1sql.Store
}

func OpenStore(options StoreOptions) (*RuntimeStore, error) {
	if options.DatabaseURL != "" {
		store, err := phase1sql.Open(options.DatabaseDriver, options.DatabaseURL)
		if err != nil {
			return nil, err
		}

		return &RuntimeStore{sql: store}, nil
	}

	return OpenRuntimeStore(options.LegacyStatePath)
}

func OpenRuntimeStore(path string) (*RuntimeStore, error) {
	store, err := phase1runtime.Open(path)
	if err != nil {
		return nil, err
	}

	return &RuntimeStore{legacy: store}, nil
}

func NewRuntimeStore(runtimeStore *phase1runtime.Store) *RuntimeStore {
	return &RuntimeStore{legacy: runtimeStore}
}

func NewSQLRuntimeStore(runtimeStore *phase1sql.Store) *RuntimeStore {
	return &RuntimeStore{sql: runtimeStore}
}

func (s *RuntimeStore) Close() error {
	if s == nil {
		return nil
	}

	switch {
	case s.sql != nil:
		return s.sql.Close()
	case s.legacy != nil:
		return s.legacy.Close()
	default:
		return nil
	}
}

func (s *RuntimeStore) GetIssuerRecord(ctx context.Context, issuerID string) (IssuerRecord, error) {
	if s.sql != nil {
		record, err := s.sql.GetIssuerRecord(ctx, issuerID)
		if err != nil {
			return IssuerRecord{}, translateSQLError(err)
		}

		return issuerRecordFromSQL(record), nil
	}

	record, err := s.legacy.GetIssuerRecord(ctx, issuerID)
	if err != nil {
		return IssuerRecord{}, translateRuntimeError(err)
	}

	return issuerRecordFromRuntime(record), nil
}

func (s *RuntimeStore) GetIssuerTrustRecord(ctx context.Context, issuerID string) (IssuerTrustRecord, error) {
	record, err := s.GetIssuerRecord(ctx, issuerID)
	if err != nil {
		return IssuerTrustRecord{}, err
	}

	return IssuerTrustRecord{
		IssuerID:                  record.IssuerID,
		TrustState:                record.TrustState,
		AllowedTemplateIDs:        append([]string(nil), record.AllowedTemplateIDs...),
		VerificationKeyReferences: append([]string(nil), record.VerificationKeyReferences...),
	}, nil
}

func (s *RuntimeStore) GetCredentialRecord(ctx context.Context, credentialID string) (CredentialRecord, error) {
	if s.sql != nil {
		record, err := s.sql.GetCredentialRecord(ctx, credentialID)
		if err != nil {
			return CredentialRecord{}, translateSQLError(err)
		}

		return credentialRecordFromSQL(record), nil
	}

	record, err := s.legacy.GetCredentialRecord(ctx, credentialID)
	if err != nil {
		return CredentialRecord{}, translateRuntimeError(err)
	}

	return credentialRecordFromRuntime(record), nil
}

func (s *RuntimeStore) GetCredentialRecordByArtifactDigest(ctx context.Context, artifactDigest string) (CredentialRecord, error) {
	if s.sql != nil {
		record, err := s.sql.GetCredentialRecordByArtifactDigest(ctx, artifactDigest)
		if err != nil {
			return CredentialRecord{}, translateSQLError(err)
		}

		return credentialRecordFromSQL(record), nil
	}

	record, err := s.legacy.GetCredentialRecordByArtifactDigest(ctx, artifactDigest)
	if err != nil {
		return CredentialRecord{}, translateRuntimeError(err)
	}

	return credentialRecordFromRuntime(record), nil
}

func (s *RuntimeStore) NextVerificationID(ctx context.Context) (string, error) {
	if s.sql != nil {
		return s.sql.NextVerificationID(ctx)
	}

	return s.legacy.NextVerificationID(ctx)
}

func (s *RuntimeStore) CreateVerificationRequestRecord(ctx context.Context, record VerificationRequestRecord) error {
	if s.sql != nil {
		return translateSQLError(s.sql.CreateVerificationRequestRecord(ctx, verificationRequestRecordToSQL(record)))
	}

	return translateRuntimeError(s.legacy.CreateVerificationRequestRecord(ctx, verificationRequestRecordToRuntime(record)))
}

func (s *RuntimeStore) GetVerificationRequestRecord(ctx context.Context, verificationID string) (VerificationRequestRecord, error) {
	if s.sql != nil {
		record, err := s.sql.GetVerificationRequestRecord(ctx, verificationID)
		if err != nil {
			return VerificationRequestRecord{}, translateSQLError(err)
		}

		return verificationRequestRecordFromSQL(record), nil
	}

	record, err := s.legacy.GetVerificationRequestRecord(ctx, verificationID)
	if err != nil {
		return VerificationRequestRecord{}, translateRuntimeError(err)
	}

	return verificationRequestRecordFromRuntime(record), nil
}

func (s *RuntimeStore) CreateVerificationResultRecord(ctx context.Context, record VerificationResultRecord) error {
	if s.sql != nil {
		return translateSQLError(s.sql.CreateVerificationResultRecord(ctx, verificationResultRecordToSQL(record)))
	}

	return translateRuntimeError(s.legacy.CreateVerificationResultRecord(ctx, verificationResultRecordToRuntime(record)))
}

func (s *RuntimeStore) GetVerificationResultRecord(ctx context.Context, verificationID string) (VerificationResultRecord, error) {
	if s.sql != nil {
		record, err := s.sql.GetVerificationResultRecord(ctx, verificationID)
		if err != nil {
			return VerificationResultRecord{}, translateSQLError(err)
		}

		return verificationResultRecordFromSQL(record), nil
	}

	record, err := s.legacy.GetVerificationResultRecord(ctx, verificationID)
	if err != nil {
		return VerificationResultRecord{}, translateRuntimeError(err)
	}

	return verificationResultRecordFromRuntime(record), nil
}

func (s *RuntimeStore) AppendAuditRecord(ctx context.Context, record AuditRecord) error {
	if s.sql != nil {
		return translateSQLError(s.sql.AppendAuditRecord(ctx, auditRecordToSQL(record)))
	}

	return translateRuntimeError(s.legacy.AppendAuditRecord(ctx, auditRecordToRuntime(record)))
}

func (s *RuntimeStore) CreateIdempotencyRecord(ctx context.Context, record IdempotencyRecord) error {
	if s.sql != nil {
		return translateSQLError(s.sql.CreateIdempotencyRecord(ctx, idempotencyRecordToSQL(record)))
	}

	return translateRuntimeError(s.legacy.CreateIdempotencyRecord(ctx, idempotencyRecordToRuntime(record)))
}

func (s *RuntimeStore) ReserveIdempotencyRecord(ctx context.Context, record IdempotencyRecord) (IdempotencyReservationResult, error) {
	if s.sql != nil {
		result, err := s.sql.ReserveIdempotencyRecord(ctx, idempotencyRecordToSQL(record))
		if err != nil {
			return IdempotencyReservationResult{}, translateSQLError(err)
		}

		return idempotencyReservationResultFromSQL(result), nil
	}

	result, err := s.legacy.ReserveIdempotencyRecord(ctx, idempotencyRecordToRuntime(record))
	if err != nil {
		return IdempotencyReservationResult{}, translateRuntimeError(err)
	}

	return idempotencyReservationResultFromRuntime(result), nil
}

func (s *RuntimeStore) CompleteIdempotencyRecord(ctx context.Context, record IdempotencyRecord) error {
	if s.sql != nil {
		return translateSQLError(s.sql.CompleteIdempotencyRecord(ctx, idempotencyRecordToSQL(record)))
	}

	return translateRuntimeError(s.legacy.CompleteIdempotencyRecord(ctx, idempotencyRecordToRuntime(record)))
}

func (s *RuntimeStore) ReleaseIdempotencyRecord(
	ctx context.Context,
	operation string,
	callerOrganizationID string,
	callerPrincipalID string,
	callerActorType string,
	idempotencyKey string,
) error {
	if s.sql != nil {
		return translateSQLError(
			s.sql.ReleaseIdempotencyRecord(ctx, operation, callerOrganizationID, callerPrincipalID, callerActorType, idempotencyKey),
		)
	}

	return translateRuntimeError(
		s.legacy.ReleaseIdempotencyRecord(ctx, operation, callerOrganizationID, callerPrincipalID, callerActorType, idempotencyKey),
	)
}

func (s *RuntimeStore) GetIdempotencyRecord(
	ctx context.Context,
	operation string,
	callerOrganizationID string,
	callerPrincipalID string,
	callerActorType string,
	idempotencyKey string,
) (IdempotencyRecord, error) {
	if s.sql != nil {
		record, err := s.sql.GetIdempotencyRecord(ctx, operation, callerOrganizationID, callerPrincipalID, callerActorType, idempotencyKey)
		if err != nil {
			return IdempotencyRecord{}, translateSQLError(err)
		}

		return idempotencyRecordFromSQL(record), nil
	}

	record, err := s.legacy.GetIdempotencyRecord(ctx, operation, callerOrganizationID, callerPrincipalID, callerActorType, idempotencyKey)
	if err != nil {
		return IdempotencyRecord{}, translateRuntimeError(err)
	}

	return idempotencyRecordFromRuntime(record), nil
}

func (s *RuntimeStore) SeedIssuerRecord(record IssuerRecord) error {
	if s.sql != nil {
		return translateSQLError(s.sql.UpsertIssuerRecord(context.Background(), issuerRecordToSQL(record)))
	}

	return translateRuntimeError(s.legacy.UpsertIssuerRecord(context.Background(), issuerRecordToRuntime(record)))
}

func (s *RuntimeStore) DeleteIssuerRecord(issuerID string) error {
	if s.sql != nil {
		return translateSQLError(s.sql.DeleteIssuerRecord(context.Background(), issuerID))
	}

	return translateRuntimeError(s.legacy.DeleteIssuerRecord(context.Background(), issuerID))
}

func (s *RuntimeStore) SeedCredentialRecord(record CredentialRecord) error {
	if s.sql != nil {
		return translateSQLError(s.sql.UpsertCredentialRecord(context.Background(), credentialRecordToSQL(record)))
	}

	return translateRuntimeError(s.legacy.UpsertCredentialRecord(context.Background(), credentialRecordToRuntime(record)))
}

func (s *RuntimeStore) UpdateCredentialStatus(
	ctx context.Context,
	credentialID string,
	status CredentialStatusSnapshot,
	statusUpdatedAt time.Time,
	supersededByCredentialID string,
) error {
	if s.sql != nil {
		return translateSQLError(
			s.sql.UpdateCredentialStatus(ctx, credentialID, string(status), statusUpdatedAt, supersededByCredentialID),
		)
	}

	return translateRuntimeError(
		s.legacy.UpdateCredentialStatus(ctx, credentialID, string(status), statusUpdatedAt, supersededByCredentialID),
	)
}

func (s *RuntimeStore) AuditRecords() ([]AuditRecord, error) {
	if s.sql != nil {
		records, err := s.sql.ListAuditRecords(context.Background())
		if err != nil {
			return nil, translateSQLError(err)
		}

		result := make([]AuditRecord, 0, len(records))
		for _, record := range records {
			result = append(result, auditRecordFromSQL(record))
		}

		return result, nil
	}

	records, err := s.legacy.ListAuditRecords(context.Background())
	if err != nil {
		return nil, translateRuntimeError(err)
	}

	result := make([]AuditRecord, 0, len(records))
	for _, record := range records {
		result = append(result, auditRecordFromRuntime(record))
	}

	return result, nil
}

func (s *RuntimeStore) IdempotencyRecords() ([]IdempotencyRecord, error) {
	if s.sql != nil {
		records, err := s.sql.ListIdempotencyRecords(context.Background())
		if err != nil {
			return nil, translateSQLError(err)
		}

		result := make([]IdempotencyRecord, 0, len(records))
		for _, record := range records {
			result = append(result, idempotencyRecordFromSQL(record))
		}

		return result, nil
	}

	records, err := s.legacy.ListIdempotencyRecords(context.Background())
	if err != nil {
		return nil, translateRuntimeError(err)
	}

	result := make([]IdempotencyRecord, 0, len(records))
	for _, record := range records {
		result = append(result, idempotencyRecordFromRuntime(record))
	}

	return result, nil
}

func translateRuntimeError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, phase1runtime.ErrRecordNotFound) {
		return ErrRecordNotFound
	}

	return err
}

func translateSQLError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, phase1sql.ErrRecordNotFound) {
		return ErrRecordNotFound
	}

	return err
}

func issuerRecordToRuntime(record IssuerRecord) phase1runtime.IssuerRecord {
	return phase1runtime.IssuerRecord{
		IssuerID:                  record.IssuerID,
		DisplayName:               record.DisplayName,
		TrustState:                record.TrustState,
		AllowedTemplateIDs:        append([]string(nil), record.AllowedTemplateIDs...),
		VerificationKeyReferences: append([]string(nil), record.VerificationKeyReferences...),
		CreatedAt:                 record.CreatedAt,
		UpdatedAt:                 record.UpdatedAt,
	}
}

func issuerRecordFromRuntime(record phase1runtime.IssuerRecord) IssuerRecord {
	return IssuerRecord{
		IssuerID:                  record.IssuerID,
		DisplayName:               record.DisplayName,
		TrustState:                record.TrustState,
		AllowedTemplateIDs:        append([]string(nil), record.AllowedTemplateIDs...),
		VerificationKeyReferences: append([]string(nil), record.VerificationKeyReferences...),
		CreatedAt:                 record.CreatedAt,
		UpdatedAt:                 record.UpdatedAt,
	}
}

func issuerRecordToSQL(record IssuerRecord) phase1sql.IssuerRecord {
	return phase1sql.IssuerRecord{
		IssuerID:                  record.IssuerID,
		DisplayName:               record.DisplayName,
		TrustState:                record.TrustState,
		AllowedTemplateIDs:        append([]string(nil), record.AllowedTemplateIDs...),
		VerificationKeyReferences: append([]string(nil), record.VerificationKeyReferences...),
		CreatedAt:                 record.CreatedAt,
		UpdatedAt:                 record.UpdatedAt,
	}
}

func issuerRecordFromSQL(record phase1sql.IssuerRecord) IssuerRecord {
	return IssuerRecord{
		IssuerID:                  record.IssuerID,
		DisplayName:               record.DisplayName,
		TrustState:                record.TrustState,
		AllowedTemplateIDs:        append([]string(nil), record.AllowedTemplateIDs...),
		VerificationKeyReferences: append([]string(nil), record.VerificationKeyReferences...),
		CreatedAt:                 record.CreatedAt,
		UpdatedAt:                 record.UpdatedAt,
	}
}

func credentialRecordToRuntime(record CredentialRecord) phase1runtime.CredentialRecord {
	return phase1runtime.CredentialRecord{
		CredentialID:   record.CredentialID,
		IssuerID:       record.IssuerID,
		TemplateID:     record.TemplateID,
		ArtifactDigest: record.ArtifactDigest,
		ExpiresAt:      record.ExpiresAt,
		Status:         string(record.Status),
	}
}

func credentialRecordFromRuntime(record phase1runtime.CredentialRecord) CredentialRecord {
	return CredentialRecord{
		CredentialID:   record.CredentialID,
		IssuerID:       record.IssuerID,
		TemplateID:     record.TemplateID,
		ArtifactDigest: record.ArtifactDigest,
		ExpiresAt:      record.ExpiresAt,
		Status:         CredentialStatusSnapshot(record.Status),
	}
}

func credentialRecordToSQL(record CredentialRecord) phase1sql.CredentialRecord {
	return phase1sql.CredentialRecord{
		CredentialID:   record.CredentialID,
		IssuerID:       record.IssuerID,
		TemplateID:     record.TemplateID,
		ArtifactDigest: record.ArtifactDigest,
		ExpiresAt:      record.ExpiresAt,
		Status:         string(record.Status),
	}
}

func credentialRecordFromSQL(record phase1sql.CredentialRecord) CredentialRecord {
	return CredentialRecord{
		CredentialID:   record.CredentialID,
		IssuerID:       record.IssuerID,
		TemplateID:     record.TemplateID,
		ArtifactDigest: record.ArtifactDigest,
		ExpiresAt:      record.ExpiresAt,
		Status:         CredentialStatusSnapshot(record.Status),
	}
}

func verificationRequestRecordToRuntime(record VerificationRequestRecord) phase1runtime.VerificationRequestRecord {
	return phase1runtime.VerificationRequestRecord{
		VerificationID:            record.VerificationID,
		VerifierID:                record.VerifierID,
		SubmittedCredentialDigest: record.SubmittedCredentialDigest,
		CredentialID:              record.CredentialID,
		PolicyID:                  record.PolicyID,
		RequestedAt:               record.RequestedAt,
		Actor:                     actorToRuntime(record.Actor),
		IdempotencyKey:            record.IdempotencyKey,
	}
}

func verificationRequestRecordFromRuntime(record phase1runtime.VerificationRequestRecord) VerificationRequestRecord {
	return VerificationRequestRecord{
		VerificationID:            record.VerificationID,
		VerifierID:                record.VerifierID,
		SubmittedCredentialDigest: record.SubmittedCredentialDigest,
		CredentialID:              record.CredentialID,
		PolicyID:                  record.PolicyID,
		RequestedAt:               record.RequestedAt,
		Actor:                     actorFromRuntime(record.Actor),
		IdempotencyKey:            record.IdempotencyKey,
	}
}

func verificationRequestRecordToSQL(record VerificationRequestRecord) phase1sql.VerificationRequestRecord {
	return phase1sql.VerificationRequestRecord{
		VerificationID:            record.VerificationID,
		VerifierID:                record.VerifierID,
		SubmittedCredentialDigest: record.SubmittedCredentialDigest,
		CredentialID:              record.CredentialID,
		PolicyID:                  record.PolicyID,
		RequestedAt:               record.RequestedAt,
		Actor:                     actorToSQL(record.Actor),
		IdempotencyKey:            record.IdempotencyKey,
	}
}

func verificationRequestRecordFromSQL(record phase1sql.VerificationRequestRecord) VerificationRequestRecord {
	return VerificationRequestRecord{
		VerificationID:            record.VerificationID,
		VerifierID:                record.VerifierID,
		SubmittedCredentialDigest: record.SubmittedCredentialDigest,
		CredentialID:              record.CredentialID,
		PolicyID:                  record.PolicyID,
		RequestedAt:               record.RequestedAt,
		Actor:                     actorFromSQL(record.Actor),
		IdempotencyKey:            record.IdempotencyKey,
	}
}

func verificationResultRecordToRuntime(record VerificationResultRecord) phase1runtime.VerificationResultRecord {
	return phase1runtime.VerificationResultRecord{
		VerificationID:   record.VerificationID,
		IssuerID:         record.IssuerID,
		Decision:         string(record.Decision),
		ReasonCodes:      append([]string(nil), record.ReasonCodes...),
		IssuerTrustState: record.IssuerTrustState,
		CredentialStatus: string(record.CredentialStatus),
		EvaluatedAt:      record.EvaluatedAt,
		ResponseVersion:  record.ResponseVersion,
	}
}

func verificationResultRecordFromRuntime(record phase1runtime.VerificationResultRecord) VerificationResultRecord {
	return VerificationResultRecord{
		VerificationID:   record.VerificationID,
		IssuerID:         record.IssuerID,
		Decision:         VerificationDecision(record.Decision),
		ReasonCodes:      append([]string(nil), record.ReasonCodes...),
		IssuerTrustState: record.IssuerTrustState,
		CredentialStatus: CredentialStatusSnapshot(record.CredentialStatus),
		EvaluatedAt:      record.EvaluatedAt,
		ResponseVersion:  record.ResponseVersion,
	}
}

func verificationResultRecordToSQL(record VerificationResultRecord) phase1sql.VerificationResultRecord {
	return phase1sql.VerificationResultRecord{
		VerificationID:   record.VerificationID,
		IssuerID:         record.IssuerID,
		Decision:         string(record.Decision),
		ReasonCodes:      append([]string(nil), record.ReasonCodes...),
		IssuerTrustState: record.IssuerTrustState,
		CredentialStatus: string(record.CredentialStatus),
		EvaluatedAt:      record.EvaluatedAt,
		ResponseVersion:  record.ResponseVersion,
	}
}

func verificationResultRecordFromSQL(record phase1sql.VerificationResultRecord) VerificationResultRecord {
	return VerificationResultRecord{
		VerificationID:   record.VerificationID,
		IssuerID:         record.IssuerID,
		Decision:         VerificationDecision(record.Decision),
		ReasonCodes:      append([]string(nil), record.ReasonCodes...),
		IssuerTrustState: record.IssuerTrustState,
		CredentialStatus: CredentialStatusSnapshot(record.CredentialStatus),
		EvaluatedAt:      record.EvaluatedAt,
		ResponseVersion:  record.ResponseVersion,
	}
}

func auditRecordToRuntime(record AuditRecord) phase1runtime.AuditRecord {
	return phase1runtime.AuditRecord{
		AuditID:        record.AuditID,
		Actor:          actorToRuntime(record.Actor),
		Action:         record.Action,
		ResourceType:   record.ResourceType,
		ResourceID:     record.ResourceID,
		RequestID:      record.RequestID,
		IdempotencyKey: record.IdempotencyKey,
		Outcome:        record.Outcome,
		OccurredAt:     record.OccurredAt,
		ServiceName:    record.ServiceName,
	}
}

func auditRecordFromRuntime(record phase1runtime.AuditRecord) AuditRecord {
	return AuditRecord{
		AuditID:        record.AuditID,
		Actor:          actorFromRuntime(record.Actor),
		Action:         record.Action,
		ResourceType:   record.ResourceType,
		ResourceID:     record.ResourceID,
		RequestID:      record.RequestID,
		IdempotencyKey: record.IdempotencyKey,
		Outcome:        record.Outcome,
		OccurredAt:     record.OccurredAt,
		ServiceName:    record.ServiceName,
	}
}

func auditRecordToSQL(record AuditRecord) phase1sql.AuditRecord {
	return phase1sql.AuditRecord{
		AuditID:        record.AuditID,
		Actor:          actorToSQL(record.Actor),
		Action:         record.Action,
		ResourceType:   record.ResourceType,
		ResourceID:     record.ResourceID,
		RequestID:      record.RequestID,
		IdempotencyKey: record.IdempotencyKey,
		Outcome:        record.Outcome,
		OccurredAt:     record.OccurredAt,
		ServiceName:    record.ServiceName,
	}
}

func auditRecordFromSQL(record phase1sql.AuditRecord) AuditRecord {
	return AuditRecord{
		AuditID:        record.AuditID,
		Actor:          actorFromSQL(record.Actor),
		Action:         record.Action,
		ResourceType:   record.ResourceType,
		ResourceID:     record.ResourceID,
		RequestID:      record.RequestID,
		IdempotencyKey: record.IdempotencyKey,
		Outcome:        record.Outcome,
		OccurredAt:     record.OccurredAt,
		ServiceName:    record.ServiceName,
	}
}

func idempotencyRecordToRuntime(record IdempotencyRecord) phase1runtime.IdempotencyRecord {
	return phase1runtime.IdempotencyRecord{
		Operation:            record.Operation,
		CallerPrincipalID:    record.CallerPrincipalID,
		CallerOrganizationID: record.CallerOrganizationID,
		CallerActorType:      record.CallerActorType,
		IdempotencyKey:       record.IdempotencyKey,
		RequestFingerprint:   record.RequestFingerprint,
		State:                record.State,
		ResponseStatusCode:   record.ResponseStatusCode,
		ResourceType:         record.ResourceType,
		ResourceID:           record.ResourceID,
		Location:             record.Location,
		ResponseBody:         append([]byte(nil), record.ResponseBody...),
		CreatedAt:            record.CreatedAt,
		UpdatedAt:            record.UpdatedAt,
	}
}

func idempotencyRecordFromRuntime(record phase1runtime.IdempotencyRecord) IdempotencyRecord {
	return IdempotencyRecord{
		Operation:            record.Operation,
		CallerPrincipalID:    record.CallerPrincipalID,
		CallerOrganizationID: record.CallerOrganizationID,
		CallerActorType:      record.CallerActorType,
		IdempotencyKey:       record.IdempotencyKey,
		RequestFingerprint:   record.RequestFingerprint,
		State:                record.State,
		ResponseStatusCode:   record.ResponseStatusCode,
		ResourceType:         record.ResourceType,
		ResourceID:           record.ResourceID,
		Location:             record.Location,
		ResponseBody:         append([]byte(nil), record.ResponseBody...),
		CreatedAt:            record.CreatedAt,
		UpdatedAt:            record.UpdatedAt,
	}
}

func idempotencyRecordToSQL(record IdempotencyRecord) phase1sql.IdempotencyRecord {
	return phase1sql.IdempotencyRecord{
		Operation:            record.Operation,
		CallerPrincipalID:    record.CallerPrincipalID,
		CallerOrganizationID: record.CallerOrganizationID,
		CallerActorType:      record.CallerActorType,
		IdempotencyKey:       record.IdempotencyKey,
		RequestFingerprint:   record.RequestFingerprint,
		State:                record.State,
		ResponseStatusCode:   record.ResponseStatusCode,
		ResourceType:         record.ResourceType,
		ResourceID:           record.ResourceID,
		Location:             record.Location,
		ResponseBody:         append([]byte(nil), record.ResponseBody...),
		CreatedAt:            record.CreatedAt,
		UpdatedAt:            record.UpdatedAt,
	}
}

func idempotencyRecordFromSQL(record phase1sql.IdempotencyRecord) IdempotencyRecord {
	return IdempotencyRecord{
		Operation:            record.Operation,
		CallerPrincipalID:    record.CallerPrincipalID,
		CallerOrganizationID: record.CallerOrganizationID,
		CallerActorType:      record.CallerActorType,
		IdempotencyKey:       record.IdempotencyKey,
		RequestFingerprint:   record.RequestFingerprint,
		State:                record.State,
		ResponseStatusCode:   record.ResponseStatusCode,
		ResourceType:         record.ResourceType,
		ResourceID:           record.ResourceID,
		Location:             record.Location,
		ResponseBody:         append([]byte(nil), record.ResponseBody...),
		CreatedAt:            record.CreatedAt,
		UpdatedAt:            record.UpdatedAt,
	}
}

func idempotencyReservationResultFromRuntime(result phase1runtime.IdempotencyReservationResult) IdempotencyReservationResult {
	return IdempotencyReservationResult{
		Outcome: result.Outcome,
		Record:  idempotencyRecordFromRuntime(result.Record),
	}
}

func idempotencyReservationResultFromSQL(result phase1sql.IdempotencyReservationResult) IdempotencyReservationResult {
	return IdempotencyReservationResult{
		Outcome: result.Outcome,
		Record:  idempotencyRecordFromSQL(result.Record),
	}
}

func actorToRuntime(record authctx.Attribution) phase1runtime.Actor {
	return phase1runtime.Actor{
		PrincipalID:             record.PrincipalID,
		OrganizationID:          record.OrganizationID,
		ActorType:               string(record.ActorType),
		Scopes:                  append([]string(nil), record.Scopes...),
		AuthenticationReference: record.AuthenticationReference,
	}
}

func actorFromRuntime(record phase1runtime.Actor) authctx.Attribution {
	return authctx.Attribution{
		PrincipalID:             record.PrincipalID,
		OrganizationID:          record.OrganizationID,
		ActorType:               authctx.ActorType(record.ActorType),
		Scopes:                  append([]string(nil), record.Scopes...),
		AuthenticationReference: record.AuthenticationReference,
	}
}

func actorToSQL(record authctx.Attribution) phase1sql.Actor {
	return phase1sql.Actor{
		PrincipalID:             record.PrincipalID,
		OrganizationID:          record.OrganizationID,
		ActorType:               string(record.ActorType),
		Scopes:                  append([]string(nil), record.Scopes...),
		AuthenticationReference: record.AuthenticationReference,
	}
}

func actorFromSQL(record phase1sql.Actor) authctx.Attribution {
	return authctx.Attribution{
		PrincipalID:             record.PrincipalID,
		OrganizationID:          record.OrganizationID,
		ActorType:               authctx.ActorType(record.ActorType),
		Scopes:                  append([]string(nil), record.Scopes...),
		AuthenticationReference: record.AuthenticationReference,
	}
}
