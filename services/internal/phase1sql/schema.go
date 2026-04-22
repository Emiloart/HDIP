package phase1sql

import (
	"context"
	"embed"
	"fmt"
	"strings"
)

//go:embed migrations/*.sql
var migrationAssetsFS embed.FS

type migrationAsset struct {
	Version string
	Path    string
}

var migrationAssets = []migrationAsset{
	{
		Version: "0001",
		Path:    "migrations/0001_phase1_schema.sql",
	},
}

var requiredSchemaTables = []string{
	"phase1_sequences",
	"trust_registry_issuer_records",
	"issuer_api_credential_records",
	"verifier_api_verification_request_records",
	"verifier_api_verification_result_records",
	"phase1_audit_records",
	"phase1_idempotency_records",
}

var requiredSchemaIndexes = []string{
	"issuer_api_credential_records_artifact_digest_idx",
}

func MigrateUp(ctx context.Context, driverName string, dsn string) error {
	store, err := Connect(driverName, dsn)
	if err != nil {
		return err
	}
	defer store.Close()

	return store.MigrateUp(ctx)
}

func (s *Store) MigrateUp(ctx context.Context) error {
	if s == nil {
		return fmt.Errorf("phase1 sql store is required")
	}

	for _, asset := range migrationAssets {
		raw, err := migrationAssetsFS.ReadFile(asset.Path)
		if err != nil {
			return fmt.Errorf("read phase1 sql migration %s: %w", asset.Version, err)
		}

		if err := executeSQLStatements(ctx, s, string(raw)); err != nil {
			return fmt.Errorf("apply phase1 sql migration %s: %w", asset.Version, err)
		}
	}

	return nil
}

func (s *Store) RequireSchema(ctx context.Context) error {
	if s == nil {
		return fmt.Errorf("phase1 sql store is required")
	}

	missing, err := s.MissingSchemaObjects(ctx)
	if err != nil {
		return err
	}
	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf(
		"phase1 sql schema is not initialized: missing %s; run phase1sql migrate up",
		strings.Join(missing, ", "),
	)
}

func (s *Store) MissingSchemaObjects(ctx context.Context) ([]string, error) {
	if s == nil {
		return nil, fmt.Errorf("phase1 sql store is required")
	}

	missing := make([]string, 0)
	for _, tableName := range requiredSchemaTables {
		exists, err := s.schemaObjectExists(ctx, "table", tableName)
		if err != nil {
			return nil, err
		}
		if !exists {
			missing = append(missing, "table:"+tableName)
		}
	}

	for _, indexName := range requiredSchemaIndexes {
		exists, err := s.schemaObjectExists(ctx, "index", indexName)
		if err != nil {
			return nil, err
		}
		if !exists {
			missing = append(missing, "index:"+indexName)
		}
	}

	return missing, nil
}

func (s *Store) schemaObjectExists(ctx context.Context, objectType string, name string) (bool, error) {
	var query string
	var args []any

	switch s.dialect {
	case "postgres":
		switch objectType {
		case "table":
			query = `
SELECT EXISTS (
  SELECT 1
  FROM information_schema.tables
  WHERE table_schema = current_schema()
    AND table_name = ?
)`
			args = []any{name}
		case "index":
			query = `
SELECT EXISTS (
  SELECT 1
  FROM pg_indexes
  WHERE schemaname = current_schema()
    AND indexname = ?
)`
			args = []any{name}
		default:
			return false, fmt.Errorf("unsupported schema object type %q", objectType)
		}
	default:
		query = `
SELECT EXISTS (
  SELECT 1
  FROM sqlite_master
  WHERE type = ?
    AND name = ?
)`
		args = []any{objectType, name}
	}

	var exists bool
	if err := s.db.QueryRowContext(ctx, s.bind(query), args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("check phase1 sql schema object %s %s: %w", objectType, name, err)
	}

	return exists, nil
}

func executeSQLStatements(ctx context.Context, store *Store, raw string) error {
	for _, statement := range strings.Split(raw, ";") {
		trimmedStatement := strings.TrimSpace(statement)
		if trimmedStatement == "" {
			continue
		}

		if _, err := store.db.ExecContext(ctx, trimmedStatement); err != nil {
			return err
		}
	}

	return nil
}
