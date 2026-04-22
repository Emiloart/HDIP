package phase1

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestHydraClientCredentialsTokenSourceRequestsBearerTokenAndCachesIt(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++

		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/oauth2/token" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if contentType := r.Header.Get("Content-Type"); contentType != "application/x-www-form-urlencoded" {
			t.Fatalf("unexpected content type: %s", contentType)
		}
		if authorization := r.Header.Get("Authorization"); authorization != "Basic "+base64.StdEncoding.EncodeToString([]byte("verifier-api:secret")) {
			t.Fatalf("unexpected authorization header: %s", authorization)
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if r.Form.Get("grant_type") != "client_credentials" {
			t.Fatalf("unexpected grant_type: %s", r.Form.Get("grant_type"))
		}
		if r.Form.Get("scope") != "trust.runtime.read" {
			t.Fatalf("unexpected scope: %s", r.Form.Get("scope"))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"hydra-access-token","token_type":"bearer","expires_in":120}`))
	}))
	defer server.Close()

	tokenSource, err := NewHydraClientCredentialsTokenSource(
		server.URL+"/oauth2/token",
		"verifier-api",
		"secret",
		"trust.runtime.read",
		server.Client(),
	)
	if err != nil {
		t.Fatalf("new hydra token source: %v", err)
	}

	firstToken, err := tokenSource.Token(context.Background())
	if err != nil {
		t.Fatalf("load first token: %v", err)
	}
	secondToken, err := tokenSource.Token(context.Background())
	if err != nil {
		t.Fatalf("load second token: %v", err)
	}

	if firstToken != "hydra-access-token" || secondToken != "hydra-access-token" {
		t.Fatalf("unexpected token values: %q %q", firstToken, secondToken)
	}
	if requests != 1 {
		t.Fatalf("expected single token request due to cache, got %d", requests)
	}
}

func TestHydraClientCredentialsTokenSourceRejectsInvalidConfig(t *testing.T) {
	_, err := NewHydraClientCredentialsTokenSource(
		"",
		"verifier-api",
		"secret",
		"trust.runtime.read",
		nil,
	)
	if err == nil || !strings.Contains(err.Error(), "token url") {
		t.Fatalf("expected token url validation error, got %v", err)
	}
}

func TestHydraClientCredentialsTokenSourceFailsOnTokenEndpointError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	tokenSource, err := NewHydraClientCredentialsTokenSource(
		server.URL,
		"verifier-api",
		"secret",
		"trust.runtime.read",
		server.Client(),
	)
	if err != nil {
		t.Fatalf("new hydra token source: %v", err)
	}

	_, err = tokenSource.Token(context.Background())
	if err == nil || !strings.Contains(err.Error(), "returned 401") {
		t.Fatalf("expected token endpoint status error, got %v", err)
	}
}

func TestHydraClientCredentialsTokenSourceAcceptsParsedURL(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}

	if _, err := NewHydraClientCredentialsTokenSource(parsedURL.String(), "verifier-api", "secret", "trust.runtime.read", server.Client()); err != nil {
		t.Fatalf("expected parsed url to be accepted, got %v", err)
	}
}

func TestHydraClientCredentialsTokenSourceCheckFailsClosed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "hydra unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	tokenSource, err := NewHydraClientCredentialsTokenSource(
		server.URL,
		"verifier-api",
		"secret",
		"trust.runtime.read",
		server.Client(),
	)
	if err != nil {
		t.Fatalf("new token source: %v", err)
	}

	err = tokenSource.Check(context.Background())
	if !errors.Is(err, ErrHydraTokenUnavailable) {
		t.Fatalf("expected ErrHydraTokenUnavailable, got %v", err)
	}
}
