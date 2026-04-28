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
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_TOKEN_URL", "http://127.0.0.1:4444/oauth2/token")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_CLIENT_ID", "verifier-api")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_CLIENT_SECRET", "trust-runtime-test-secret")

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
		ServiceName:                   serviceName,
		Host:                          "127.0.0.1",
		Port:                          70000,
		LogLevel:                      "INFO",
		RequestTimeout:                time.Second,
		ReadHeaderTimeout:             time.Second,
		ShutdownTimeout:               time.Second,
		Phase1DatabaseDriver:          "pgx",
		Phase1DatabaseURL:             "postgres://phase1",
		TrustRegistryBaseURL:          "http://127.0.0.1:8083",
		TrustRuntimeHydraTokenURL:     "http://127.0.0.1:4444/oauth2/token",
		TrustRuntimeHydraClientID:     "verifier-api",
		TrustRuntimeHydraClientSecret: "trust-runtime-test-secret",
		TrustRuntimeHydraScope:        "trust.runtime.read",
		BuildVersion:                  "dev",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestLoadReadsTrustRegistryAndDatabaseSettings(t *testing.T) {
	t.Setenv("HDIP_PHASE1_DATABASE_DRIVER", "sqlite")
	t.Setenv("HDIP_PHASE1_DATABASE_URL", "file:test.db?mode=memory&cache=shared")
	t.Setenv("HDIP_TRUST_REGISTRY_BASE_URL", "http://127.0.0.1:19083")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_TOKEN_URL", "http://127.0.0.1:4444/oauth2/token")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_CLIENT_ID", "verifier-api")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_CLIENT_SECRET", "trust-runtime-test-secret")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_SCOPE", "trust.runtime.read")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Phase1DatabaseDriver != "sqlite" ||
		cfg.TrustRegistryBaseURL != "http://127.0.0.1:19083" ||
		cfg.TrustRuntimeHydraTokenURL != "http://127.0.0.1:4444/oauth2/token" ||
		cfg.TrustRuntimeHydraClientID != "verifier-api" ||
		cfg.TrustRuntimeHydraScope != "trust.runtime.read" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestLoadRejectsMalformedPortEnv(t *testing.T) {
	t.Setenv("HDIP_PORT", "not-a-port")
	t.Setenv("HDIP_PHASE1_DATABASE_URL", "postgres://phase1")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_TOKEN_URL", "http://127.0.0.1:4444/oauth2/token")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_CLIENT_ID", "verifier-api")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_CLIENT_SECRET", "trust-runtime-test-secret")

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
	t.Setenv("HDIP_PHASE1_DATABASE_URL", "postgres://phase1")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_TOKEN_URL", "http://127.0.0.1:4444/oauth2/token")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_CLIENT_ID", "verifier-api")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_CLIENT_SECRET", "trust-runtime-test-secret")

	_, err := Load()
	if err == nil {
		t.Fatal("expected malformed duration error")
	}

	if !strings.Contains(err.Error(), "HDIP_REQUEST_TIMEOUT") {
		t.Fatalf("expected HDIP_REQUEST_TIMEOUT in error, got %v", err)
	}
}

func TestLoadRejectsMissingTrustRuntimeHydraTokenConfig(t *testing.T) {
	t.Setenv("HDIP_PHASE1_DATABASE_URL", "postgres://phase1")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_TOKEN_URL", "")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_CLIENT_ID", "")
	t.Setenv("HDIP_TRUST_RUNTIME_HYDRA_CLIENT_SECRET", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected missing trust runtime hydra config error")
	}

	if !strings.Contains(err.Error(), "trust runtime hydra token url") {
		t.Fatalf("expected hydra token url error, got %v", err)
	}
}

func TestValidateRejectsSQLPrimaryWithoutDatabaseURL(t *testing.T) {
	cfg := Config{
		ServiceName:                   serviceName,
		Host:                          "127.0.0.1",
		Port:                          8082,
		LogLevel:                      "INFO",
		RequestTimeout:                time.Second,
		ReadHeaderTimeout:             time.Second,
		ShutdownTimeout:               time.Second,
		TrustRegistryBaseURL:          "http://127.0.0.1:8083",
		TrustRuntimeHydraTokenURL:     "http://127.0.0.1:4444/oauth2/token",
		TrustRuntimeHydraClientID:     "verifier-api",
		TrustRuntimeHydraClientSecret: "trust-runtime-test-secret",
		TrustRuntimeHydraScope:        "trust.runtime.read",
		BuildVersion:                  "dev",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected sql-primary validation error")
	}

	if !strings.Contains(err.Error(), "HDIP_PHASE1_DATABASE_URL") {
		t.Fatalf("expected database url validation error, got %v", err)
	}
}
