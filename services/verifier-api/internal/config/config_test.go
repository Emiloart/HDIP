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
		ServiceName:       serviceName,
		Host:              "127.0.0.1",
		Port:              70000,
		LogLevel:          "INFO",
		RequestTimeout:    time.Second,
		ReadHeaderTimeout: time.Second,
		ShutdownTimeout:   time.Second,
		Phase1RuntimePath: "phase1-runtime.sqlite",
		BuildVersion:      "dev",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
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
