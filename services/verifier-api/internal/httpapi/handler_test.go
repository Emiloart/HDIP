package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Emiloart/HDIP/packages/go/foundation/httpx"
	"github.com/Emiloart/HDIP/packages/go/foundation/testutil"
	"github.com/Emiloart/HDIP/services/verifier-api/internal/config"
)

func TestHealthHandler(t *testing.T) {
	handler := NewMux(slog.Default(), config.Config{
		ServiceName:       "verifier-api",
		Host:              "127.0.0.1",
		Port:              8082,
		LogLevel:          "INFO",
		RequestTimeout:    time.Second,
		ReadHeaderTimeout: time.Second,
		ShutdownTimeout:   time.Second,
		BuildVersion:      "test",
	})

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
	handler := NewMux(slog.Default(), config.Config{
		ServiceName:       "verifier-api",
		Host:              "127.0.0.1",
		Port:              8082,
		LogLevel:          "INFO",
		RequestTimeout:    time.Second,
		ReadHeaderTimeout: time.Second,
		ShutdownTimeout:   time.Second,
		BuildVersion:      "test",
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/verifier/policy-requests/kyc-passport-basic", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	testutil.AssertJSONMatchesFixture(t, recorder.Body.Bytes(), "schemas/examples/verifier/verifier-policy-request.kyc-passport-basic.json")
}

func TestVerifierStubResultHandler(t *testing.T) {
	handler := NewMux(slog.Default(), config.Config{
		ServiceName:       "verifier-api",
		Host:              "127.0.0.1",
		Port:              8082,
		LogLevel:          "INFO",
		RequestTimeout:    time.Second,
		ReadHeaderTimeout: time.Second,
		ShutdownTimeout:   time.Second,
		BuildVersion:      "test",
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/verifier/results/kyc-passport-basic-review/stub", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	testutil.AssertJSONMatchesFixture(t, recorder.Body.Bytes(), "schemas/examples/verifier/verifier-result.kyc-passport-basic-review.json")
}

func TestPhase1CreateVerificationHandler(t *testing.T) {
	handler := newTestVerifierHandler()
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
	handler := newTestVerifierHandler()
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
	handler := newTestVerifierHandler()
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
	handler := newTestVerifierHandler()
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

func TestPhase1GetVerificationHandler(t *testing.T) {
	handler := newTestVerifierHandler()

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

	readRequest := httptest.NewRequest(http.MethodGet, "/v1/verifier/verifications/"+placeholderVerificationID, nil)
	setVerifierPhase1Headers(readRequest, []string{verifierReadScope})
	readRecorder := httptest.NewRecorder()
	handler.ServeHTTP(readRecorder, readRequest)

	if readRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", readRecorder.Code)
	}

	testutil.AssertJSONMatchesFixture(t, readRecorder.Body.Bytes(), "schemas/examples/verifier/verification-result.allow.json")
}

func newTestVerifierHandler() http.Handler {
	return NewMux(slog.Default(), config.Config{
		ServiceName:       "verifier-api",
		Host:              "127.0.0.1",
		Port:              8082,
		LogLevel:          "INFO",
		RequestTimeout:    time.Second,
		ReadHeaderTimeout: time.Second,
		ShutdownTimeout:   time.Second,
		BuildVersion:      "test",
	})
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
