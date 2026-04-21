package phase1

import (
	"context"
	"errors"
	"time"

	"github.com/Emiloart/HDIP/packages/go/foundation/authctx"
	phase1runtime "github.com/Emiloart/HDIP/services/internal/phase1runtime"
)

type RuntimeStore struct {
	runtime *phase1runtime.Store
}

func OpenRuntimeStore(path string) (*RuntimeStore, error) {
	store, err := phase1runtime.Open(path)
	if err != nil {
		return nil, err
	}

	return NewRuntimeStore(store), nil
}

func NewRuntimeStore(runtimeStore *phase1runtime.Store) *RuntimeStore {
	return &RuntimeStore{runtime: runtimeStore}
}

func (s *RuntimeStore) Close() error {
	if s == nil {
		return nil
	}

	return s.runtime.Close()
}

func (s *RuntimeStore) GetIssuerRecord(ctx context.Context, issuerID string) (IssuerRecord, error) {
	record, err := s.runtime.GetIssuerRecord(ctx, issuerID)
	if err != nil {
		return IssuerRecord{}, translateRuntimeError(err)
	}

	return issuerRecordFromRuntime(record), nil
}

func (s *RuntimeStore) NextCredentialID(ctx context.Context, templateID string) (string, error) {
	return s.runtime.NextCredentialID(ctx, templateID)
}

func (s *RuntimeStore) CreateCredentialRecord(ctx context.Context, record CredentialRecord) error {
	return translateRuntimeError(s.runtime.CreateCredentialRecord(ctx, credentialRecordToRuntime(record)))
}

func (s *RuntimeStore) GetCredentialRecord(ctx context.Context, credentialID string) (CredentialRecord, error) {
	record, err := s.runtime.GetCredentialRecord(ctx, credentialID)
	if err != nil {
		return CredentialRecord{}, translateRuntimeError(err)
	}

	return credentialRecordFromRuntime(record), nil
}

func (s *RuntimeStore) UpdateCredentialStatus(
	ctx context.Context,
	credentialID string,
	status CredentialStatus,
	statusUpdatedAt time.Time,
	supersededByCredentialID string,
) error {
	return translateRuntimeError(
		s.runtime.UpdateCredentialStatus(ctx, credentialID, string(status), statusUpdatedAt, supersededByCredentialID),
	)
}

func (s *RuntimeStore) AppendAuditRecord(ctx context.Context, record AuditRecord) error {
	return translateRuntimeError(s.runtime.AppendAuditRecord(ctx, auditRecordToRuntime(record)))
}

func (s *RuntimeStore) SeedIssuerRecord(record IssuerRecord) error {
	return translateRuntimeError(s.runtime.UpsertIssuerRecord(context.Background(), issuerRecordToRuntime(record)))
}

func (s *RuntimeStore) DeleteIssuerRecord(issuerID string) error {
	return translateRuntimeError(s.runtime.DeleteIssuerRecord(context.Background(), issuerID))
}

func (s *RuntimeStore) AuditRecords() ([]AuditRecord, error) {
	records, err := s.runtime.ListAuditRecords(context.Background())
	if err != nil {
		return nil, translateRuntimeError(err)
	}

	result := make([]AuditRecord, 0, len(records))
	for _, record := range records {
		result = append(result, auditRecordFromRuntime(record))
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

func credentialArtifactToRuntime(record *CredentialArtifact) *phase1runtime.CredentialArtifact {
	if record == nil {
		return nil
	}

	return &phase1runtime.CredentialArtifact{
		Kind:      record.Kind,
		MediaType: record.MediaType,
		Value:     record.Value,
	}
}

func credentialArtifactFromRuntime(record *phase1runtime.CredentialArtifact) *CredentialArtifact {
	if record == nil {
		return nil
	}

	return &CredentialArtifact{
		Kind:      record.Kind,
		MediaType: record.MediaType,
		Value:     record.Value,
	}
}

func credentialRecordToRuntime(record CredentialRecord) phase1runtime.CredentialRecord {
	return phase1runtime.CredentialRecord{
		CredentialID:             record.CredentialID,
		IssuerID:                 record.IssuerID,
		TemplateID:               record.TemplateID,
		SubjectReference:         record.SubjectReference,
		Claims:                   phase1runtime.KYCClaims(record.Claims),
		ArtifactDigest:           record.ArtifactDigest,
		CredentialArtifact:       credentialArtifactToRuntime(record.CredentialArtifact),
		ArtifactReference:        record.ArtifactReference,
		Status:                   string(record.Status),
		StatusReference:          record.StatusReference,
		IssuedAt:                 record.IssuedAt,
		ExpiresAt:                record.ExpiresAt,
		StatusUpdatedAt:          record.StatusUpdatedAt,
		SupersededByCredentialID: record.SupersededByCredentialID,
	}
}

func credentialRecordFromRuntime(record phase1runtime.CredentialRecord) CredentialRecord {
	return CredentialRecord{
		CredentialID:             record.CredentialID,
		IssuerID:                 record.IssuerID,
		TemplateID:               record.TemplateID,
		SubjectReference:         record.SubjectReference,
		Claims:                   KYCClaims(record.Claims),
		ArtifactDigest:           record.ArtifactDigest,
		CredentialArtifact:       credentialArtifactFromRuntime(record.CredentialArtifact),
		ArtifactReference:        record.ArtifactReference,
		Status:                   CredentialStatus(record.Status),
		StatusReference:          record.StatusReference,
		IssuedAt:                 record.IssuedAt,
		ExpiresAt:                record.ExpiresAt,
		StatusUpdatedAt:          record.StatusUpdatedAt,
		SupersededByCredentialID: record.SupersededByCredentialID,
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
