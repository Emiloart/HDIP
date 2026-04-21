package phase1runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestStoreSharesCredentialAndTrustStateAcrossInstances(t *testing.T) {
	path := filepath.Join(t.TempDir(), "phase1-state.json")
	issuerRuntime, err := Open(path)
	if err != nil {
		t.Fatalf("open issuer runtime: %v", err)
	}
	verifierRuntime, err := Open(path)
	if err != nil {
		t.Fatalf("open verifier runtime: %v", err)
	}

	now := time.Date(2026, time.April, 21, 9, 0, 0, 0, time.UTC)
	issuerRecord := IssuerRecord{
		IssuerID:                  "did:web:issuer.hdip.dev",
		DisplayName:               "HDIP Passport Issuer",
		TrustState:                "active",
		AllowedTemplateIDs:        []string{"hdip-passport-basic"},
		VerificationKeyReferences: []string{"key:issuer.hdip.dev:2026-04"},
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}
	if err := issuerRuntime.UpsertIssuerRecord(context.Background(), issuerRecord); err != nil {
		t.Fatalf("seed issuer record: %v", err)
	}

	artifactValue := "opaque-artifact:v1:test-value"
	credentialRecord := CredentialRecord{
		CredentialID:    "cred_hdip_passport_basic_001",
		IssuerID:        issuerRecord.IssuerID,
		TemplateID:      "hdip-passport-basic",
		ArtifactDigest:  digest(artifactValue),
		Status:          "active",
		IssuedAt:        now,
		ExpiresAt:       now.Add(365 * 24 * time.Hour),
		StatusUpdatedAt: now,
	}
	if err := issuerRuntime.CreateCredentialRecord(context.Background(), credentialRecord); err != nil {
		t.Fatalf("create credential record: %v", err)
	}

	loadedIssuer, err := verifierRuntime.GetIssuerRecord(context.Background(), issuerRecord.IssuerID)
	if err != nil {
		t.Fatalf("load issuer record from second runtime: %v", err)
	}
	if loadedIssuer.TrustState != "active" {
		t.Fatalf("expected active issuer trust state, got %+v", loadedIssuer)
	}

	loadedCredential, err := verifierRuntime.GetCredentialRecordByArtifactDigest(context.Background(), credentialRecord.ArtifactDigest)
	if err != nil {
		t.Fatalf("load credential record from second runtime: %v", err)
	}
	if loadedCredential.CredentialID != credentialRecord.CredentialID {
		t.Fatalf("expected credential %q, got %+v", credentialRecord.CredentialID, loadedCredential)
	}

	if err := issuerRuntime.UpdateCredentialStatus(
		context.Background(),
		credentialRecord.CredentialID,
		"revoked",
		now.Add(time.Hour),
		"",
	); err != nil {
		t.Fatalf("update credential status: %v", err)
	}

	updatedCredential, err := verifierRuntime.GetCredentialRecord(context.Background(), credentialRecord.CredentialID)
	if err != nil {
		t.Fatalf("reload credential record from second runtime: %v", err)
	}
	if updatedCredential.Status != "revoked" {
		t.Fatalf("expected revoked credential status, got %+v", updatedCredential)
	}
}

func TestStorePersistsIdempotencyRecordsAcrossInstances(t *testing.T) {
	path := filepath.Join(t.TempDir(), "phase1-state.json")
	firstRuntime, err := Open(path)
	if err != nil {
		t.Fatalf("open first runtime: %v", err)
	}
	secondRuntime, err := Open(path)
	if err != nil {
		t.Fatalf("open second runtime: %v", err)
	}

	record := IdempotencyRecord{
		Operation:            "issuer.credentials.issue",
		CallerPrincipalID:    "issuer_operator_alex",
		CallerOrganizationID: "did:web:issuer.hdip.dev",
		CallerActorType:      "issuer_operator",
		IdempotencyKey:       "issue-replay-1",
		RequestFingerprint:   "request-fingerprint-1",
		ResponseStatusCode:   201,
		ResourceType:         "credential",
		ResourceID:           "cred_hdip_passport_basic_001",
		Location:             "/v1/issuer/credentials/cred_hdip_passport_basic_001",
		ResponseBody:         json.RawMessage(`{"credentialId":"cred_hdip_passport_basic_001","status":"active"}`),
		CreatedAt:            time.Date(2026, time.April, 21, 9, 5, 0, 0, time.UTC),
	}
	if err := firstRuntime.CreateIdempotencyRecord(context.Background(), record); err != nil {
		t.Fatalf("create idempotency record: %v", err)
	}

	loaded, err := secondRuntime.GetIdempotencyRecord(
		context.Background(),
		record.Operation,
		record.CallerOrganizationID,
		record.CallerPrincipalID,
		record.CallerActorType,
		record.IdempotencyKey,
	)
	if err != nil {
		t.Fatalf("load idempotency record: %v", err)
	}

	var expectedBody any
	if err := json.Unmarshal(record.ResponseBody, &expectedBody); err != nil {
		t.Fatalf("unmarshal expected response body: %v", err)
	}
	var actualBody any
	if err := json.Unmarshal(loaded.ResponseBody, &actualBody); err != nil {
		t.Fatalf("unmarshal actual response body: %v", err)
	}
	if !reflect.DeepEqual(expectedBody, actualBody) {
		t.Fatalf("expected persisted response body %#v, got %#v", expectedBody, actualBody)
	}

	records, err := secondRuntime.ListIdempotencyRecords(context.Background())
	if err != nil {
		t.Fatalf("list idempotency records: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 idempotency record, got %d", len(records))
	}
}

func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
