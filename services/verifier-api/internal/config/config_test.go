package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("HDIP_HOST", "")
	t.Setenv("HDIP_PORT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != 8082 {
		t.Fatalf("expected default port 8082, got %d", cfg.Port)
	}
}

func TestValidateRejectsInvalidPort(t *testing.T) {
	cfg := Config{
		ServiceName:          serviceName,
		Host:                 "127.0.0.1",
		Port:                 70000,
		LogLevel:             "INFO",
		RequestTimeout:       time.Second,
		ReadHeaderTimeout:    time.Second,
		ShutdownTimeout:      time.Second,
		Phase1DatabaseDriver: "pgx",
		Phase1StatePath:      "phase1-state.json",
		TrustRegistryBaseURL: "http://127.0.0.1:8083",
		BuildVersion:         "dev",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestLoadUsesLegacyRuntimePathEnvFallback(t *testing.T) {
	t.Setenv("HDIP_PHASE1_STATE_PATH", "")
	t.Setenv("HDIP_PHASE1_RUNTIME_PATH", "legacy-phase1-state.json")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Phase1StatePath != "legacy-phase1-state.json" {
		t.Fatalf("expected legacy state path fallback, got %q", cfg.Phase1StatePath)
	}
}

func TestLoadReadsTrustRegistryAndDatabaseSettings(t *testing.T) {
	t.Setenv("HDIP_PHASE1_DATABASE_DRIVER", "sqlite")
	t.Setenv("HDIP_PHASE1_DATABASE_URL", "file:test.db?mode=memory&cache=shared")
	t.Setenv("HDIP_TRUST_REGISTRY_BASE_URL", "http://127.0.0.1:19083")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Phase1DatabaseDriver != "sqlite" || cfg.TrustRegistryBaseURL != "http://127.0.0.1:19083" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestLoadRejectsMalformedPortEnv(t *testing.T) {
	t.Setenv("HDIP_PORT", "not-a-port")

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
	t.Setenv("HDIP_REQUEST_TIMEOUT", "forever-ish")

	_, err := Load()
	if err == nil {
		t.Fatal("expected malformed duration error")
	}

	if !strings.Contains(err.Error(), "HDIP_REQUEST_TIMEOUT") {
		t.Fatalf("expected HDIP_REQUEST_TIMEOUT in error, got %v", err)
	}
}
