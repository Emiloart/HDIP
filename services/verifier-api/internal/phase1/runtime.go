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

func (s *RuntimeStore) GetIssuerTrustRecord(ctx context.Context, issuerID string) (IssuerTrustRecord, error) {
	record, err := s.runtime.GetIssuerRecord(ctx, issuerID)
	if err != nil {
		return IssuerTrustRecord{}, translateRuntimeError(err)
	}

	return issuerTrustRecordFromRuntime(record), nil
}

func (s *RuntimeStore) GetCredentialRecord(ctx context.Context, credentialID string) (CredentialRecord, error) {
	record, err := s.runtime.GetCredentialRecord(ctx, credentialID)
	if err != nil {
		return CredentialRecord{}, translateRuntimeError(err)
	}

	return credentialRecordFromRuntime(record), nil
}

func (s *RuntimeStore) GetCredentialRecordByArtifactDigest(ctx context.Context, artifactDigest string) (CredentialRecord, error) {
	record, err := s.runtime.GetCredentialRecordByArtifactDigest(ctx, artifactDigest)
	if err != nil {
		return CredentialRecord{}, translateRuntimeError(err)
	}

	return credentialRecordFromRuntime(record), nil
}

func (s *RuntimeStore) NextVerificationID(ctx context.Context) (string, error) {
	return s.runtime.NextVerificationID(ctx)
}

func (s *RuntimeStore) CreateVerificationRequestRecord(ctx context.Context, record VerificationRequestRecord) error {
	return translateRuntimeError(s.runtime.CreateVerificationRequestRecord(ctx, verificationRequestRecordToRuntime(record)))
}

func (s *RuntimeStore) GetVerificationRequestRecord(ctx context.Context, verificationID string) (VerificationRequestRecord, error) {
	record, err := s.runtime.GetVerificationRequestRecord(ctx, verificationID)
	if err != nil {
		return VerificationRequestRecord{}, translateRuntimeError(err)
	}

	return verificationRequestRecordFromRuntime(record), nil
}

func (s *RuntimeStore) CreateVerificationResultRecord(ctx context.Context, record VerificationResultRecord) error {
	return translateRuntimeError(s.runtime.CreateVerificationResultRecord(ctx, verificationResultRecordToRuntime(record)))
}

func (s *RuntimeStore) GetVerificationResultRecord(ctx context.Context, verificationID string) (VerificationResultRecord, error) {
	record, err := s.runtime.GetVerificationResultRecord(ctx, verificationID)
	if err != nil {
		return VerificationResultRecord{}, translateRuntimeError(err)
	}

	return verificationResultRecordFromRuntime(record), nil
}

func (s *RuntimeStore) AppendAuditRecord(ctx context.Context, record AuditRecord) error {
	return translateRuntimeError(s.runtime.AppendAuditRecord(ctx, auditRecordToRuntime(record)))
}

func (s *RuntimeStore) CreateIdempotencyRecord(ctx context.Context, record IdempotencyRecord) error {
	return translateRuntimeError(s.runtime.CreateIdempotencyRecord(ctx, idempotencyRecordToRuntime(record)))
}

func (s *RuntimeStore) GetIdempotencyRecord(
	ctx context.Context,
	operation string,
	callerOrganizationID string,
	callerPrincipalID string,
	callerActorType string,
	idempotencyKey string,
) (IdempotencyRecord, error) {
	record, err := s.runtime.GetIdempotencyRecord(
		ctx,
		operation,
		callerOrganizationID,
		callerPrincipalID,
		callerActorType,
		idempotencyKey,
	)
	if err != nil {
		return IdempotencyRecord{}, translateRuntimeError(err)
	}

	return idempotencyRecordFromRuntime(record), nil
}

func (s *RuntimeStore) SeedIssuerRecord(record IssuerRecord) error {
	return translateRuntimeError(s.runtime.UpsertIssuerRecord(context.Background(), issuerRecordToRuntime(record)))
}

func (s *RuntimeStore) DeleteIssuerRecord(issuerID string) error {
	return translateRuntimeError(s.runtime.DeleteIssuerRecord(context.Background(), issuerID))
}

func (s *RuntimeStore) SeedCredentialRecord(record CredentialRecord) error {
	return translateRuntimeError(s.runtime.UpsertCredentialRecord(context.Background(), credentialRecordToRuntime(record)))
}

func (s *RuntimeStore) UpdateCredentialStatus(
	ctx context.Context,
	credentialID string,
	status CredentialStatusSnapshot,
	statusUpdatedAt time.Time,
	supersededByCredentialID string,
) error {
	return translateRuntimeError(
		s.runtime.UpdateCredentialStatus(ctx, credentialID, string(status), statusUpdatedAt, supersededByCredentialID),
	)
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

func (s *RuntimeStore) IdempotencyRecords() ([]IdempotencyRecord, error) {
	records, err := s.runtime.ListIdempotencyRecords(context.Background())
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

func issuerTrustRecordFromRuntime(record phase1runtime.IssuerRecord) IssuerTrustRecord {
	return IssuerTrustRecord{
		IssuerID:                  record.IssuerID,
		TrustState:                record.TrustState,
		AllowedTemplateIDs:        append([]string(nil), record.AllowedTemplateIDs...),
		VerificationKeyReferences: append([]string(nil), record.VerificationKeyReferences...),
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

func idempotencyRecordToRuntime(record IdempotencyRecord) phase1runtime.IdempotencyRecord {
	return phase1runtime.IdempotencyRecord{
		Operation:            record.Operation,
		CallerPrincipalID:    record.CallerPrincipalID,
		CallerOrganizationID: record.CallerOrganizationID,
		CallerActorType:      record.CallerActorType,
		IdempotencyKey:       record.IdempotencyKey,
		RequestFingerprint:   record.RequestFingerprint,
		ResponseStatusCode:   record.ResponseStatusCode,
		ResourceType:         record.ResourceType,
		ResourceID:           record.ResourceID,
		Location:             record.Location,
		ResponseBody:         append([]byte(nil), record.ResponseBody...),
		CreatedAt:            record.CreatedAt,
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
		ResponseStatusCode:   record.ResponseStatusCode,
		ResourceType:         record.ResourceType,
		ResourceID:           record.ResourceID,
		Location:             record.Location,
		ResponseBody:         append([]byte(nil), record.ResponseBody...),
		CreatedAt:            record.CreatedAt,
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
