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
	"github.com/Emiloart/HDIP/services/verifier-api/internal/config"
	phase1 "github.com/Emiloart/HDIP/services/verifier-api/internal/phase1"
)

func TestHealthHandler(t *testing.T) {
	handler := newTestVerifierHandler(t)

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

	if response.Service != "verifier-api" || response.Status != "ok" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestVerifierPolicyRequestHandler(t *testing.T) {
	handler := newTestVerifierHandler(t)

	request := httptest.NewRequest(http.MethodGet, "/v1/verifier/policy-requests/kyc-passport-basic", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	testutil.AssertJSONMatchesFixture(t, recorder.Body.Bytes(), "schemas/examples/verifier/verifier-policy-request.kyc-passport-basic.json")
}

func TestVerifierStubResultHandler(t *testing.T) {
	handler := newTestVerifierHandler(t)

	request := httptest.NewRequest(http.MethodGet, "/v1/verifier/results/kyc-passport-basic-review/stub", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	testutil.AssertJSONMatchesFixture(t, recorder.Body.Bytes(), "schemas/examples/verifier/verifier-result.kyc-passport-basic-review.json")
}

func TestPhase1CreateVerificationHandler(t *testing.T) {
	store := newVerifierStoreWithDefaults(t)
	handler := newTestVerifierHandlerWithStore(t, store)
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/verifier/verifications",
		strings.NewReader(loadVerifierFixtureText(t, "schemas/examples/verifier/verification-submission-request.hdip-passport-basic.json")),
	)
	setVerifierPhase1Headers(request, []string{verifierCreateScope})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}

	testutil.AssertJSONMatchesFixture(t, recorder.Body.Bytes(), "schemas/examples/verifier/verification-result.allow.json")
}

func TestPhase1CreateVerificationRejectsMissingAuth(t *testing.T) {
	handler := newTestVerifierHandler(t)
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/verifier/verifications",
		strings.NewReader(loadVerifierFixtureText(t, "schemas/examples/verifier/verification-submission-request.hdip-passport-basic.json")),
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestPhase1CreateVerificationRejectsMissingScope(t *testing.T) {
	handler := newTestVerifierHandler(t)
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/verifier/verifications",
		strings.NewReader(loadVerifierFixtureText(t, "schemas/examples/verifier/verification-submission-request.hdip-passport-basic.json")),
	)
	setVerifierPhase1Headers(request, []string{verifierReadScope})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}

func TestPhase1CreateVerificationRejectsInvalidPayload(t *testing.T) {
	handler := newTestVerifierHandler(t)
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/verifier/verifications",
		strings.NewReader(loadVerifierFixtureText(t, "schemas/examples/verifier/verification-submission-request.with-verifier-id.invalid.json")),
	)
	setVerifierPhase1Headers(request, []string{verifierCreateScope})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestPhase1CreateVerificationDeniesSuspendedIssuer(t *testing.T) {
	store := newVerifierStoreWithDefaults(t)
	suspended := defaultVerifierIssuerRecord()
	suspended.TrustState = "suspended"
	if err := store.SeedIssuerRecord(suspended); err != nil {
		t.Fatalf("seed suspended issuer: %v", err)
	}
	handler := newTestVerifierHandlerWithStore(t, store)

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/verifier/verifications",
		strings.NewReader(loadVerifierFixtureText(t, "schemas/examples/verifier/verification-submission-request.hdip-passport-basic.json")),
	)
	setVerifierPhase1Headers(request, []string{verifierCreateScope})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}

	var response verificationResultPayload
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Decision != "deny" || len(response.ReasonCodes) != 1 || response.ReasonCodes[0] != "issuer_suspended" {
		t.Fatalf("unexpected response: %+v", response)
	}

	if response.CredentialStatus != "active" {
		t.Fatalf("expected active credential status, got %q", response.CredentialStatus)
	}
}

func TestPhase1CreateVerificationDeniesMissingOrNonActiveIssuer(t *testing.T) {
	testCases := []struct {
		name   string
		mutate func(*phase1.RuntimeStore)
	}{
		{
			name: "missing issuer",
			mutate: func(store *phase1.RuntimeStore) {
				if err := store.DeleteIssuerRecord(defaultPhase1IssuerID); err != nil {
					t.Fatalf("delete issuer record: %v", err)
				}
			},
		},
		{
			name: "pending issuer",
			mutate: func(store *phase1.RuntimeStore) {
				pending := defaultVerifierIssuerRecord()
				pending.TrustState = "pending"
				if err := store.SeedIssuerRecord(pending); err != nil {
					t.Fatalf("seed pending issuer: %v", err)
				}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := newVerifierStoreWithDefaults(t)
			testCase.mutate(store)
			handler := newTestVerifierHandlerWithStore(t, store)
			request := httptest.NewRequest(
				http.MethodPost,
				"/v1/verifier/verifications",
				strings.NewReader(loadVerifierFixtureText(t, "schemas/examples/verifier/verification-submission-request.hdip-passport-basic.json")),
			)
			setVerifierPhase1Headers(request, []string{verifierCreateScope})
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusCreated {
				t.Fatalf("expected 201, got %d", recorder.Code)
			}

			var response verificationResultPayload
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if response.Decision != "deny" || len(response.ReasonCodes) != 1 || response.ReasonCodes[0] != "issuer_not_trusted" {
				t.Fatalf("unexpected response: %+v", response)
			}
		})
	}
}

func TestPhase1CreateVerificationDeniesRevokedCredential(t *testing.T) {
	store := newVerifierStoreWithDefaults(t)
	revoked := defaultVerifierCredentialRecord()
	revoked.Status = phase1.CredentialStatusSnapshotRevoked
	if err := store.SeedCredentialRecord(revoked); err != nil {
		t.Fatalf("seed revoked credential: %v", err)
	}
	handler := newTestVerifierHandlerWithStore(t, store)
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/verifier/verifications",
		strings.NewReader(loadVerifierFixtureText(t, "schemas/examples/verifier/verification-submission-request.hdip-passport-basic.json")),
	)
	setVerifierPhase1Headers(request, []string{verifierCreateScope})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}

	var response verificationResultPayload
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Decision != "deny" || response.CredentialStatus != "revoked" {
		t.Fatalf("unexpected response: %+v", response)
	}

	if len(response.ReasonCodes) != 1 || response.ReasonCodes[0] != "credential_status_revoked" {
		t.Fatalf("unexpected reason codes: %+v", response.ReasonCodes)
	}
}

func TestPhase1CreateVerificationDeniesArtifactContinuityFailure(t *testing.T) {
	store := newVerifierStoreWithDefaults(t)
	handler := newTestVerifierHandlerWithStore(t, store)
	mutatedArtifactValue := materializeCredentialArtifactValue(
		"cred_hdip_passport_basic_001",
		defaultPhase1IssuerID,
		defaultTemplateID,
		time.Date(2027, time.April, 21, 9, 0, 0, 0, time.UTC),
	)
	body := `{"policyId":"kyc-passport-basic","credentialId":"cred_hdip_passport_basic_001","credentialArtifact":{"kind":"phase1_opaque_artifact","mediaType":"application/vnd.hdip.phase1-opaque-artifact","value":"` + mutatedArtifactValue + `"}}`
	request := httptest.NewRequest(http.MethodPost, "/v1/verifier/verifications", strings.NewReader(body))
	setVerifierPhase1Headers(request, []string{verifierCreateScope})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}

	var response verificationResultPayload
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Decision != "deny" || len(response.ReasonCodes) != 1 || response.ReasonCodes[0] != "artifact_continuity_failed" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestPhase1CreateVerificationReplayReturnsOriginalResult(t *testing.T) {
	store := newVerifierStoreWithDefaults(t)
	handler := newTestVerifierHandlerWithStore(t, store)
	body := loadVerifierFixtureText(t, "schemas/examples/verifier/verification-submission-request.hdip-passport-basic.json")

	firstRequest := httptest.NewRequest(http.MethodPost, "/v1/verifier/verifications", strings.NewReader(body))
	firstRequest.Header.Set("Idempotency-Key", "verify-replay-1")
	setVerifierPhase1Headers(firstRequest, []string{verifierCreateScope})
	firstRecorder := httptest.NewRecorder()
	handler.ServeHTTP(firstRecorder, firstRequest)
	if firstRecorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", firstRecorder.Code)
	}

	secondRequest := httptest.NewRequest(http.MethodPost, "/v1/verifier/verifications", strings.NewReader(body))
	secondRequest.Header.Set("Idempotency-Key", "verify-replay-1")
	setVerifierPhase1Headers(secondRequest, []string{verifierCreateScope})
	secondRecorder := httptest.NewRecorder()
	handler.ServeHTTP(secondRecorder, secondRequest)
	if secondRecorder.Code != http.StatusCreated {
		t.Fatalf("expected replay 201, got %d", secondRecorder.Code)
	}

	assertVerifierJSONEqual(t, firstRecorder.Body.Bytes(), secondRecorder.Body.Bytes())

	records, err := store.IdempotencyRecords()
	if err != nil {
		t.Fatalf("load idempotency records: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 idempotency record, got %d", len(records))
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

func TestPhase1CreateVerificationReplayConflictFailsCleanly(t *testing.T) {
	store := newVerifierStoreWithDefaults(t)
	handler := newTestVerifierHandlerWithStore(t, store)
	firstBody := loadVerifierFixtureText(t, "schemas/examples/verifier/verification-submission-request.hdip-passport-basic.json")
	conflictingBody := strings.ReplaceAll(firstBody, "\"kyc-passport-basic\"", "\"kyc-passport-plus\"")

	firstRequest := httptest.NewRequest(http.MethodPost, "/v1/verifier/verifications", strings.NewReader(firstBody))
	firstRequest.Header.Set("Idempotency-Key", "verify-conflict-1")
	setVerifierPhase1Headers(firstRequest, []string{verifierCreateScope})
	firstRecorder := httptest.NewRecorder()
	handler.ServeHTTP(firstRecorder, firstRequest)
	if firstRecorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", firstRecorder.Code)
	}

	secondRequest := httptest.NewRequest(http.MethodPost, "/v1/verifier/verifications", strings.NewReader(conflictingBody))
	secondRequest.Header.Set("Idempotency-Key", "verify-conflict-1")
	setVerifierPhase1Headers(secondRequest, []string{verifierCreateScope})
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

func TestPhase1GetVerificationHandler(t *testing.T) {
	store := newVerifierStoreWithDefaults(t)
	handler := newTestVerifierHandlerWithStore(t, store)

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/v1/verifier/verifications",
		strings.NewReader(loadVerifierFixtureText(t, "schemas/examples/verifier/verification-submission-request.hdip-passport-basic.json")),
	)
	setVerifierPhase1Headers(createRequest, []string{verifierCreateScope})
	createRecorder := httptest.NewRecorder()
	handler.ServeHTTP(createRecorder, createRequest)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d", createRecorder.Code)
	}

	var created verificationResultPayload
	if err := json.Unmarshal(createRecorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	readRequest := httptest.NewRequest(http.MethodGet, "/v1/verifier/verifications/"+created.VerificationID, nil)
	setVerifierPhase1Headers(readRequest, []string{verifierReadScope})
	readRecorder := httptest.NewRecorder()
	handler.ServeHTTP(readRecorder, readRequest)

	if readRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", readRecorder.Code)
	}

	testutil.AssertJSONMatchesFixture(t, readRecorder.Body.Bytes(), "schemas/examples/verifier/verification-result.allow.json")
}

func TestPhase1GetVerificationReturnsNotFound(t *testing.T) {
	handler := newTestVerifierHandler(t)
	request := httptest.NewRequest(http.MethodGet, "/v1/verifier/verifications/verification_missing_001", nil)
	setVerifierPhase1Headers(request, []string{verifierReadScope})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestPhase1VerificationAuditsReadsAndWrites(t *testing.T) {
	store := newVerifierStoreWithDefaults(t)
	handler := newTestVerifierHandlerWithStore(t, store)

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/v1/verifier/verifications",
		strings.NewReader(loadVerifierFixtureText(t, "schemas/examples/verifier/verification-submission-request.hdip-passport-basic.json")),
	)
	createRequest.Header.Set("Idempotency-Key", "verify-1")
	setVerifierPhase1Headers(createRequest, []string{verifierCreateScope})
	createRecorder := httptest.NewRecorder()
	handler.ServeHTTP(createRecorder, createRequest)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d", createRecorder.Code)
	}

	var created verificationResultPayload
	if err := json.Unmarshal(createRecorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	readRequest := httptest.NewRequest(http.MethodGet, "/v1/verifier/verifications/"+created.VerificationID, nil)
	setVerifierPhase1Headers(readRequest, []string{verifierReadScope})
	readRecorder := httptest.NewRecorder()
	handler.ServeHTTP(readRecorder, readRequest)
	if readRecorder.Code != http.StatusOK {
		t.Fatalf("expected read status 200, got %d", readRecorder.Code)
	}

	audits, err := store.AuditRecords()
	if err != nil {
		t.Fatalf("load audit records: %v", err)
	}

	if len(audits) != 2 {
		t.Fatalf("expected 2 audit records, got %d", len(audits))
	}

	if audits[0].Action != verifierCreateScope || audits[0].Outcome != "succeeded" {
		t.Fatalf("unexpected create audit: %+v", audits[0])
	}

	if audits[0].IdempotencyKey != "verify-1" {
		t.Fatalf("expected idempotency key to be recorded, got %q", audits[0].IdempotencyKey)
	}

	if audits[1].Action != verifierReadScope || audits[1].Outcome != "succeeded" {
		t.Fatalf("unexpected read audit: %+v", audits[1])
	}
}

func TestPhase1CreateVerificationReflectsPersistedStatusTransition(t *testing.T) {
	store := newVerifierStoreWithDefaults(t)
	handler := newTestVerifierHandlerWithStore(t, store)

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/verifier/verifications",
		strings.NewReader(loadVerifierFixtureText(t, "schemas/examples/verifier/verification-submission-request.hdip-passport-basic.json")),
	)
	setVerifierPhase1Headers(request, []string{verifierCreateScope})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected initial 201, got %d", recorder.Code)
	}

	if err := store.UpdateCredentialStatus(
		context.Background(),
		"cred_hdip_passport_basic_001",
		phase1.CredentialStatusSnapshotRevoked,
		time.Date(2026, time.April, 20, 10, 0, 0, 0, time.UTC),
		"",
	); err != nil {
		t.Fatalf("update credential status: %v", err)
	}

	secondRequest := httptest.NewRequest(
		http.MethodPost,
		"/v1/verifier/verifications",
		strings.NewReader(loadVerifierFixtureText(t, "schemas/examples/verifier/verification-submission-request.hdip-passport-basic.json")),
	)
	setVerifierPhase1Headers(secondRequest, []string{verifierCreateScope})
	secondRecorder := httptest.NewRecorder()
	handler.ServeHTTP(secondRecorder, secondRequest)
	if secondRecorder.Code != http.StatusCreated {
		t.Fatalf("expected 201 after persisted status transition, got %d", secondRecorder.Code)
	}

	var response verificationResultPayload
	if err := json.Unmarshal(secondRecorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Decision != "deny" || response.CredentialStatus != "revoked" {
		t.Fatalf("expected deny/revoked after status transition, got %+v", response)
	}

	if len(response.ReasonCodes) != 1 || response.ReasonCodes[0] != "credential_status_revoked" {
		t.Fatalf("unexpected reason codes after status transition: %+v", response.ReasonCodes)
	}
}

func TestPhase1CreateVerificationUsesExplicitTrustAdapter(t *testing.T) {
	store := newVerifierStoreWithDefaults(t)
	trusts := &spyTrustReadRepository{
		record: phase1.IssuerTrustRecord{
			IssuerID:                  defaultPhase1IssuerID,
			TrustState:                "suspended",
			AllowedTemplateIDs:        []string{defaultTemplateID},
			VerificationKeyReferences: []string{"key:issuer.hdip.dev:2026-04"},
		},
	}
	handler := newMuxWithPhase1Handler(
		slog.Default(),
		testVerifierConfig(t),
		&phase1VerifierHandler{
			verifierExtractor: authctx.HeaderVerifierIntegratorExtractor{},
			trusts:            trusts,
			credentials:       store,
			requests:          store,
			results:           store,
			idempotency:       store,
			audits:            store,
		},
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/verifier/verifications",
		strings.NewReader(loadVerifierFixtureText(t, "schemas/examples/verifier/verification-submission-request.hdip-passport-basic.json")),
	)
	setVerifierPhase1Headers(request, []string{verifierCreateScope})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}

	if !trusts.called {
		t.Fatal("expected explicit trust adapter to be called")
	}

	var response verificationResultPayload
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Decision != "deny" || len(response.ReasonCodes) != 1 || response.ReasonCodes[0] != "issuer_suspended" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func newTestVerifierHandler(t *testing.T) http.Handler {
	t.Helper()

	handler, err := NewMux(slog.Default(), testVerifierConfig(t))
	if err != nil {
		t.Fatalf("new verifier mux: %v", err)
	}

	return handler
}

func newTestVerifierHandlerWithStore(t *testing.T, store *phase1.RuntimeStore) http.Handler {
	t.Helper()

	return newMuxWithPhase1Handler(slog.Default(), testVerifierConfig(t), newPhase1VerifierHandlerWithStore(store))
}

func newVerifierStoreWithDefaults(t *testing.T) *phase1.RuntimeStore {
	t.Helper()

	store, err := phase1.OpenRuntimeStore(filepath.Join(t.TempDir(), "verifier-phase1-state.json"))
	if err != nil {
		t.Fatalf("open runtime store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	if err := store.SeedIssuerRecord(defaultVerifierIssuerRecord()); err != nil {
		t.Fatalf("seed issuer record: %v", err)
	}
	if err := store.SeedCredentialRecord(defaultVerifierCredentialRecord()); err != nil {
		t.Fatalf("seed credential record: %v", err)
	}

	return store
}

func testVerifierConfig(t *testing.T) config.Config {
	t.Helper()

	return config.Config{
		ServiceName:       "verifier-api",
		Host:              "127.0.0.1",
		Port:              8082,
		LogLevel:          "INFO",
		RequestTimeout:    time.Second,
		ReadHeaderTimeout: time.Second,
		ShutdownTimeout:   time.Second,
		Phase1StatePath:   filepath.Join(t.TempDir(), "verifier-phase1-state.json"),
		BuildVersion:      "test",
	}
}

func setVerifierPhase1Headers(request *http.Request, scopes []string) {
	request.Header.Set("X-HDIP-Principal-ID", "verifier_integrator_alpha")
	request.Header.Set("X-HDIP-Organization-ID", "verifier_org_marketplace_alpha")
	request.Header.Set("X-HDIP-Auth-Reference", "credential_verifier_001")
	request.Header.Set("X-HDIP-Scopes", strings.Join(scopes, ","))
}

func loadVerifierFixtureText(t *testing.T, relativePath string) string {
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

func assertVerifierJSONEqual(t *testing.T, expected []byte, actual []byte) {
	t.Helper()

	var expectedValue any
	if err := json.Unmarshal(expected, &expectedValue); err != nil {
		t.Fatalf("unmarshal expected json: %v", err)
	}

	var actualValue any
	if err := json.Unmarshal(actual, &actualValue); err != nil {
		t.Fatalf("unmarshal actual json: %v", err)
	}

	if !reflect.DeepEqual(expectedValue, actualValue) {
		t.Fatalf("unexpected json body\nexpected: %#v\nactual: %#v", expectedValue, actualValue)
	}
}

type spyTrustReadRepository struct {
	record   phase1.IssuerTrustRecord
	err      error
	called   bool
	issuerID string
}

func (s *spyTrustReadRepository) GetIssuerTrustRecord(_ context.Context, issuerID string) (phase1.IssuerTrustRecord, error) {
	s.called = true
	s.issuerID = issuerID
	if s.err != nil {
		return phase1.IssuerTrustRecord{}, s.err
	}

	return s.record, nil
}
