package phase1

import (
	"context"
	"errors"

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
}

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

func NewRuntimeStore(store *phase1runtime.Store) *RuntimeStore {
	return &RuntimeStore{legacy: store}
}

func NewSQLRuntimeStore(store *phase1sql.Store) *RuntimeStore {
	return &RuntimeStore{sql: store}
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

func issuerRecordFromRuntime(record phase1runtime.IssuerRecord) IssuerRecord {
	return IssuerRecord{
		IssuerID:                  record.IssuerID,
		DisplayName:               record.DisplayName,
		TrustState:                record.TrustState,
		AllowedTemplateIDs:        append([]string(nil), record.AllowedTemplateIDs...),
		VerificationKeyReferences: append([]string(nil), record.VerificationKeyReferences...),
	}
}

func issuerRecordToRuntime(record IssuerRecord) phase1runtime.IssuerRecord {
	return phase1runtime.IssuerRecord{
		IssuerID:                  record.IssuerID,
		DisplayName:               record.DisplayName,
		TrustState:                record.TrustState,
		AllowedTemplateIDs:        append([]string(nil), record.AllowedTemplateIDs...),
		VerificationKeyReferences: append([]string(nil), record.VerificationKeyReferences...),
	}
}

func issuerRecordFromSQL(record phase1sql.IssuerRecord) IssuerRecord {
	return IssuerRecord{
		IssuerID:                  record.IssuerID,
		DisplayName:               record.DisplayName,
		TrustState:                record.TrustState,
		AllowedTemplateIDs:        append([]string(nil), record.AllowedTemplateIDs...),
		VerificationKeyReferences: append([]string(nil), record.VerificationKeyReferences...),
	}
}

func issuerRecordToSQL(record IssuerRecord) phase1sql.IssuerRecord {
	return phase1sql.IssuerRecord{
		IssuerID:                  record.IssuerID,
		DisplayName:               record.DisplayName,
		TrustState:                record.TrustState,
		AllowedTemplateIDs:        append([]string(nil), record.AllowedTemplateIDs...),
		VerificationKeyReferences: append([]string(nil), record.VerificationKeyReferences...),
	}
}
