package phase1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type staticTokenSource string

func (s staticTokenSource) Token(context.Context) (string, error) {
	return string(s), nil
}

type failingTokenSource struct {
	err error
}

func (s failingTokenSource) Token(context.Context) (string, error) {
	return "", s.err
}

func TestTrustReadClientLoadsIssuerTrustRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/v1/phase1/issuers/did:web:issuer.hdip.dev/trust" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if authorizationHeader := r.Header.Get("Authorization"); authorizationHeader != "Bearer trust-runtime-test-token" {
			t.Fatalf("unexpected authorization header: %q", authorizationHeader)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issuerId":"did:web:issuer.hdip.dev","trustState":"active","allowedTemplateIds":["hdip-passport-basic"],"verificationKeyReferences":["key:issuer.hdip.dev:2026-04"]}`))
	}))
	defer server.Close()

	client, err := NewTrustReadClient(server.URL, staticTokenSource("trust-runtime-test-token"), server.Client())
	if err != nil {
		t.Fatalf("new trust client: %v", err)
	}

	record, err := client.GetIssuerTrustRecord(context.Background(), "did:web:issuer.hdip.dev")
	if err != nil {
		t.Fatalf("get issuer trust record: %v", err)
	}

	if record.IssuerID != "did:web:issuer.hdip.dev" || record.TrustState != "active" {
		t.Fatalf("unexpected record: %+v", record)
	}
}

func TestTrustReadClientReturnsNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	client, err := NewTrustReadClient(server.URL, staticTokenSource("trust-runtime-test-token"), server.Client())
	if err != nil {
		t.Fatalf("new trust client: %v", err)
	}

	_, err = client.GetIssuerTrustRecord(context.Background(), "did:web:issuer.hdip.dev")
	if !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestTrustReadClientReturnsUnauthorizedOnMissingOrInvalidInternalAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	client, err := NewTrustReadClient(server.URL, staticTokenSource("wrong-token"), server.Client())
	if err != nil {
		t.Fatalf("new trust client: %v", err)
	}

	_, err = client.GetIssuerTrustRecord(context.Background(), "did:web:issuer.hdip.dev")
	if !errors.Is(err, ErrTrustRuntimeUnauthorized) {
		t.Fatalf("expected ErrTrustRuntimeUnauthorized, got %v", err)
	}
}

func TestTrustReadClientFailsWhenTokenAcquisitionFails(t *testing.T) {
	client, err := NewTrustReadClient("http://127.0.0.1:8083", failingTokenSource{err: fmt.Errorf("token endpoint unavailable")}, nil)
	if err != nil {
		t.Fatalf("new trust client: %v", err)
	}

	_, err = client.GetIssuerTrustRecord(context.Background(), "did:web:issuer.hdip.dev")
	if err == nil || !strings.Contains(err.Error(), "acquire trust runtime access token") {
		t.Fatalf("expected token acquisition error, got %v", err)
	}
}
