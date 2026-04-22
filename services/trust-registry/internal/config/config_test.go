package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("HDIP_HOST", "")
	t.Setenv("HDIP_PORT", "")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_URL", "http://127.0.0.1:4445/admin/oauth2/introspect")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_CLIENT_ID", "trust-registry")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_CLIENT_SECRET", "trust-runtime-test-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != 8083 {
		t.Fatalf("expected default port 8083, got %d", cfg.Port)
	}
}

func TestValidateRejectsInvalidPort(t *testing.T) {
	cfg := Config{
		ServiceName:                       serviceName,
		Host:                              "127.0.0.1",
		Port:                              -1,
		LogLevel:                          "INFO",
		RequestTimeout:                    time.Second,
		ReadHeaderTimeout:                 time.Second,
		ShutdownTimeout:                   time.Second,
		Phase1DatabaseDriver:              "pgx",
		Phase1StatePath:                   "phase1-state.json",
		TrustRuntimeHydraIntrospectionURL: "http://127.0.0.1:4445/admin/oauth2/introspect",
		TrustRuntimeHydraClientID:         "trust-registry",
		TrustRuntimeHydraClientSecret:     "trust-runtime-test-secret",
		TrustRuntimeHydraExpectedClientID: "verifier-api",
		TrustRuntimeHydraRequiredScope:    "trust.runtime.read",
		BuildVersion:                      "dev",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestLoadReadsDatabaseSettings(t *testing.T) {
	t.Setenv("HDIP_PHASE1_DATABASE_DRIVER", "sqlite")
	t.Setenv("HDIP_PHASE1_DATABASE_URL", "file:test.db?mode=memory&cache=shared")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_URL", "http://127.0.0.1:4445/admin/oauth2/introspect")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_CLIENT_ID", "trust-registry")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_CLIENT_SECRET", "trust-runtime-test-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Phase1DatabaseDriver != "sqlite" ||
		cfg.Phase1DatabaseURL == "" ||
		cfg.TrustRuntimeHydraIntrospectionURL != "http://127.0.0.1:4445/admin/oauth2/introspect" ||
		cfg.TrustRuntimeHydraExpectedClientID != "verifier-api" ||
		cfg.TrustRuntimeHydraRequiredScope != "trust.runtime.read" {
		t.Fatalf("unexpected database config: %+v", cfg)
	}
}

func TestLoadRejectsMalformedPortEnv(t *testing.T) {
	t.Setenv("HDIP_PORT", "registry-port")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_URL", "http://127.0.0.1:4445/admin/oauth2/introspect")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_CLIENT_ID", "trust-registry")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_CLIENT_SECRET", "trust-runtime-test-secret")

	_, err := Load()
	if err == nil {
		t.Fatal("expected malformed port error")
	}

	if !strings.Contains(err.Error(), "HDIP_PORT") {
		t.Fatalf("expected HDIP_PORT in error, got %v", err)
	}
}

func TestLoadRejectsMalformedDurationEnv(t *testing.T) {
	t.Setenv("HDIP_PORT", "")
	t.Setenv("HDIP_REQUEST_TIMEOUT", "not-a-duration")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_URL", "http://127.0.0.1:4445/admin/oauth2/introspect")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_CLIENT_ID", "trust-registry")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_CLIENT_SECRET", "trust-runtime-test-secret")

	_, err := Load()
	if err == nil {
		t.Fatal("expected malformed duration error")
	}

	if !strings.Contains(err.Error(), "HDIP_REQUEST_TIMEOUT") {
		t.Fatalf("expected HDIP_REQUEST_TIMEOUT in error, got %v", err)
	}
}

func TestLoadRejectsMissingHydraIntrospectionConfig(t *testing.T) {
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_URL", "")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_CLIENT_ID", "")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_CLIENT_SECRET", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected missing hydra introspection config error")
	}

	if !strings.Contains(err.Error(), "trust runtime hydra introspection url") {
		t.Fatalf("expected hydra introspection url error, got %v", err)
	}
}

func TestValidateRejectsBootstrapPathOnPrimarySQLPath(t *testing.T) {
	cfg := Config{
		ServiceName:                       serviceName,
		Host:                              "127.0.0.1",
		Port:                              8083,
		LogLevel:                          "INFO",
		RequestTimeout:                    time.Second,
		ReadHeaderTimeout:                 time.Second,
		ShutdownTimeout:                   time.Second,
		Phase1DatabaseDriver:              "pgx",
		Phase1DatabaseURL:                 "postgres://phase1",
		Phase1StatePath:                   "phase1-state.json",
		TrustBootstrapPath:                "trust-bootstrap.json",
		TrustRuntimeHydraIntrospectionURL: "http://127.0.0.1:4445/admin/oauth2/introspect",
		TrustRuntimeHydraClientID:         "trust-registry",
		TrustRuntimeHydraClientSecret:     "trust-runtime-test-secret",
		TrustRuntimeHydraExpectedClientID: "verifier-api",
		TrustRuntimeHydraRequiredScope:    "trust.runtime.read",
		BuildVersion:                      "dev",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected bootstrap path validation error")
	}

	if !strings.Contains(err.Error(), "phase1sql CLI") {
		t.Fatalf("expected phase1sql cli validation error, got %v", err)
	}
}
