package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
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
