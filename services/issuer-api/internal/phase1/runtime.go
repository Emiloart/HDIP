package phase1

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Emiloart/HDIP/packages/go/foundation/authctx"
	phase1sql "github.com/Emiloart/HDIP/services/internal/phase1sql"
)

type StoreOptions struct {
	DatabaseDriver string
	DatabaseURL    string
}

type RuntimeStore struct {
	sql *phase1sql.Store
}

const RuntimeModeSQLPrimary = "sql-primary"

func OpenStore(options StoreOptions) (*RuntimeStore, error) {
	if strings.TrimSpace(options.DatabaseURL) == "" {
		return nil, fmt.Errorf("phase1 sql-primary runtime requires HDIP_PHASE1_DATABASE_URL")
	}

	store, err := phase1sql.Open(options.DatabaseDriver, options.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := store.RequireTrustBootstrap(context.Background()); err != nil {
		_ = store.Close()
		return nil, err
	}

	return NewSQLRuntimeStore(store), nil
}

func NewSQLRuntimeStore(runtimeStore *phase1sql.Store) *RuntimeStore {
	return &RuntimeStore{sql: runtimeStore}
}

func (s *RuntimeStore) RuntimeMode() string {
	return RuntimeModeSQLPrimary
}

func (s *RuntimeStore) CheckReadiness(ctx context.Context, requireTrustBootstrap bool) error {
	if s == nil || s.sql == nil {
		return fmt.Errorf("phase1 runtime store is required")
	}

	if err := s.sql.RequireSchema(ctx); err != nil {
		return err
	}
	if requireTrustBootstrap {
		if err := s.sql.RequireTrustBootstrap(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *RuntimeStore) Close() error {
	if s == nil || s.sql == nil {
		return nil
	}

	return s.sql.Close()
}

func (s *RuntimeStore) GetIssuerRecord(ctx context.Context, issuerID string) (IssuerRecord, error) {
	record, err := s.sql.GetIssuerRecord(ctx, issuerID)
	if err != nil {
		return IssuerRecord{}, translateSQLError(err)
	}

	return issuerRecordFromSQL(record), nil
}

func (s *RuntimeStore) NextCredentialID(ctx context.Context, templateID string) (string, error) {
	return s.sql.NextCredentialID(ctx, templateID)
}

func (s *RuntimeStore) CreateCredentialRecord(ctx context.Context, record CredentialRecord) error {
	return translateSQLError(s.sql.CreateCredentialRecord(ctx, credentialRecordToSQL(record)))
}

func (s *RuntimeStore) GetCredentialRecord(ctx context.Context, credentialID string) (CredentialRecord, error) {
	record, err := s.sql.GetCredentialRecord(ctx, credentialID)
	if err != nil {
		return CredentialRecord{}, translateSQLError(err)
	}

	return credentialRecordFromSQL(record), nil
}

func (s *RuntimeStore) UpdateCredentialStatus(
	ctx context.Context,
	credentialID string,
	status CredentialStatus,
	statusUpdatedAt time.Time,
	supersededByCredentialID string,
) error {
	return translateSQLError(
		s.sql.UpdateCredentialStatus(ctx, credentialID, string(status), statusUpdatedAt, supersededByCredentialID),
	)
}

func (s *RuntimeStore) AppendAuditRecord(ctx context.Context, record AuditRecord) error {
	return translateSQLError(s.sql.AppendAuditRecord(ctx, auditRecordToSQL(record)))
}

func (s *RuntimeStore) CreateIdempotencyRecord(ctx context.Context, record IdempotencyRecord) error {
	return translateSQLError(s.sql.CreateIdempotencyRecord(ctx, idempotencyRecordToSQL(record)))
}

func (s *RuntimeStore) ReserveIdempotencyRecord(ctx context.Context, record IdempotencyRecord) (IdempotencyReservationResult, error) {
	result, err := s.sql.ReserveIdempotencyRecord(ctx, idempotencyRecordToSQL(record))
	if err != nil {
		return IdempotencyReservationResult{}, translateSQLError(err)
	}

	return idempotencyReservationResultFromSQL(result), nil
}

func (s *RuntimeStore) CompleteIdempotencyRecord(ctx context.Context, record IdempotencyRecord) error {
	return translateSQLError(s.sql.CompleteIdempotencyRecord(ctx, idempotencyRecordToSQL(record)))
}

func (s *RuntimeStore) ReleaseIdempotencyRecord(
	ctx context.Context,
	operation string,
	callerOrganizationID string,
	callerPrincipalID string,
	callerActorType string,
	idempotencyKey string,
) error {
	return translateSQLError(
		s.sql.ReleaseIdempotencyRecord(ctx, operation, callerOrganizationID, callerPrincipalID, callerActorType, idempotencyKey),
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
	record, err := s.sql.GetIdempotencyRecord(ctx, operation, callerOrganizationID, callerPrincipalID, callerActorType, idempotencyKey)
	if err != nil {
		return IdempotencyRecord{}, translateSQLError(err)
	}

	return idempotencyRecordFromSQL(record), nil
}

func (s *RuntimeStore) SeedIssuerRecord(record IssuerRecord) error {
	return translateSQLError(s.sql.UpsertIssuerRecord(context.Background(), issuerRecordToSQL(record)))
}

func (s *RuntimeStore) DeleteIssuerRecord(issuerID string) error {
	return translateSQLError(s.sql.DeleteIssuerRecord(context.Background(), issuerID))
}

func (s *RuntimeStore) AuditRecords() ([]AuditRecord, error) {
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

func (s *RuntimeStore) IdempotencyRecords() ([]IdempotencyRecord, error) {
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

func translateSQLError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, phase1sql.ErrRecordNotFound) {
		return ErrRecordNotFound
	}

	return err
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

func credentialArtifactToSQL(record *CredentialArtifact) *phase1sql.CredentialArtifact {
	if record == nil {
		return nil
	}

	return &phase1sql.CredentialArtifact{
		Kind:      record.Kind,
		MediaType: record.MediaType,
		Value:     record.Value,
	}
}

func credentialArtifactFromSQL(record *phase1sql.CredentialArtifact) *CredentialArtifact {
	if record == nil {
		return nil
	}

	return &CredentialArtifact{
		Kind:      record.Kind,
		MediaType: record.MediaType,
		Value:     record.Value,
	}
}

func credentialRecordToSQL(record CredentialRecord) phase1sql.CredentialRecord {
	return phase1sql.CredentialRecord{
		CredentialID:             record.CredentialID,
		IssuerID:                 record.IssuerID,
		TemplateID:               record.TemplateID,
		SubjectReference:         record.SubjectReference,
		Claims:                   phase1sql.KYCClaims(record.Claims),
		ArtifactDigest:           record.ArtifactDigest,
		CredentialArtifact:       credentialArtifactToSQL(record.CredentialArtifact),
		ArtifactReference:        record.ArtifactReference,
		Status:                   string(record.Status),
		StatusReference:          record.StatusReference,
		IssuedAt:                 record.IssuedAt,
		ExpiresAt:                record.ExpiresAt,
		StatusUpdatedAt:          record.StatusUpdatedAt,
		SupersededByCredentialID: record.SupersededByCredentialID,
	}
}

func credentialRecordFromSQL(record phase1sql.CredentialRecord) CredentialRecord {
	return CredentialRecord{
		CredentialID:             record.CredentialID,
		IssuerID:                 record.IssuerID,
		TemplateID:               record.TemplateID,
		SubjectReference:         record.SubjectReference,
		Claims:                   KYCClaims(record.Claims),
		ArtifactDigest:           record.ArtifactDigest,
		CredentialArtifact:       credentialArtifactFromSQL(record.CredentialArtifact),
		ArtifactReference:        record.ArtifactReference,
		Status:                   CredentialStatus(record.Status),
		StatusReference:          record.StatusReference,
		IssuedAt:                 record.IssuedAt,
		ExpiresAt:                record.ExpiresAt,
		StatusUpdatedAt:          record.StatusUpdatedAt,
		SupersededByCredentialID: record.SupersededByCredentialID,
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

func idempotencyReservationResultFromSQL(result phase1sql.IdempotencyReservationResult) IdempotencyReservationResult {
	return IdempotencyReservationResult{
		Outcome: result.Outcome,
		Record:  idempotencyRecordFromSQL(result.Record),
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
