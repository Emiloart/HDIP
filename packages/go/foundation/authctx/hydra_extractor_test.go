package authctx

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestHydraIssuerOperatorExtractorAcceptsActiveBearerToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertIntrospectionRequest(t, r, "issuer-api", "issuer-secret", "issuer-token")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"active":true,"client_id":"did:web:issuer.hdip.dev","scope":"issuer.credentials.issue issuer.credentials.read","jti":"token-issuer-1"}`))
	}))
	t.Cleanup(server.Close)

	extractor, err := NewHydraIssuerOperatorExtractor(HydraIntrospectionConfig{
		IntrospectionURL: server.URL,
		ClientID:         "issuer-api",
		ClientSecret:     "issuer-secret",
		HTTPClient:       server.Client(),
	})
	if err != nil {
		t.Fatalf("new extractor: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("Authorization", "Bearer issuer-token")

	attribution, err := extractor.IssuerOperatorFromRequest(request)
	if err != nil {
		t.Fatalf("extract attribution: %v", err)
	}

	if attribution.PrincipalID != "did:web:issuer.hdip.dev" ||
		attribution.OrganizationID != "did:web:issuer.hdip.dev" ||
		attribution.ActorType != ActorTypeIssuerOperator ||
		attribution.AuthenticationReference != "hydra-token:token-issuer-1" {
		t.Fatalf("unexpected attribution: %+v", attribution)
	}
	if !attribution.HasScope("issuer.credentials.issue") {
		t.Fatalf("expected issuer scope in attribution: %+v", attribution)
	}
}

func TestHydraVerifierIntegratorExtractorRejectsInactiveToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"active":false}`))
	}))
	t.Cleanup(server.Close)

	extractor, err := NewHydraVerifierIntegratorExtractor(HydraIntrospectionConfig{
		IntrospectionURL: server.URL,
		ClientID:         "verifier-api",
		ClientSecret:     "verifier-secret",
		HTTPClient:       server.Client(),
	})
	if err != nil {
		t.Fatalf("new extractor: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("Authorization", "Bearer verifier-token")

	_, err = extractor.VerifierIntegratorFromRequest(request)
	if !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("expected ErrUnauthenticated, got %v", err)
	}
}

func TestHydraVerifierIntegratorExtractorRejectsMissingBearer(t *testing.T) {
	extractor, err := NewHydraVerifierIntegratorExtractor(HydraIntrospectionConfig{
		IntrospectionURL: "http://127.0.0.1/introspect",
		ClientID:         "verifier-api",
		ClientSecret:     "verifier-secret",
	})
	if err != nil {
		t.Fatalf("new extractor: %v", err)
	}

	_, err = extractor.VerifierIntegratorFromRequest(httptest.NewRequest(http.MethodPost, "/", nil))
	if !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("expected ErrUnauthenticated, got %v", err)
	}
}

func TestHydraExtractorRejectsRelativeIntrospectionURL(t *testing.T) {
	_, err := NewHydraVerifierIntegratorExtractor(HydraIntrospectionConfig{
		IntrospectionURL: "/admin/oauth2/introspect",
		ClientID:         "verifier-api",
		ClientSecret:     "verifier-secret",
	})
	if err == nil {
		t.Fatal("expected relative introspection URL to be rejected")
	}
}

func TestHydraExtractorReportsUnavailableIntrospection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "hydra unavailable", http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	extractor, err := NewHydraVerifierIntegratorExtractor(HydraIntrospectionConfig{
		IntrospectionURL: server.URL,
		ClientID:         "verifier-api",
		ClientSecret:     "verifier-secret",
		HTTPClient:       server.Client(),
	})
	if err != nil {
		t.Fatalf("new extractor: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("Authorization", "Bearer verifier-token")

	_, err = extractor.VerifierIntegratorFromRequest(request)
	if !errors.Is(err, ErrAuthUnavailable) {
		t.Fatalf("expected ErrAuthUnavailable, got %v", err)
	}
}

func TestHydraExtractorReadinessChecksIntrospectionReachability(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertIntrospectionRequest(t, r, "verifier-api", "verifier-secret", "phase1-public-auth-readiness-probe")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"active":false}`))
	}))
	t.Cleanup(server.Close)

	extractor, err := NewHydraVerifierIntegratorExtractor(HydraIntrospectionConfig{
		IntrospectionURL: server.URL,
		ClientID:         "verifier-api",
		ClientSecret:     "verifier-secret",
		HTTPClient:       server.Client(),
	})
	if err != nil {
		t.Fatalf("new extractor: %v", err)
	}

	if err := extractor.Check(context.Background()); err != nil {
		t.Fatalf("expected readiness check to pass, got %v", err)
	}
}

func assertIntrospectionRequest(t *testing.T, r *http.Request, clientID string, clientSecret string, token string) {
	t.Helper()

	if r.Method != http.MethodPost {
		t.Fatalf("unexpected method: %s", r.Method)
	}

	username, password, ok := r.BasicAuth()
	if !ok || username != clientID || password != clientSecret {
		t.Fatalf("unexpected basic auth credentials")
	}

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	form, err := url.ParseQuery(string(raw))
	if err != nil {
		t.Fatalf("parse form: %v", err)
	}
	if form.Get("token") != token {
		t.Fatalf("unexpected token: %q", form.Get("token"))
	}
	if form.Get("token_type_hint") != "access_token" {
		t.Fatalf("unexpected token_type_hint: %q", form.Get("token_type_hint"))
	}
}
