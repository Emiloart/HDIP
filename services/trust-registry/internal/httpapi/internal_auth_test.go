package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHydraIntrospectionAuthorizerAcceptsAuthorizedVerifierClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if token := r.Form.Get("token"); token != trustRegistryTestToken && token != "phase1-readiness-probe" {
			t.Fatalf("unexpected token: %s", r.Form.Get("token"))
		}
		if r.Form.Get("token_type_hint") != "access_token" {
			t.Fatalf("unexpected token type hint: %s", r.Form.Get("token_type_hint"))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"active":true,"client_id":"verifier-api","scope":"trust.runtime.read other.scope"}`))
	}))
	defer server.Close()

	authorizer, err := newHydraIntrospectionAuthorizer(
		server.URL,
		"trust-registry",
		"secret",
		"verifier-api",
		"trust.runtime.read",
		server.Client(),
	)
	if err != nil {
		t.Fatalf("new hydra introspection authorizer: %v", err)
	}

	principal, err := authorizer.Authorize(context.Background(), trustRegistryTestToken)
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}

	if principal.ClientID != "verifier-api" {
		t.Fatalf("unexpected principal: %+v", principal)
	}

	if err := authorizer.Check(context.Background()); err != nil {
		t.Fatalf("check readiness: %v", err)
	}
}

func TestHydraIntrospectionAuthorizerRejectsInactiveToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"active":false}`))
	}))
	defer server.Close()

	authorizer, err := newHydraIntrospectionAuthorizer(
		server.URL,
		"trust-registry",
		"secret",
		"verifier-api",
		"trust.runtime.read",
		server.Client(),
	)
	if err != nil {
		t.Fatalf("new hydra introspection authorizer: %v", err)
	}

	_, err = authorizer.Authorize(context.Background(), trustRegistryTestToken)
	if !errors.Is(err, ErrInternalAuthUnauthenticated) {
		t.Fatalf("expected unauthenticated error, got %v", err)
	}
}

func TestHydraIntrospectionAuthorizerRejectsWrongClientOrMissingScope(t *testing.T) {
	testCases := []struct {
		name         string
		responseBody string
	}{
		{
			name:         "wrong client id",
			responseBody: `{"active":true,"client_id":"issuer-api","scope":"trust.runtime.read"}`,
		},
		{
			name:         "missing scope",
			responseBody: `{"active":true,"client_id":"verifier-api","scope":"other.scope"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(testCase.responseBody))
			}))
			defer server.Close()

			authorizer, err := newHydraIntrospectionAuthorizer(
				server.URL,
				"trust-registry",
				"secret",
				"verifier-api",
				"trust.runtime.read",
				server.Client(),
			)
			if err != nil {
				t.Fatalf("new hydra introspection authorizer: %v", err)
			}

			_, err = authorizer.Authorize(context.Background(), trustRegistryTestToken)
			if !errors.Is(err, ErrInternalAuthForbidden) {
				t.Fatalf("expected forbidden error, got %v", err)
			}
		})
	}
}

func TestHydraIntrospectionAuthorizerFailsClosedWhenIntrospectionUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "hydra unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	authorizer, err := newHydraIntrospectionAuthorizer(
		server.URL,
		"trust-registry",
		"secret",
		"verifier-api",
		"trust.runtime.read",
		server.Client(),
	)
	if err != nil {
		t.Fatalf("new hydra introspection authorizer: %v", err)
	}

	_, err = authorizer.Authorize(context.Background(), trustRegistryTestToken)
	if !errors.Is(err, ErrInternalAuthUnavailable) {
		t.Fatalf("expected unavailable error, got %v", err)
	}
	if !strings.Contains(err.Error(), "returned 503") {
		t.Fatalf("expected hydra status details, got %v", err)
	}
}
