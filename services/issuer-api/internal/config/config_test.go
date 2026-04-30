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
	if cfg.PublicAuthMode != "header" {
		t.Fatalf("expected default public auth mode header, got %q", cfg.PublicAuthMode)
	}
}

func TestValidateRejectsInvalidPort(t *testing.T) {
	cfg := Config{
		ServiceName:          serviceName,
		Host:                 "127.0.0.1",
		Port:                 0,
		LogLevel:             "INFO",
		Environment:          "development",
		RequestTimeout:       time.Second,
		ReadHeaderTimeout:    time.Second,
		ShutdownTimeout:      time.Second,
		Phase1DatabaseDriver: "pgx",
		Phase1DatabaseURL:    "postgres://phase1",
		PublicAuthMode:       "header",
		BuildVersion:         "dev",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
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

func TestLoadReadsHydraPublicAuthSettings(t *testing.T) {
	t.Setenv("HDIP_PHASE1_DATABASE_URL", "postgres://phase1")
	t.Setenv("HDIP_PUBLIC_AUTH_MODE", "hydra")
	t.Setenv("HDIP_PUBLIC_AUTH_HYDRA_INTROSPECTION_URL", "http://127.0.0.1:4445/admin/oauth2/introspect")
	t.Setenv("HDIP_PUBLIC_AUTH_HYDRA_INTROSPECTION_CLIENT_ID", "issuer-api")
	t.Setenv("HDIP_PUBLIC_AUTH_HYDRA_INTROSPECTION_CLIENT_SECRET", "issuer-introspection-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.PublicAuthMode != "hydra" ||
		cfg.PublicAuthHydraIntrospectionURL != "http://127.0.0.1:4445/admin/oauth2/introspect" ||
		cfg.PublicAuthHydraIntrospectionClientID != "issuer-api" {
		t.Fatalf("unexpected public auth config: %+v", cfg)
	}
}

func TestLoadRejectsHydraPublicAuthWithoutIntrospectionConfig(t *testing.T) {
	t.Setenv("HDIP_PHASE1_DATABASE_URL", "postgres://phase1")
	t.Setenv("HDIP_PUBLIC_AUTH_MODE", "hydra")

	_, err := Load()
	if err == nil {
		t.Fatal("expected missing hydra public auth config error")
	}

	if !strings.Contains(err.Error(), "HDIP_PUBLIC_AUTH_HYDRA_INTROSPECTION_URL") {
		t.Fatalf("expected hydra introspection url error, got %v", err)
	}
}

func TestValidateRejectsHeaderPublicAuthInProduction(t *testing.T) {
	cfg := Config{
		ServiceName:          serviceName,
		Host:                 "127.0.0.1",
		Port:                 8081,
		LogLevel:             "INFO",
		Environment:          "production",
		RequestTimeout:       time.Second,
		ReadHeaderTimeout:    time.Second,
		ShutdownTimeout:      time.Second,
		Phase1DatabaseDriver: "pgx",
		Phase1DatabaseURL:    "postgres://phase1",
		PublicAuthMode:       "header",
		BuildVersion:         "dev",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected production header auth validation error")
	}
	if !strings.Contains(err.Error(), "HDIP_ENVIRONMENT=production") {
		t.Fatalf("expected production validation error, got %v", err)
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
		Environment:       "development",
		RequestTimeout:    time.Second,
		ReadHeaderTimeout: time.Second,
		ShutdownTimeout:   time.Second,
		PublicAuthMode:    "header",
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
