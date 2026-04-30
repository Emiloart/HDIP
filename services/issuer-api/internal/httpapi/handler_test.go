package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Emiloart/HDIP/packages/go/foundation/authctx"
	"github.com/Emiloart/HDIP/packages/go/foundation/httpx"
	"github.com/Emiloart/HDIP/packages/go/foundation/testutil"
	phase1sql "github.com/Emiloart/HDIP/services/internal/phase1sql"
	phase1sqltest "github.com/Emiloart/HDIP/services/internal/phase1sqltest"
	"github.com/Emiloart/HDIP/services/issuer-api/internal/config"
	phase1 "github.com/Emiloart/HDIP/services/issuer-api/internal/phase1"
)

func TestHealthHandler(t *testing.T) {
	handler := newTestIssuerHandler(t)

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var response httpx.HealthResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Service != "issuer-api" || response.Status != "ok" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestReadyHandlerReportsSQLPrimaryRuntimeMode(t *testing.T) {
	handler := newTestIssuerHandler(t)

	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	if runtimeMode := recorder.Header().Get("X-HDIP-Phase1-Runtime-Mode"); runtimeMode != phase1.RuntimeModeSQLPrimary {
		t.Fatalf("expected sql-primary runtime mode header, got %q", runtimeMode)
	}

	var response httpx.HealthResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Status != "ready" {
		t.Fatalf("unexpected readiness response: %+v", response)
	}
}

func TestIssuerProfileHandler(t *testing.T) {
	handler := newTestIssuerHandler(t)

	request := httptest.NewRequest(http.MethodGet, "/v1/issuer/profile", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	testutil.AssertJSONMatchesFixture(t, recorder.Body.Bytes(), "schemas/examples/issuer/issuer-profile.default.json")
}

func TestIssuerTemplateHandler(t *testing.T) {
	handler := newTestIssuerHandler(t)

	request := httptest.NewRequest(http.MethodGet, "/v1/issuer/templates/hdip-passport-basic", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	testutil.AssertJSONMatchesFixture(t, recorder.Body.Bytes(), "schemas/examples/credentials/credential-template-metadata.hdip-passport-basic.json")
}

func TestPhase1IssueCredentialHandler(t *testing.T) {
	store := newIssuerStoreWithDefaults(t)
	handler := newTestIssuerHandlerWithStore(t, store)
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/issuer/credentials",
		strings.NewReader(loadFixtureText(t, "schemas/examples/issuer/issuance-request.hdip-passport-basic.json")),
	)
	setIssuerPhase1Headers(request, []string{issuerIssueScope})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}

	testutil.AssertJSONMatchesFixture(t, recorder.Body.Bytes(), "schemas/examples/issuer/issuance-response.hdip-passport-basic.json")
}

func TestPhase1IssueCredentialRejectsMissingAuth(t *testing.T) {
	handler := newTestIssuerHandler(t)
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/issuer/credentials",
		strings.NewReader(loadFixtureText(t, "schemas/examples/issuer/issuance-request.hdip-passport-basic.json")),
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestPhase1IssueCredentialRejectsMissingScope(t *testing.T) {
	handler := newTestIssuerHandler(t)
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/issuer/credentials",
		strings.NewReader(loadFixtureText(t, "schemas/examples/issuer/issuance-request.hdip-passport-basic.json")),
	)
	setIssuerPhase1Headers(request, []string{issuerReadScope})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}

func TestPhase1IssueCredentialAcceptsHydraBearerAttribution(t *testing.T) {
	store := newIssuerStoreWithDefaults(t)
	extractor := newTestIssuerHydraExtractor(t, []string{issuerIssueScope})
	handler := newMuxWithPhase1Handler(
		slog.Default(),
		testIssuerConfig(t),
		newPhase1IssuerHandlerWithStoreAndExtractor(store, extractor, extractor),
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/issuer/credentials",
		strings.NewReader(loadFixtureText(t, "schemas/examples/issuer/issuance-request.hdip-passport-basic.json")),
	)
	request.Header.Set("Authorization", "Bearer issuer-token")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestPhase1IssueCredentialRejectsHydraBearerMissingScope(t *testing.T) {
	store := newIssuerStoreWithDefaults(t)
	extractor := newTestIssuerHydraExtractor(t, []string{issuerReadScope})
	handler := newMuxWithPhase1Handler(
		slog.Default(),
		testIssuerConfig(t),
		newPhase1IssuerHandlerWithStoreAndExtractor(store, extractor, extractor),
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/issuer/credentials",
		strings.NewReader(loadFixtureText(t, "schemas/examples/issuer/issuance-request.hdip-passport-basic.json")),
	)
	request.Header.Set("Authorization", "Bearer issuer-token")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}

func TestPhase1IssueCredentialRejectsInvalidPayload(t *testing.T) {
	handler := newTestIssuerHandler(t)
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/issuer/credentials",
		strings.NewReader(loadFixtureText(t, "schemas/examples/issuer/issuance-request.with-issuer-id.invalid.json")),
	)
	setIssuerPhase1Headers(request, []string{issuerIssueScope})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestPhase1GetCredentialHandler(t *testing.T) {
	store := newIssuerStoreWithDefaults(t)
	handler := newTestIssuerHandlerWithStore(t, store)

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/v1/issuer/credentials",
		strings.NewReader(loadFixtureText(t, "schemas/examples/issuer/issuance-request.hdip-passport-basic.json")),
	)
	setIssuerPhase1Headers(createRequest, []string{issuerIssueScope})
	createRecorder := httptest.NewRecorder()
	handler.ServeHTTP(createRecorder, createRequest)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d", createRecorder.Code)
	}

	var created issuanceResponsePayload
	if err := json.Unmarshal(createRecorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	readRequest := httptest.NewRequest(http.MethodGet, "/v1/issuer/credentials/"+created.CredentialID, nil)
	setIssuerPhase1Headers(readRequest, []string{issuerReadScope})
	readRecorder := httptest.NewRecorder()
	handler.ServeHTTP(readRecorder, readRequest)

	if readRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", readRecorder.Code)
	}

	testutil.AssertJSONMatchesFixture(t, readRecorder.Body.Bytes(), "schemas/examples/credentials/credential-record.hdip-passport-basic.json")
}

func TestPhase1GetCredentialReturnsNotFound(t *testing.T) {
	handler := newTestIssuerHandler(t)
	request := httptest.NewRequest(http.MethodGet, "/v1/issuer/credentials/cred_missing_001", nil)
	setIssuerPhase1Headers(request, []string{issuerReadScope})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestPhase1IssueCredentialPersistsDeterministicArtifactAndAudits(t *testing.T) {
	store := newIssuerStoreWithDefaults(t)
	handler := newTestIssuerHandlerWithStore(t, store)
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/issuer/credentials",
		strings.NewReader(loadFixtureText(t, "schemas/examples/issuer/issuance-request.hdip-passport-basic.json")),
	)
	request.Header.Set("Idempotency-Key", "issue-1")
	setIssuerPhase1Headers(request, []string{issuerIssueScope})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}

	var response issuanceResponsePayload
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode issuance response: %v", err)
	}

	record, err := store.GetCredentialRecord(context.Background(), response.CredentialID)
	if err != nil {
		t.Fatalf("load credential record: %v", err)
	}

	expectedArtifact, err := materializeCredentialArtifact(
		response.CredentialID,
		defaultPhase1IssuerID,
		defaultTemplateID,
		time.Date(2027, time.April, 20, 9, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("materialize expected artifact: %v", err)
	}

	if record.CredentialArtifact == nil {
		t.Fatal("expected credential artifact to be stored")
	}

	if record.CredentialArtifact.Value != expectedArtifact.Value {
		t.Fatalf("unexpected artifact value: %s", record.CredentialArtifact.Value)
	}

	if record.ArtifactDigest != artifactDigest(expectedArtifact.Value) {
		t.Fatalf("unexpected artifact digest: %s", record.ArtifactDigest)
	}

	audits, err := store.AuditRecords()
	if err != nil {
		t.Fatalf("load audit records: %v", err)
	}

	if len(audits) != 1 {
		t.Fatalf("expected 1 audit record, got %d", len(audits))
	}

	if audits[0].Action != issuerIssueScope || audits[0].Outcome != "succeeded" {
		t.Fatalf("unexpected audit record: %+v", audits[0])
	}

	if audits[0].IdempotencyKey != "issue-1" {
		t.Fatalf("expected idempotency key to be recorded, got %q", audits[0].IdempotencyKey)
	}
}

func TestPhase1IssueCredentialReplayReturnsOriginalResult(t *testing.T) {
	store := newIssuerStoreWithDefaults(t)
	handler := newTestIssuerHandlerWithStore(t, store)
	body := loadFixtureText(t, "schemas/examples/issuer/issuance-request.hdip-passport-basic.json")

	firstRequest := httptest.NewRequest(http.MethodPost, "/v1/issuer/credentials", strings.NewReader(body))
	firstRequest.Header.Set("Idempotency-Key", "issue-replay-1")
	setIssuerPhase1Headers(firstRequest, []string{issuerIssueScope})
	firstRecorder := httptest.NewRecorder()
	handler.ServeHTTP(firstRecorder, firstRequest)
	if firstRecorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", firstRecorder.Code)
	}

	secondRequest := httptest.NewRequest(http.MethodPost, "/v1/issuer/credentials", strings.NewReader(body))
	secondRequest.Header.Set("Idempotency-Key", "issue-replay-1")
	setIssuerPhase1Headers(secondRequest, []string{issuerIssueScope})
	secondRecorder := httptest.NewRecorder()
	handler.ServeHTTP(secondRecorder, secondRequest)
	if secondRecorder.Code != http.StatusCreated {
		t.Fatalf("expected replay 201, got %d", secondRecorder.Code)
	}

	assertJSONEqual(t, firstRecorder.Body.Bytes(), secondRecorder.Body.Bytes())

	records, err := store.IdempotencyRecords()
	if err != nil {
		t.Fatalf("load idempotency records: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 idempotency record, got %d", len(records))
	}

	if strings.Contains(string(records[0].ResponseBody), "fullLegalName") || strings.Contains(string(records[0].ResponseBody), "Alex Example") {
		t.Fatalf("expected bounded replay snapshot, got %s", string(records[0].ResponseBody))
	}

	audits, err := store.AuditRecords()
	if err != nil {
		t.Fatalf("load audit records: %v", err)
	}

	if len(audits) != 2 {
		t.Fatalf("expected 2 audit records, got %d", len(audits))
	}

	if audits[1].Outcome != "replayed" {
		t.Fatalf("expected replay audit outcome, got %+v", audits[1])
	}
}

func TestPhase1IssueCredentialReplayConflictFailsCleanly(t *testing.T) {
	store := newIssuerStoreWithDefaults(t)
	handler := newTestIssuerHandlerWithStore(t, store)
	originalBody := loadFixtureText(t, "schemas/examples/issuer/issuance-request.hdip-passport-basic.json")
	conflictingBody := strings.ReplaceAll(originalBody, "\"subject_ref_alex_example\"", "\"subject_ref_conflict_example\"")

	firstRequest := httptest.NewRequest(http.MethodPost, "/v1/issuer/credentials", strings.NewReader(originalBody))
	firstRequest.Header.Set("Idempotency-Key", "issue-conflict-1")
	setIssuerPhase1Headers(firstRequest, []string{issuerIssueScope})
	firstRecorder := httptest.NewRecorder()
	handler.ServeHTTP(firstRecorder, firstRequest)
	if firstRecorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", firstRecorder.Code)
	}

	secondRequest := httptest.NewRequest(http.MethodPost, "/v1/issuer/credentials", strings.NewReader(conflictingBody))
	secondRequest.Header.Set("Idempotency-Key", "issue-conflict-1")
	setIssuerPhase1Headers(secondRequest, []string{issuerIssueScope})
	secondRecorder := httptest.NewRecorder()
	handler.ServeHTTP(secondRecorder, secondRequest)
	if secondRecorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", secondRecorder.Code)
	}

	var response httpx.ErrorEnvelope
	if err := json.Unmarshal(secondRecorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Error.Code != "idempotency_conflict" {
		t.Fatalf("unexpected error response: %+v", response)
	}
}

func TestPhase1IssueCredentialInProgressReservationFailsCleanly(t *testing.T) {
	store := newIssuerStoreWithDefaults(t)
	handler := newTestIssuerHandlerWithStore(t, store)
	body := loadFixtureText(t, "schemas/examples/issuer/issuance-request.hdip-passport-basic.json")
	var requestPayload issuanceRequestPayload
	if err := json.Unmarshal([]byte(body), &requestPayload); err != nil {
		t.Fatalf("decode request fixture: %v", err)
	}

	if _, err := store.ReserveIdempotencyRecord(context.Background(), phase1.IdempotencyRecord{
		Operation:            issuerIssueScope,
		CallerPrincipalID:    "issuer_operator_alex",
		CallerOrganizationID: defaultPhase1IssuerID,
		CallerActorType:      "issuer_operator",
		IdempotencyKey:       "issue-in-progress-1",
		RequestFingerprint:   issuanceRequestFingerprint(requestPayload),
		ResourceType:         "credential",
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
	}); err != nil {
		t.Fatalf("reserve idempotency record: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/v1/issuer/credentials", strings.NewReader(body))
	request.Header.Set("Idempotency-Key", "issue-in-progress-1")
	setIssuerPhase1Headers(request, []string{issuerIssueScope})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", recorder.Code)
	}

	var response httpx.ErrorEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Error.Code != "idempotency_in_progress" {
		t.Fatalf("unexpected error response: %+v", response)
	}
}

func TestPhase1UpdateCredentialStatusHandler(t *testing.T) {
	store := newIssuerStoreWithDefaults(t)
	handler := newTestIssuerHandlerWithStore(t, store)

	credentialID := issueCredentialForIssuerTest(t, handler)

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/issuer/credentials/"+credentialID+"/status",
		strings.NewReader(loadFixtureText(t, "schemas/examples/issuer/credential-status-update-request.revoked.json")),
	)
	request.Header.Set("Idempotency-Key", "status-1")
	setIssuerPhase1Headers(request, []string{issuerStatusWriteScope})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var response credentialStatusPayload
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Status != "revoked" {
		t.Fatalf("expected revoked status, got %+v", response)
	}

	readRequest := httptest.NewRequest(http.MethodGet, "/v1/issuer/credentials/"+credentialID, nil)
	setIssuerPhase1Headers(readRequest, []string{issuerReadScope})
	readRecorder := httptest.NewRecorder()
	handler.ServeHTTP(readRecorder, readRequest)
	if readRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 after status update, got %d", readRecorder.Code)
	}

	var record credentialRecordPayload
	if err := json.Unmarshal(readRecorder.Body.Bytes(), &record); err != nil {
		t.Fatalf("decode credential record: %v", err)
	}

	if record.Status != "revoked" {
		t.Fatalf("expected revoked credential record status, got %+v", record)
	}

	audits, err := store.AuditRecords()
	if err != nil {
		t.Fatalf("load audit records: %v", err)
	}

	if len(audits) != 3 {
		t.Fatalf("expected 3 audit records, got %d", len(audits))
	}

	if audits[1].Action != issuerStatusWriteScope || audits[1].Outcome != "succeeded" {
		t.Fatalf("unexpected status audit record: %+v", audits[1])
	}

	if audits[1].IdempotencyKey != "status-1" {
		t.Fatalf("expected status idempotency key to be recorded, got %q", audits[1].IdempotencyKey)
	}
}

func TestPhase1UpdateCredentialStatusRejectsInvalidTransition(t *testing.T) {
	store := newIssuerStoreWithDefaults(t)
	handler := newTestIssuerHandlerWithStore(t, store)

	credentialID := issueCredentialForIssuerTest(t, handler)
	updateCredentialStatusForIssuerTest(
		t,
		handler,
		credentialID,
		loadFixtureText(t, "schemas/examples/issuer/credential-status-update-request.revoked.json"),
		http.StatusOK,
	)

	updateCredentialStatusForIssuerTest(
		t,
		handler,
		credentialID,
		loadFixtureText(t, "schemas/examples/issuer/credential-status-update-request.revoked.json"),
		http.StatusConflict,
	)
}

func TestPhase1UpdateCredentialStatusReplayReturnsOriginalResult(t *testing.T) {
	store := newIssuerStoreWithDefaults(t)
	handler := newTestIssuerHandlerWithStore(t, store)
	credentialID := issueCredentialForIssuerTest(t, handler)
	body := loadFixtureText(t, "schemas/examples/issuer/credential-status-update-request.revoked.json")

	firstRequest := httptest.NewRequest(http.MethodPost, "/v1/issuer/credentials/"+credentialID+"/status", strings.NewReader(body))
	firstRequest.Header.Set("Idempotency-Key", "status-replay-1")
	setIssuerPhase1Headers(firstRequest, []string{issuerStatusWriteScope})
	firstRecorder := httptest.NewRecorder()
	handler.ServeHTTP(firstRecorder, firstRequest)
	if firstRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", firstRecorder.Code)
	}

	secondRequest := httptest.NewRequest(http.MethodPost, "/v1/issuer/credentials/"+credentialID+"/status", strings.NewReader(body))
	secondRequest.Header.Set("Idempotency-Key", "status-replay-1")
	setIssuerPhase1Headers(secondRequest, []string{issuerStatusWriteScope})
	secondRecorder := httptest.NewRecorder()
	handler.ServeHTTP(secondRecorder, secondRequest)
	if secondRecorder.Code != http.StatusOK {
		t.Fatalf("expected replay 200, got %d", secondRecorder.Code)
	}

	assertJSONEqual(t, firstRecorder.Body.Bytes(), secondRecorder.Body.Bytes())
}

func TestPhase1UpdateCredentialStatusReplayConflictFailsCleanly(t *testing.T) {
	store := newIssuerStoreWithDefaults(t)
	handler := newTestIssuerHandlerWithStore(t, store)
	credentialID := issueCredentialForIssuerTest(t, handler)

	firstRequest := httptest.NewRequest(
		http.MethodPost,
		"/v1/issuer/credentials/"+credentialID+"/status",
		strings.NewReader(loadFixtureText(t, "schemas/examples/issuer/credential-status-update-request.revoked.json")),
	)
	firstRequest.Header.Set("Idempotency-Key", "status-conflict-1")
	setIssuerPhase1Headers(firstRequest, []string{issuerStatusWriteScope})
	firstRecorder := httptest.NewRecorder()
	handler.ServeHTTP(firstRecorder, firstRequest)
	if firstRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", firstRecorder.Code)
	}

	secondRequest := httptest.NewRequest(
		http.MethodPost,
		"/v1/issuer/credentials/"+credentialID+"/status",
		strings.NewReader(`{"status":"superseded","supersededByCredentialId":"cred_hdip_passport_basic_002"}`),
	)
	secondRequest.Header.Set("Idempotency-Key", "status-conflict-1")
	setIssuerPhase1Headers(secondRequest, []string{issuerStatusWriteScope})
	secondRecorder := httptest.NewRecorder()
	handler.ServeHTTP(secondRecorder, secondRequest)
	if secondRecorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", secondRecorder.Code)
	}

	var response httpx.ErrorEnvelope
	if err := json.Unmarshal(secondRecorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Error.Code != "idempotency_conflict" {
		t.Fatalf("unexpected error response: %+v", response)
	}
}

func TestPhase1UpdateCredentialStatusInProgressReservationFailsCleanly(t *testing.T) {
	store := newIssuerStoreWithDefaults(t)
	handler := newTestIssuerHandlerWithStore(t, store)
	credentialID := issueCredentialForIssuerTest(t, handler)
	body := loadFixtureText(t, "schemas/examples/issuer/credential-status-update-request.revoked.json")
	var requestPayload credentialStatusUpdateRequestPayload
	if err := json.Unmarshal([]byte(body), &requestPayload); err != nil {
		t.Fatalf("decode request fixture: %v", err)
	}

	if _, err := store.ReserveIdempotencyRecord(context.Background(), phase1.IdempotencyRecord{
		Operation:            issuerStatusWriteScope,
		CallerPrincipalID:    "issuer_operator_alex",
		CallerOrganizationID: defaultPhase1IssuerID,
		CallerActorType:      "issuer_operator",
		IdempotencyKey:       "status-in-progress-1",
		RequestFingerprint:   credentialStatusUpdateFingerprint(credentialID, requestPayload),
		ResourceType:         "credential",
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
	}); err != nil {
		t.Fatalf("reserve idempotency record: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/v1/issuer/credentials/"+credentialID+"/status", strings.NewReader(body))
	request.Header.Set("Idempotency-Key", "status-in-progress-1")
	setIssuerPhase1Headers(request, []string{issuerStatusWriteScope})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", recorder.Code)
	}

	var response httpx.ErrorEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Error.Code != "idempotency_in_progress" {
		t.Fatalf("unexpected error response: %+v", response)
	}
}

func TestPhase1UpdateCredentialStatusRejectsMissingAuth(t *testing.T) {
	store := newIssuerStoreWithDefaults(t)
	handler := newTestIssuerHandlerWithStore(t, store)

	credentialID := issueCredentialForIssuerTest(t, handler)
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/issuer/credentials/"+credentialID+"/status",
		strings.NewReader(loadFixtureText(t, "schemas/examples/issuer/credential-status-update-request.revoked.json")),
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestPhase1UpdateCredentialStatusRejectsMissingScope(t *testing.T) {
	store := newIssuerStoreWithDefaults(t)
	handler := newTestIssuerHandlerWithStore(t, store)

	credentialID := issueCredentialForIssuerTest(t, handler)
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/issuer/credentials/"+credentialID+"/status",
		strings.NewReader(loadFixtureText(t, "schemas/examples/issuer/credential-status-update-request.revoked.json")),
	)
	setIssuerPhase1Headers(request, []string{issuerReadScope})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}

func TestPhase1UpdateCredentialStatusRejectsInvalidPayload(t *testing.T) {
	store := newIssuerStoreWithDefaults(t)
	handler := newTestIssuerHandlerWithStore(t, store)

	credentialID := issueCredentialForIssuerTest(t, handler)
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/issuer/credentials/"+credentialID+"/status",
		strings.NewReader(loadFixtureText(t, "schemas/examples/issuer/credential-status-update-request.superseded-without-reference.invalid.json")),
	)
	setIssuerPhase1Headers(request, []string{issuerStatusWriteScope})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestPhase1PrimarySQLStoreRoundTrip(t *testing.T) {
	dsn := os.Getenv("HDIP_PHASE1_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HDIP_PHASE1_TEST_DATABASE_URL is not set")
	}

	if err := phase1sql.MigrateUp(context.Background(), "pgx", dsn); err != nil {
		t.Fatalf("migrate sql store: %v", err)
	}

	sqlStore, err := phase1sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open sql store: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlStore.Close()
	})

	store, err := phase1.OpenStore(phase1.StoreOptions{
		DatabaseDriver: "pgx",
		DatabaseURL:    dsn,
	})
	if err != nil {
		t.Fatalf("open runtime store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	if err := store.SeedIssuerRecord(defaultIssuerRecord()); err != nil {
		t.Fatalf("seed issuer record: %v", err)
	}

	handler := newTestIssuerHandlerWithStore(t, store)
	credentialID := issueCredentialForIssuerTest(t, handler)
	updateCredentialStatusForIssuerTest(
		t,
		handler,
		credentialID,
		loadFixtureText(t, "schemas/examples/issuer/credential-status-update-request.revoked.json"),
		http.StatusOK,
	)

	record, err := sqlStore.GetCredentialRecord(context.Background(), credentialID)
	if err != nil {
		t.Fatalf("load credential from sql store: %v", err)
	}

	if record.Status != "revoked" || record.IssuerID != defaultPhase1IssuerID {
		t.Fatalf("unexpected sql-backed credential record: %+v", record)
	}
}

func newTestIssuerHandler(t *testing.T) http.Handler {
	t.Helper()

	return newTestIssuerHandlerWithStore(t, newIssuerStoreWithDefaults(t))
}

func newTestIssuerHandlerWithStore(t *testing.T, store *phase1.RuntimeStore) http.Handler {
	t.Helper()

	return newMuxWithPhase1Handler(slog.Default(), testIssuerConfig(t), newPhase1IssuerHandlerWithStore(store))
}

func newIssuerStoreWithDefaults(t *testing.T) *phase1.RuntimeStore {
	t.Helper()

	sqlStore := phase1sqltest.OpenSQLiteStore(t)
	store := phase1.NewSQLRuntimeStore(sqlStore)

	if err := store.SeedIssuerRecord(defaultIssuerRecord()); err != nil {
		t.Fatalf("seed issuer record: %v", err)
	}

	return store
}

func testIssuerConfig(t *testing.T) config.Config {
	t.Helper()

	return config.Config{
		ServiceName:       "issuer-api",
		Host:              "127.0.0.1",
		Port:              8081,
		LogLevel:          "INFO",
		RequestTimeout:    time.Second,
		ReadHeaderTimeout: time.Second,
		ShutdownTimeout:   time.Second,
		BuildVersion:      "test",
	}
}

func setIssuerPhase1Headers(request *http.Request, scopes []string) {
	request.Header.Set("X-HDIP-Principal-ID", "issuer_operator_alex")
	request.Header.Set("X-HDIP-Organization-ID", defaultPhase1IssuerID)
	request.Header.Set("X-HDIP-Auth-Reference", "session_issuer_001")
	request.Header.Set("X-HDIP-Scopes", strings.Join(scopes, ","))
}

func newTestIssuerHydraExtractor(t *testing.T, scopes []string) *authctx.HydraIssuerOperatorExtractor {
	t.Helper()

	scopeSet := strings.Join(scopes, " ")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected hydra introspection method: %s", r.Method)
		}
		if username, password, ok := r.BasicAuth(); !ok || username != "issuer-api" || password != "issuer-introspection-secret" {
			t.Fatalf("unexpected hydra introspection client credentials")
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse hydra introspection form: %v", err)
		}
		if r.Form.Get("token") != "issuer-token" && r.Form.Get("token") != "phase1-public-auth-readiness-probe" {
			t.Fatalf("unexpected hydra introspection token: %q", r.Form.Get("token"))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"active":true,"client_id":"` + defaultPhase1IssuerID + `","scope":"` + scopeSet + `","jti":"issuer-token-id"}`))
	}))
	t.Cleanup(server.Close)

	extractor, err := authctx.NewHydraIssuerOperatorExtractor(authctx.HydraIntrospectionConfig{
		IntrospectionURL: server.URL,
		ClientID:         "issuer-api",
		ClientSecret:     "issuer-introspection-secret",
		HTTPClient:       server.Client(),
	})
	if err != nil {
		t.Fatalf("new hydra issuer extractor: %v", err)
	}

	return extractor
}

func loadFixtureText(t *testing.T, relativePath string) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine test path")
	}

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "..", ".."))
	raw, err := os.ReadFile(filepath.Join(repoRoot, relativePath))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	return string(raw)
}

func issueCredentialForIssuerTest(t *testing.T, handler http.Handler) string {
	t.Helper()

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/issuer/credentials",
		strings.NewReader(loadFixtureText(t, "schemas/examples/issuer/issuance-request.hdip-passport-basic.json")),
	)
	setIssuerPhase1Headers(request, []string{issuerIssueScope})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}

	var response issuanceResponsePayload
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode issuance response: %v", err)
	}

	return response.CredentialID
}

func updateCredentialStatusForIssuerTest(
	t *testing.T,
	handler http.Handler,
	credentialID string,
	body string,
	expectedStatus int,
) {
	t.Helper()

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/issuer/credentials/"+credentialID+"/status",
		strings.NewReader(body),
	)
	setIssuerPhase1Headers(request, []string{issuerStatusWriteScope})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != expectedStatus {
		t.Fatalf("expected %d, got %d", expectedStatus, recorder.Code)
	}
}

func assertJSONEqual(t *testing.T, expected []byte, actual []byte) {
	t.Helper()

	var expectedValue any
	if err := json.Unmarshal(expected, &expectedValue); err != nil {
		t.Fatalf("unmarshal expected json: %v", err)
	}

	var actualValue any
	if err := json.Unmarshal(actual, &actualValue); err != nil {
		t.Fatalf("unmarshal actual json: %v", err)
	}

	if expectedJSON, actualJSON := expectedValue, actualValue; !reflect.DeepEqual(expectedJSON, actualJSON) {
		t.Fatalf("unexpected json body\nexpected: %#v\nactual: %#v", expectedJSON, actualJSON)
	}
}
