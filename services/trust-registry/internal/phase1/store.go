package phase1

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	phase1runtime "github.com/Emiloart/HDIP/services/internal/phase1runtime"
	phase1sql "github.com/Emiloart/HDIP/services/internal/phase1sql"
)

var ErrRecordNotFound = errors.New("phase1 trust record not found")

type IssuerRecord struct {
	IssuerID                  string
	DisplayName               string
	TrustState                string
	AllowedTemplateIDs        []string
	VerificationKeyReferences []string
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

type Actor struct {
	PrincipalID             string
	OrganizationID          string
	ActorType               string
	Scopes                  []string
	AuthenticationReference string
}

type AuditRecord struct {
	AuditID        string
	Actor          Actor
	Action         string
	ResourceType   string
	ResourceID     string
	RequestID      string
	IdempotencyKey string
	Outcome        string
	OccurredAt     time.Time
	ServiceName    string
}

type StoreOptions struct {
	RuntimeMode           string
	DatabaseDriver        string
	DatabaseURL           string
	TransitionalStatePath string
}

type RuntimeStore struct {
	mode   string
	legacy *phase1runtime.Store
	sql    *phase1sql.Store
}

const (
	RuntimeModeSQLPrimary       = "sql-primary"
	RuntimeModeTransitionalJSON = "transitional-json"
)

func OpenStore(options StoreOptions) (*RuntimeStore, error) {
	switch normalizeRuntimeMode(options.RuntimeMode) {
	case RuntimeModeSQLPrimary:
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

		return &RuntimeStore{mode: RuntimeModeSQLPrimary, sql: store}, nil
	case RuntimeModeTransitionalJSON:
		if strings.TrimSpace(options.TransitionalStatePath) == "" {
			return nil, fmt.Errorf("transitional-json runtime requires HDIP_PHASE1_TRANSITIONAL_STATE_PATH")
		}

		return OpenRuntimeStore(options.TransitionalStatePath)
	default:
		return nil, fmt.Errorf("unsupported phase1 runtime mode %q", options.RuntimeMode)
	}
}

func OpenRuntimeStore(path string) (*RuntimeStore, error) {
	store, err := phase1runtime.Open(path)
	if err != nil {
		return nil, err
	}

	return &RuntimeStore{mode: RuntimeModeTransitionalJSON, legacy: store}, nil
}

func NewRuntimeStore(store *phase1runtime.Store) *RuntimeStore {
	return &RuntimeStore{mode: RuntimeModeTransitionalJSON, legacy: store}
}

func NewSQLRuntimeStore(store *phase1sql.Store) *RuntimeStore {
	return &RuntimeStore{mode: RuntimeModeSQLPrimary, sql: store}
}

func (s *RuntimeStore) RuntimeMode() string {
	if s == nil || strings.TrimSpace(s.mode) == "" {
		return RuntimeModeSQLPrimary
	}

	return s.mode
}

func (s *RuntimeStore) CheckReadiness(ctx context.Context, requireTrustBootstrap bool) error {
	if s == nil {
		return fmt.Errorf("phase1 runtime store is required")
	}

	if s.sql != nil {
		if err := s.sql.RequireSchema(ctx); err != nil {
			return err
		}
		if requireTrustBootstrap {
			if err := s.sql.RequireTrustBootstrap(ctx); err != nil {
				return err
			}
		}
	}

	return nil
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

func (s *RuntimeStore) UpsertIssuerRecord(ctx context.Context, record IssuerRecord) error {
	if s.sql != nil {
		return translateSQLError(s.sql.UpsertIssuerRecord(ctx, issuerRecordToSQL(record)))
	}

	return translateRuntimeError(s.legacy.UpsertIssuerRecord(ctx, issuerRecordToRuntime(record)))
}

func (s *RuntimeStore) SeedIssuerRecord(record IssuerRecord) error {
	return s.UpsertIssuerRecord(context.Background(), record)
}

func (s *RuntimeStore) DeleteIssuerRecord(issuerID string) error {
	if s.sql != nil {
		return translateSQLError(s.sql.DeleteIssuerRecord(context.Background(), issuerID))
	}

	return translateRuntimeError(s.legacy.DeleteIssuerRecord(context.Background(), issuerID))
}

func (s *RuntimeStore) AppendAuditRecord(ctx context.Context, record AuditRecord) error {
	if s.sql != nil {
		return translateSQLError(s.sql.AppendAuditRecord(ctx, auditRecordToSQL(record)))
	}

	return translateRuntimeError(s.legacy.AppendAuditRecord(ctx, auditRecordToRuntime(record)))
}

func (s *RuntimeStore) ListAuditRecords(ctx context.Context) ([]AuditRecord, error) {
	if s.sql != nil {
		records, err := s.sql.ListAuditRecords(ctx)
		if err != nil {
			return nil, translateSQLError(err)
		}

		translated := make([]AuditRecord, 0, len(records))
		for _, record := range records {
			translated = append(translated, auditRecordFromSQL(record))
		}

		return translated, nil
	}

	records, err := s.legacy.ListAuditRecords(ctx)
	if err != nil {
		return nil, translateRuntimeError(err)
	}

	translated := make([]AuditRecord, 0, len(records))
	for _, record := range records {
		translated = append(translated, auditRecordFromRuntime(record))
	}

	return translated, nil
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

func normalizeRuntimeMode(mode string) string {
	normalizedMode := strings.TrimSpace(mode)
	if normalizedMode == "" {
		return RuntimeModeSQLPrimary
	}

	return normalizedMode
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

func actorFromRuntime(record phase1runtime.Actor) Actor {
	return Actor{
		PrincipalID:             record.PrincipalID,
		OrganizationID:          record.OrganizationID,
		ActorType:               record.ActorType,
		Scopes:                  append([]string(nil), record.Scopes...),
		AuthenticationReference: record.AuthenticationReference,
	}
}

func actorToRuntime(record Actor) phase1runtime.Actor {
	return phase1runtime.Actor{
		PrincipalID:             record.PrincipalID,
		OrganizationID:          record.OrganizationID,
		ActorType:               record.ActorType,
		Scopes:                  append([]string(nil), record.Scopes...),
		AuthenticationReference: record.AuthenticationReference,
	}
}

func actorFromSQL(record phase1sql.Actor) Actor {
	return Actor{
		PrincipalID:             record.PrincipalID,
		OrganizationID:          record.OrganizationID,
		ActorType:               record.ActorType,
		Scopes:                  append([]string(nil), record.Scopes...),
		AuthenticationReference: record.AuthenticationReference,
	}
}

func actorToSQL(record Actor) phase1sql.Actor {
	return phase1sql.Actor{
		PrincipalID:             record.PrincipalID,
		OrganizationID:          record.OrganizationID,
		ActorType:               record.ActorType,
		Scopes:                  append([]string(nil), record.Scopes...),
		AuthenticationReference: record.AuthenticationReference,
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
