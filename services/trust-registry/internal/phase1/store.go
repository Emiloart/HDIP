package phase1

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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

func NewSQLRuntimeStore(store *phase1sql.Store) *RuntimeStore {
	return &RuntimeStore{sql: store}
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

func (s *RuntimeStore) UpsertIssuerRecord(ctx context.Context, record IssuerRecord) error {
	return translateSQLError(s.sql.UpsertIssuerRecord(ctx, issuerRecordToSQL(record)))
}

func (s *RuntimeStore) SeedIssuerRecord(record IssuerRecord) error {
	return s.UpsertIssuerRecord(context.Background(), record)
}

func (s *RuntimeStore) DeleteIssuerRecord(issuerID string) error {
	return translateSQLError(s.sql.DeleteIssuerRecord(context.Background(), issuerID))
}

func (s *RuntimeStore) AppendAuditRecord(ctx context.Context, record AuditRecord) error {
	return translateSQLError(s.sql.AppendAuditRecord(ctx, auditRecordToSQL(record)))
}

func (s *RuntimeStore) ListAuditRecords(ctx context.Context) ([]AuditRecord, error) {
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

func translateSQLError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, phase1sql.ErrRecordNotFound) {
		return ErrRecordNotFound
	}

	return err
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
