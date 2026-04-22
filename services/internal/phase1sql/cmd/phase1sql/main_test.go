package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunRejectsUnknownCommand(t *testing.T) {
	err := run(context.Background(), []string{"unknown"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("expected unknown command error, got %v", err)
	}
}

func TestRunMigrateRequiresDSN(t *testing.T) {
	err := run(context.Background(), []string{"migrate", "up"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--dsn") {
		t.Fatalf("expected dsn requirement error, got %v", err)
	}
}

func TestRunBootstrapRequiresFile(t *testing.T) {
	err := run(context.Background(), []string{"bootstrap", "trust", "--dsn", "postgres://phase1"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--file") {
		t.Fatalf("expected file requirement error, got %v", err)
	}
}
