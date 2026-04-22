package phase1

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	phase1sql "github.com/Emiloart/HDIP/services/internal/phase1sql"
)

func TestApplyBootstrapFileUpsertsIssuerRecordsAndPreservesCreatedAt(t *testing.T) {
	store, err := OpenRuntimeStore(filepath.Join(t.TempDir(), "trust-phase1-state.json"))
	if err != nil {
		t.Fatalf("open runtime store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	bootstrapPath := filepath.Join(t.TempDir(), "trust-bootstrap.json")
	if err := os.WriteFile(bootstrapPath, []byte(`{
  "issuers": [
    {
      "issuerId": "did:web:issuer.hdip.dev",
      "displayName": "HDIP Passport Issuer",
      "trustState": "active",
      "allowedTemplateIds": ["hdip-passport-basic"],
      "verificationKeyReferences": ["key:issuer.hdip.dev:2026-04"]
    }
  ]
}`), 0o600); err != nil {
		t.Fatalf("write bootstrap file: %v", err)
	}

	firstAppliedAt := time.Date(2026, time.April, 22, 10, 0, 0, 0, time.UTC)
	if _, err := ApplyBootstrapFile(context.Background(), store, bootstrapPath, firstAppliedAt); err != nil {
		t.Fatalf("apply initial bootstrap file: %v", err)
	}

	if err := os.WriteFile(bootstrapPath, []byte(`{
  "issuers": [
    {
      "issuerId": "did:web:issuer.hdip.dev",
      "displayName": "HDIP Passport Issuer",
      "trustState": "suspended",
      "allowedTemplateIds": ["hdip-passport-basic"],
      "verificationKeyReferences": ["key:issuer.hdip.dev:2026-05"]
    }
  ]
}`), 0o600); err != nil {
		t.Fatalf("rewrite bootstrap file: %v", err)
	}

	secondAppliedAt := firstAppliedAt.Add(45 * time.Minute)
	if _, err := ApplyBootstrapFile(context.Background(), store, bootstrapPath, secondAppliedAt); err != nil {
		t.Fatalf("apply updated bootstrap file: %v", err)
	}

	record, err := store.GetIssuerRecord(context.Background(), "did:web:issuer.hdip.dev")
	if err != nil {
		t.Fatalf("get issuer record: %v", err)
	}

	if record.TrustState != "suspended" {
		t.Fatalf("expected suspended trust state, got %+v", record)
	}

	if len(record.VerificationKeyReferences) != 1 || record.VerificationKeyReferences[0] != "key:issuer.hdip.dev:2026-05" {
		t.Fatalf("unexpected verification key references: %+v", record.VerificationKeyReferences)
	}

	if !record.CreatedAt.Equal(firstAppliedAt) {
		t.Fatalf("expected createdAt %s, got %s", firstAppliedAt, record.CreatedAt)
	}

	if !record.UpdatedAt.Equal(secondAppliedAt) {
		t.Fatalf("expected updatedAt %s, got %s", secondAppliedAt, record.UpdatedAt)
	}

	audits, err := store.ListAuditRecords(context.Background())
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}

	if len(audits) != 2 {
		t.Fatalf("expected 2 audit records, got %d", len(audits))
	}

	for _, audit := range audits {
		if audit.Action != bootstrapAction || audit.ServiceName != "trust-registry" || audit.Actor.AuthenticationReference != "bootstrap:trust-bootstrap.json" {
			t.Fatalf("unexpected audit record: %+v", audit)
		}
	}
}

func TestApplyBootstrapDocumentPersistsToPrimarySQLPathWhenConfigured(t *testing.T) {
	dsn := os.Getenv("HDIP_PHASE1_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HDIP_PHASE1_TEST_DATABASE_URL is not set")
	}

	sqlStore, err := phase1sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open sql store: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlStore.Close()
	})

	store, err := OpenStore(StoreOptions{
		DatabaseDriver: "pgx",
		DatabaseURL:    dsn,
	})
	if err != nil {
		t.Fatalf("open trust runtime store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	appliedAt := time.Date(2026, time.April, 22, 11, 0, 0, 0, time.UTC)
	issuerID := "did:web:issuer-sql-bootstrap.hdip.dev"
	result, err := ApplyBootstrapDocument(context.Background(), store, "sql-bootstrap.json", BootstrapDocument{
		Issuers: []BootstrapIssuerRecord{
			{
				IssuerID:                  issuerID,
				DisplayName:               "HDIP Passport Issuer",
				TrustState:                "active",
				AllowedTemplateIDs:        []string{"hdip-passport-basic"},
				VerificationKeyReferences: []string{"key:issuer.hdip.dev:2026-04"},
			},
		},
	}, appliedAt)
	if err != nil {
		t.Fatalf("apply sql bootstrap document: %v", err)
	}

	if result.Applied != 1 {
		t.Fatalf("expected 1 applied issuer, got %d", result.Applied)
	}

	record, err := sqlStore.GetIssuerRecord(context.Background(), issuerID)
	if err != nil {
		t.Fatalf("load issuer record from sql store: %v", err)
	}

	if record.TrustState != "active" || !record.CreatedAt.Equal(appliedAt) || !record.UpdatedAt.Equal(appliedAt) {
		t.Fatalf("unexpected stored issuer record: %+v", record)
	}

	audits, err := sqlStore.ListAuditRecords(context.Background())
	if err != nil {
		t.Fatalf("list audit records from sql store: %v", err)
	}

	foundAudit := false
	for _, audit := range audits {
		if audit.ResourceID == issuerID && audit.Action == bootstrapAction {
			foundAudit = true
			break
		}
	}

	if !foundAudit {
		t.Fatal("expected trust bootstrap audit record in sql store")
	}
}
