package config

import (
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
		BuildVersion:      "dev",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}
