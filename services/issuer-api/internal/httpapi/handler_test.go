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
	"github.com/Emiloart/HDIP/services/issuer-api/internal/config"
)

func TestHealthHandler(t *testing.T) {
	handler := NewMux(slog.Default(), config.Config{
		ServiceName:       "issuer-api",
		Host:              "127.0.0.1",
		Port:              8081,
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

	if response.Service != "issuer-api" || response.Status != "ok" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestIssuerProfileHandler(t *testing.T) {
	handler := NewMux(slog.Default(), config.Config{
		ServiceName:       "issuer-api",
		Host:              "127.0.0.1",
		Port:              8081,
		LogLevel:          "INFO",
		RequestTimeout:    time.Second,
		ReadHeaderTimeout: time.Second,
		ShutdownTimeout:   time.Second,
		BuildVersion:      "test",
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/issuer/profile", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	testutil.AssertJSONMatchesFixture(t, recorder.Body.Bytes(), "schemas/examples/issuer/issuer-profile.default.json")
}

func TestIssuerTemplateHandler(t *testing.T) {
	handler := NewMux(slog.Default(), config.Config{
		ServiceName:       "issuer-api",
		Host:              "127.0.0.1",
		Port:              8081,
		LogLevel:          "INFO",
		RequestTimeout:    time.Second,
		ReadHeaderTimeout: time.Second,
		ShutdownTimeout:   time.Second,
		BuildVersion:      "test",
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/issuer/templates/hdip-passport-basic", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	testutil.AssertJSONMatchesFixture(t, recorder.Body.Bytes(), "schemas/examples/credentials/credential-template-metadata.hdip-passport-basic.json")
}

func TestPhase1IssueCredentialHandler(t *testing.T) {
	handler := newTestIssuerHandler()
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
	handler := newTestIssuerHandler()
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
	handler := newTestIssuerHandler()
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

func TestPhase1IssueCredentialRejectsInvalidPayload(t *testing.T) {
	handler := newTestIssuerHandler()
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
	handler := newTestIssuerHandler()

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

	readRequest := httptest.NewRequest(http.MethodGet, "/v1/issuer/credentials/"+placeholderCredentialID, nil)
	setIssuerPhase1Headers(readRequest, []string{issuerReadScope})
	readRecorder := httptest.NewRecorder()
	handler.ServeHTTP(readRecorder, readRequest)

	if readRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", readRecorder.Code)
	}

	testutil.AssertJSONMatchesFixture(t, readRecorder.Body.Bytes(), "schemas/examples/credentials/credential-record.hdip-passport-basic.json")
}

func newTestIssuerHandler() http.Handler {
	return NewMux(slog.Default(), config.Config{
		ServiceName:       "issuer-api",
		Host:              "127.0.0.1",
		Port:              8081,
		LogLevel:          "INFO",
		RequestTimeout:    time.Second,
		ReadHeaderTimeout: time.Second,
		ShutdownTimeout:   time.Second,
		BuildVersion:      "test",
	})
}

func setIssuerPhase1Headers(request *http.Request, scopes []string) {
	request.Header.Set("X-HDIP-Principal-ID", "issuer_operator_alex")
	request.Header.Set("X-HDIP-Organization-ID", "did:web:issuer.hdip.dev")
	request.Header.Set("X-HDIP-Auth-Reference", "session_issuer_001")
	request.Header.Set("X-HDIP-Scopes", strings.Join(scopes, ","))
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
