package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("HDIP_HOST", "")
	t.Setenv("HDIP_PORT", "")
	t.Setenv("HDIP_TRUST_REGISTRY_INTERNAL_AUTH_TOKEN", "trust-runtime-test-token")

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
		ServiceName:          serviceName,
		Host:                 "127.0.0.1",
		Port:                 -1,
		LogLevel:             "INFO",
		RequestTimeout:       time.Second,
		ReadHeaderTimeout:    time.Second,
		ShutdownTimeout:      time.Second,
		Phase1DatabaseDriver: "pgx",
		Phase1StatePath:      "phase1-state.json",
		InternalAuthToken:    "trust-runtime-test-token",
		BuildVersion:         "dev",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestLoadReadsDatabaseSettings(t *testing.T) {
	t.Setenv("HDIP_PHASE1_DATABASE_DRIVER", "sqlite")
	t.Setenv("HDIP_PHASE1_DATABASE_URL", "file:test.db?mode=memory&cache=shared")
	t.Setenv("HDIP_TRUST_REGISTRY_INTERNAL_AUTH_TOKEN", "trust-runtime-test-token")
	t.Setenv("HDIP_TRUST_REGISTRY_BOOTSTRAP_PATH", "trust-bootstrap.json")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Phase1DatabaseDriver != "sqlite" || cfg.Phase1DatabaseURL == "" || cfg.TrustBootstrapPath != "trust-bootstrap.json" {
		t.Fatalf("unexpected database config: %+v", cfg)
	}
}

func TestLoadRejectsMalformedPortEnv(t *testing.T) {
	t.Setenv("HDIP_PORT", "registry-port")
	t.Setenv("HDIP_TRUST_REGISTRY_INTERNAL_AUTH_TOKEN", "trust-runtime-test-token")

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
	t.Setenv("HDIP_TRUST_REGISTRY_INTERNAL_AUTH_TOKEN", "trust-runtime-test-token")

	_, err := Load()
	if err == nil {
		t.Fatal("expected malformed duration error")
	}

	if !strings.Contains(err.Error(), "HDIP_REQUEST_TIMEOUT") {
		t.Fatalf("expected HDIP_REQUEST_TIMEOUT in error, got %v", err)
	}
}

func TestLoadRejectsMissingInternalAuthToken(t *testing.T) {
	t.Setenv("HDIP_TRUST_REGISTRY_INTERNAL_AUTH_TOKEN", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected missing internal auth token error")
	}

	if !strings.Contains(err.Error(), "internal auth token") {
		t.Fatalf("expected internal auth token error, got %v", err)
	}
}
