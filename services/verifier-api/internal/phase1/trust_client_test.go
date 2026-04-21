package phase1

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrustReadClientLoadsIssuerTrustRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/v1/phase1/issuers/did:web:issuer.hdip.dev/trust" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issuerId":"did:web:issuer.hdip.dev","trustState":"active","allowedTemplateIds":["hdip-passport-basic"],"verificationKeyReferences":["key:issuer.hdip.dev:2026-04"]}`))
	}))
	defer server.Close()

	client, err := NewTrustReadClient(server.URL, server.Client())
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

	client, err := NewTrustReadClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("new trust client: %v", err)
	}

	_, err = client.GetIssuerTrustRecord(context.Background(), "did:web:issuer.hdip.dev")
	if !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}
