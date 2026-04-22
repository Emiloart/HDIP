package phase1sql

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestStoreCrossInstanceContinuityAndReservation(t *testing.T) {
	dsn := os.Getenv("HDIP_PHASE1_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HDIP_PHASE1_TEST_DATABASE_URL is not set")
	}

	storeA := openTestStore(t, dsn)
	storeB := openTestStore(t, dsn)
	resetTestTables(t, storeA)

	now := time.Date(2026, time.April, 21, 12, 0, 0, 0, time.UTC)
	issuer := IssuerRecord{
		IssuerID:                  "did:web:issuer.hdip.dev",
		DisplayName:               "HDIP Passport Issuer",
		TrustState:                "active",
		AllowedTemplateIDs:        []string{"hdip-passport-basic"},
		VerificationKeyReferences: []string{"key:issuer.hdip.dev:2026-04"},
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}
	if err := storeA.UpsertIssuerRecord(context.Background(), issuer); err != nil {
		t.Fatalf("upsert issuer record: %v", err)
	}

	credential := CredentialRecord{
		CredentialID:     "cred_hdip_passport_basic_001",
		IssuerID:         issuer.IssuerID,
		TemplateID:       "hdip-passport-basic",
		SubjectReference: "subject_ref_alex_example",
		Claims: KYCClaims{
			FullLegalName:      "Alex Example",
			DateOfBirth:        "1990-04-18",
			CountryOfResidence: "US",
			DocumentCountry:    "US",
			KYCLevel:           "standard",
			VerifiedAt:         now,
			ExpiresAt:          now.Add(365 * 24 * time.Hour),
		},
		ArtifactDigest: "artifact-digest-001",
		CredentialArtifact: &CredentialArtifact{
			Kind:      "phase1_opaque_artifact",
			MediaType: "application/vnd.hdip.phase1-opaque-artifact",
			Value:     "opaque-artifact:v1:test",
		},
		Status:          "active",
		StatusReference: "status:cred_hdip_passport_basic_001",
		IssuedAt:        now,
		ExpiresAt:       now.Add(365 * 24 * time.Hour),
		StatusUpdatedAt: now,
	}
	if err := storeA.CreateCredentialRecord(context.Background(), credential); err != nil {
		t.Fatalf("create credential record: %v", err)
	}

	loadedCredential, err := storeB.GetCredentialRecord(context.Background(), credential.CredentialID)
	if err != nil {
		t.Fatalf("load credential record from second store: %v", err)
	}

	if loadedCredential.CredentialID != credential.CredentialID || loadedCredential.Status != "active" {
		t.Fatalf("unexpected loaded credential: %+v", loadedCredential)
	}

	loadedByDigest, err := storeB.GetCredentialRecordByArtifactDigest(context.Background(), credential.ArtifactDigest)
	if err != nil {
		t.Fatalf("load credential by digest from second store: %v", err)
	}

	if loadedByDigest.CredentialID != credential.CredentialID {
		t.Fatalf("unexpected digest lookup record: %+v", loadedByDigest)
	}

	requestRecord := VerificationRequestRecord{
		VerificationID:            "verification_hdip_001",
		VerifierID:                "verifier_org_marketplace_alpha",
		SubmittedCredentialDigest: credential.ArtifactDigest,
		CredentialID:              credential.CredentialID,
		PolicyID:                  "kyc-passport-basic",
		RequestedAt:               now.Add(5 * time.Minute),
		Actor: Actor{
			PrincipalID:             "verifier_integrator_alpha",
			OrganizationID:          "verifier_org_marketplace_alpha",
			ActorType:               "verifier_integrator",
			Scopes:                  []string{"verifier.requests.create"},
			AuthenticationReference: "credential_verifier_001",
		},
		IdempotencyKey: "verify-1",
	}
	if err := storeA.CreateVerificationRequestRecord(context.Background(), requestRecord); err != nil {
		t.Fatalf("create verification request record: %v", err)
	}

	resultRecord := VerificationResultRecord{
		VerificationID:   requestRecord.VerificationID,
		IssuerID:         issuer.IssuerID,
		Decision:         "allow",
		ReasonCodes:      []string{"issuer_trusted", "credential_status_active", "template_match"},
		IssuerTrustState: "active",
		CredentialStatus: "active",
		EvaluatedAt:      now.Add(6 * time.Minute),
		ResponseVersion:  "2026.04",
	}
	if err := storeA.CreateVerificationResultRecord(context.Background(), resultRecord); err != nil {
		t.Fatalf("create verification result record: %v", err)
	}

	loadedRequest, err := storeB.GetVerificationRequestRecord(context.Background(), requestRecord.VerificationID)
	if err != nil {
		t.Fatalf("load verification request from second store: %v", err)
	}
	if loadedRequest.VerifierID != requestRecord.VerifierID {
		t.Fatalf("unexpected loaded verification request: %+v", loadedRequest)
	}

	loadedResult, err := storeB.GetVerificationResultRecord(context.Background(), resultRecord.VerificationID)
	if err != nil {
		t.Fatalf("load verification result from second store: %v", err)
	}
	if loadedResult.Decision != "allow" || loadedResult.IssuerTrustState != "active" {
		t.Fatalf("unexpected loaded verification result: %+v", loadedResult)
	}

	reserved, err := storeA.ReserveIdempotencyRecord(context.Background(), IdempotencyRecord{
		Operation:            "issuer.credentials.issue",
		CallerPrincipalID:    "issuer_operator_alex",
		CallerOrganizationID: issuer.IssuerID,
		CallerActorType:      "issuer_operator",
		IdempotencyKey:       "issue-1",
		RequestFingerprint:   "fingerprint-1",
		ResourceType:         "credential",
		CreatedAt:            now,
		UpdatedAt:            now,
	})
	if err != nil {
		t.Fatalf("reserve idempotency record: %v", err)
	}
	if reserved.Outcome != IdempotencyReservationReserved {
		t.Fatalf("expected reserved outcome, got %+v", reserved)
	}

	inProgress, err := storeB.ReserveIdempotencyRecord(context.Background(), IdempotencyRecord{
		Operation:            "issuer.credentials.issue",
		CallerPrincipalID:    "issuer_operator_alex",
		CallerOrganizationID: issuer.IssuerID,
		CallerActorType:      "issuer_operator",
		IdempotencyKey:       "issue-1",
		RequestFingerprint:   "fingerprint-1",
		ResourceType:         "credential",
		CreatedAt:            now,
		UpdatedAt:            now,
	})
	if err != nil {
		t.Fatalf("reserve idempotency record from second store: %v", err)
	}
	if inProgress.Outcome != IdempotencyReservationInProgress {
		t.Fatalf("expected in-progress outcome, got %+v", inProgress)
	}

	if err := storeA.CompleteIdempotencyRecord(context.Background(), IdempotencyRecord{
		Operation:            "issuer.credentials.issue",
		CallerPrincipalID:    "issuer_operator_alex",
		CallerOrganizationID: issuer.IssuerID,
		CallerActorType:      "issuer_operator",
		IdempotencyKey:       "issue-1",
		RequestFingerprint:   "fingerprint-1",
		State:                IdempotencyStateCompleted,
		ResponseStatusCode:   201,
		ResourceType:         "credential",
		ResourceID:           credential.CredentialID,
		Location:             "/v1/issuer/credentials/" + credential.CredentialID,
		ResponseBody:         []byte(`{"credentialId":"cred_hdip_passport_basic_001"}`),
		CreatedAt:            now,
		UpdatedAt:            now.Add(time.Minute),
	}); err != nil {
		t.Fatalf("complete idempotency record: %v", err)
	}

	replay, err := storeB.ReserveIdempotencyRecord(context.Background(), IdempotencyRecord{
		Operation:            "issuer.credentials.issue",
		CallerPrincipalID:    "issuer_operator_alex",
		CallerOrganizationID: issuer.IssuerID,
		CallerActorType:      "issuer_operator",
		IdempotencyKey:       "issue-1",
		RequestFingerprint:   "fingerprint-1",
		ResourceType:         "credential",
		CreatedAt:            now,
		UpdatedAt:            now,
	})
	if err != nil {
		t.Fatalf("reserve idempotency record for replay: %v", err)
	}
	if replay.Outcome != IdempotencyReservationReplay || replay.Record.ResourceID != credential.CredentialID {
		t.Fatalf("unexpected replay outcome: %+v", replay)
	}

	if err := storeA.UpdateCredentialStatus(context.Background(), credential.CredentialID, "revoked", now.Add(2*time.Minute), ""); err != nil {
		t.Fatalf("update credential status: %v", err)
	}

	updatedCredential, err := storeB.GetCredentialRecord(context.Background(), credential.CredentialID)
	if err != nil {
		t.Fatalf("reload credential after status transition: %v", err)
	}
	if updatedCredential.Status != "revoked" {
		t.Fatalf("expected revoked status, got %+v", updatedCredential)
	}

	if err := storeA.AppendAuditRecord(context.Background(), AuditRecord{
		AuditID:        "audit-1",
		Actor:          requestRecord.Actor,
		Action:         "issuer.credentials.issue",
		ResourceType:   "credential",
		ResourceID:     credential.CredentialID,
		RequestID:      "req-1",
		IdempotencyKey: "issue-1",
		Outcome:        "succeeded",
		OccurredAt:     now,
		ServiceName:    "issuer-api",
	}); err != nil {
		t.Fatalf("append audit record: %v", err)
	}

	audits, err := storeB.ListAuditRecords(context.Background())
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}
	if len(audits) != 1 || audits[0].ResourceID != credential.CredentialID {
		t.Fatalf("unexpected audits: %+v", audits)
	}
}

func TestExplicitMigrationAndTrustBootstrapLifecycle(t *testing.T) {
	dsn := os.Getenv("HDIP_PHASE1_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HDIP_PHASE1_TEST_DATABASE_URL is not set")
	}

	if err := MigrateUp(context.Background(), "pgx", dsn); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	store := openTestStore(t, dsn)
	resetTestTables(t, store)

	now := time.Date(2026, time.April, 22, 14, 0, 0, 0, time.UTC)
	result, err := ApplyTrustBootstrapDocument(context.Background(), store, "phase1sql-bootstrap.json", TrustBootstrapDocument{
		Issuers: []TrustBootstrapIssuerRecord{
			{
				IssuerID:                  "did:web:issuer.lifecycle.hdip.dev",
				DisplayName:               "HDIP Passport Issuer",
				TrustState:                "active",
				AllowedTemplateIDs:        []string{"hdip-passport-basic"},
				VerificationKeyReferences: []string{"key:issuer.hdip.dev:2026-04"},
			},
		},
	}, now)
	if err != nil {
		t.Fatalf("apply trust bootstrap document: %v", err)
	}

	if result.Applied != 1 {
		t.Fatalf("expected 1 applied issuer, got %d", result.Applied)
	}

	if err := store.RequireSchema(context.Background()); err != nil {
		t.Fatalf("require schema after migrate: %v", err)
	}

	record, err := store.GetIssuerRecord(context.Background(), "did:web:issuer.lifecycle.hdip.dev")
	if err != nil {
		t.Fatalf("load issuer record: %v", err)
	}
	if record.TrustState != "active" {
		t.Fatalf("unexpected issuer record: %+v", record)
	}

	audits, err := store.ListAuditRecords(context.Background())
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}
	if len(audits) != 1 || audits[0].Action != TrustBootstrapAction {
		t.Fatalf("unexpected audit records: %+v", audits)
	}
}

func openTestStore(t *testing.T, dsn string) *Store {
	t.Helper()

	if err := MigrateUp(context.Background(), "pgx", dsn); err != nil {
		t.Fatalf("migrate test store: %v", err)
	}

	store, err := Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	return store
}

func resetTestTables(t *testing.T, store *Store) {
	t.Helper()

	statements := []string{
		`DELETE FROM phase1_audit_records`,
		`DELETE FROM verifier_api_verification_result_records`,
		`DELETE FROM verifier_api_verification_request_records`,
		`DELETE FROM issuer_api_credential_records`,
		`DELETE FROM trust_registry_issuer_records`,
		`DELETE FROM phase1_idempotency_records`,
		`DELETE FROM phase1_sequences`,
	}

	for _, statement := range statements {
		if _, err := store.db.ExecContext(context.Background(), statement); err != nil {
			t.Fatalf("reset test tables with %q: %v", statement, err)
		}
	}
}
