package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("HDIP_HOST", "")
	t.Setenv("HDIP_PORT", "")
	t.Setenv("HDIP_PHASE1_DATABASE_URL", "postgres://phase1")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != 8081 {
		t.Fatalf("expected default port 8081, got %d", cfg.Port)
	}
}

func TestValidateRejectsInvalidPort(t *testing.T) {
	cfg := Config{
		ServiceName:          serviceName,
		Host:                 "127.0.0.1",
		Port:                 0,
		LogLevel:             "INFO",
		RequestTimeout:       time.Second,
		ReadHeaderTimeout:    time.Second,
		ShutdownTimeout:      time.Second,
		Phase1RuntimeMode:    "sql-primary",
		Phase1DatabaseDriver: "pgx",
		Phase1DatabaseURL:    "postgres://phase1",
		BuildVersion:         "dev",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestLoadUsesTransitionalStatePathEnvFallback(t *testing.T) {
	t.Setenv("HDIP_PHASE1_RUNTIME_MODE", "transitional-json")
	t.Setenv("HDIP_PHASE1_TRANSITIONAL_STATE_PATH", "")
	t.Setenv("HDIP_PHASE1_STATE_PATH", "")
	t.Setenv("HDIP_PHASE1_RUNTIME_PATH", "legacy-phase1-state.json")
	t.Setenv("HDIP_PHASE1_DATABASE_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Phase1StatePath != "legacy-phase1-state.json" {
		t.Fatalf("expected legacy state path fallback, got %q", cfg.Phase1StatePath)
	}
}

func TestLoadReadsDatabaseSettings(t *testing.T) {
	t.Setenv("HDIP_PHASE1_DATABASE_DRIVER", "sqlite")
	t.Setenv("HDIP_PHASE1_DATABASE_URL", "file:test.db?mode=memory&cache=shared")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Phase1DatabaseDriver != "sqlite" || cfg.Phase1DatabaseURL == "" {
		t.Fatalf("unexpected database config: %+v", cfg)
	}
}

func TestLoadRejectsMalformedPortEnv(t *testing.T) {
	t.Setenv("HDIP_PORT", "eight-zero-eight-one")
	t.Setenv("HDIP_PHASE1_DATABASE_URL", "postgres://phase1")

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
	t.Setenv("HDIP_REQUEST_TIMEOUT", "five-seconds")
	t.Setenv("HDIP_PHASE1_DATABASE_URL", "postgres://phase1")

	_, err := Load()
	if err == nil {
		t.Fatal("expected malformed duration error")
	}

	if !strings.Contains(err.Error(), "HDIP_REQUEST_TIMEOUT") {
		t.Fatalf("expected HDIP_REQUEST_TIMEOUT in error, got %v", err)
	}
}

func TestValidateRejectsSQLPrimaryWithoutDatabaseURL(t *testing.T) {
	cfg := Config{
		ServiceName:       serviceName,
		Host:              "127.0.0.1",
		Port:              8081,
		LogLevel:          "INFO",
		RequestTimeout:    time.Second,
		ReadHeaderTimeout: time.Second,
		ShutdownTimeout:   time.Second,
		Phase1RuntimeMode: "sql-primary",
		BuildVersion:      "dev",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected sql-primary validation error")
	}

	if !strings.Contains(err.Error(), "HDIP_PHASE1_DATABASE_URL") {
		t.Fatalf("expected database url validation error, got %v", err)
	}
}
