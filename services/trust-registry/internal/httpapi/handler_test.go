package httpapi

import (
	"context"
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

const trustRegistryTestToken = "trust-runtime-test-token"

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
		InternalAuthToken: trustRegistryTestToken,
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
		InternalAuthToken: trustRegistryTestToken,
		BuildVersion:      "test",
	}, newPhase1TrustHandlerWithStoreAndToken(store, trustRegistryTestToken))

	request := httptest.NewRequest(http.MethodGet, "/internal/v1/phase1/issuers/did:web:issuer.hdip.dev/trust", nil)
	request.Header.Set("Authorization", "Bearer "+trustRegistryTestToken)
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
		InternalAuthToken: trustRegistryTestToken,
		BuildVersion:      "test",
	}, newPhase1TrustHandlerWithStoreAndToken(store, trustRegistryTestToken))

	request := httptest.NewRequest(http.MethodGet, "/internal/v1/phase1/issuers/did:web:issuer.hdip.dev/trust", nil)
	request.Header.Set("Authorization", "Bearer "+trustRegistryTestToken)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestInternalPhase1IssuerTrustHandlerRejectsMissingInternalAuth(t *testing.T) {
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
		InternalAuthToken: trustRegistryTestToken,
		BuildVersion:      "test",
	}, newPhase1TrustHandlerWithStoreAndToken(store, trustRegistryTestToken))

	request := httptest.NewRequest(http.MethodGet, "/internal/v1/phase1/issuers/did:web:issuer.hdip.dev/trust", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestInternalPhase1IssuerTrustBootstrapAppliesOwnedRecordsAndAudits(t *testing.T) {
	store, err := phase1.OpenRuntimeStore(t.TempDir() + "\\trust-phase1-state.json")
	if err != nil {
		t.Fatalf("open trust store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	now := time.Date(2026, time.April, 22, 9, 30, 0, 0, time.UTC)
	result, err := phase1.ApplyBootstrapDocument(context.Background(), store, "trust-bootstrap.json", phase1.BootstrapDocument{
		Issuers: []phase1.BootstrapIssuerRecord{
			{
				IssuerID:                  "did:web:issuer.hdip.dev",
				DisplayName:               "HDIP Passport Issuer",
				TrustState:                "active",
				AllowedTemplateIDs:        []string{"hdip-passport-basic"},
				VerificationKeyReferences: []string{"key:issuer.hdip.dev:2026-04"},
			},
		},
	}, now)
	if err != nil {
		t.Fatalf("apply bootstrap document: %v", err)
	}

	if result.Applied != 1 {
		t.Fatalf("expected 1 applied issuer, got %d", result.Applied)
	}

	record, err := store.GetIssuerRecord(context.Background(), "did:web:issuer.hdip.dev")
	if err != nil {
		t.Fatalf("load issuer record: %v", err)
	}

	if record.CreatedAt != now || record.UpdatedAt != now {
		t.Fatalf("unexpected timestamps: %+v", record)
	}

	audits, err := store.ListAuditRecords(context.Background())
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}

	if len(audits) != 1 {
		t.Fatalf("expected 1 audit record, got %d", len(audits))
	}

	if audits[0].Action != "trust-registry.phase1.bootstrap.apply" || audits[0].Actor.PrincipalID != "trust-registry-bootstrap" {
		t.Fatalf("unexpected audit record: %+v", audits[0])
	}
}
