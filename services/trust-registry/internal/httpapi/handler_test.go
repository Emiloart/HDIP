package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Emiloart/HDIP/packages/go/foundation/httpx"
	"github.com/Emiloart/HDIP/services/trust-registry/internal/config"
	phase1 "github.com/Emiloart/HDIP/services/trust-registry/internal/phase1"
)

func TestHealthHandler(t *testing.T) {
	handler, err := NewMux(slog.Default(), config.Config{
		ServiceName:       "trust-registry",
		Host:              "127.0.0.1",
		Port:              8083,
		LogLevel:          "INFO",
		RequestTimeout:    time.Second,
		ReadHeaderTimeout: time.Second,
		ShutdownTimeout:   time.Second,
		Phase1StatePath:   t.TempDir() + "\\trust-phase1-state.json",
		BuildVersion:      "test",
	})
	if err != nil {
		t.Fatalf("new trust mux: %v", err)
	}

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

	if response.Service != "trust-registry" || response.Status != "ok" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestInternalPhase1IssuerTrustHandler(t *testing.T) {
	store, err := phase1.OpenRuntimeStore(t.TempDir() + "\\trust-phase1-state.json")
	if err != nil {
		t.Fatalf("open trust store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	if err := store.SeedIssuerRecord(phase1.IssuerRecord{
		IssuerID:                  "did:web:issuer.hdip.dev",
		DisplayName:               "HDIP Passport Issuer",
		TrustState:                "active",
		AllowedTemplateIDs:        []string{"hdip-passport-basic"},
		VerificationKeyReferences: []string{"key:issuer.hdip.dev:2026-04"},
	}); err != nil {
		t.Fatalf("seed issuer record: %v", err)
	}

	handler := newMuxWithPhase1Handler(slog.Default(), config.Config{
		ServiceName:       "trust-registry",
		Host:              "127.0.0.1",
		Port:              8083,
		LogLevel:          "INFO",
		RequestTimeout:    time.Second,
		ReadHeaderTimeout: time.Second,
		ShutdownTimeout:   time.Second,
		Phase1StatePath:   t.TempDir() + "\\unused-state.json",
		BuildVersion:      "test",
	}, newPhase1TrustHandlerWithStore(store))

	request := httptest.NewRequest(http.MethodGet, "/internal/v1/phase1/issuers/did:web:issuer.hdip.dev/trust", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var payload issuerTrustPayload
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.IssuerID != "did:web:issuer.hdip.dev" || payload.TrustState != "active" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestInternalPhase1IssuerTrustHandlerReturnsNotFound(t *testing.T) {
	store, err := phase1.OpenRuntimeStore(t.TempDir() + "\\trust-phase1-state.json")
	if err != nil {
		t.Fatalf("open trust store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	handler := newMuxWithPhase1Handler(slog.Default(), config.Config{
		ServiceName:       "trust-registry",
		Host:              "127.0.0.1",
		Port:              8083,
		LogLevel:          "INFO",
		RequestTimeout:    time.Second,
		ReadHeaderTimeout: time.Second,
		ShutdownTimeout:   time.Second,
		Phase1StatePath:   t.TempDir() + "\\unused-state.json",
		BuildVersion:      "test",
	}, newPhase1TrustHandlerWithStore(store))

	request := httptest.NewRequest(http.MethodGet, "/internal/v1/phase1/issuers/did:web:issuer.hdip.dev/trust", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}
