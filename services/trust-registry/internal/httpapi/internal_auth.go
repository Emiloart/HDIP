package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrInternalAuthUnauthenticated = errors.New("internal trust runtime request unauthenticated")
	ErrInternalAuthForbidden       = errors.New("internal trust runtime request forbidden")
	ErrInternalAuthUnavailable     = errors.New("internal trust runtime auth unavailable")
)

type internalPrincipal struct {
	ClientID string
}

type internalAuthorizer interface {
	Authorize(ctx context.Context, bearerToken string) (internalPrincipal, error)
}

type hydraIntrospectionAuthorizer struct {
	introspectionURL string
	clientID         string
	clientSecret     string
	expectedClientID string
	requiredScope    string
	httpClient       *http.Client
}

type hydraIntrospectionResponse struct {
	Active   bool   `json:"active"`
	ClientID string `json:"client_id"`
	Scope    string `json:"scope"`
}

func newHydraIntrospectionAuthorizer(
	introspectionURL string,
	clientID string,
	clientSecret string,
	expectedClientID string,
	requiredScope string,
	httpClient *http.Client,
) (*hydraIntrospectionAuthorizer, error) {
	normalizedURL := strings.TrimSpace(introspectionURL)
	if normalizedURL == "" {
		return nil, fmt.Errorf("trust runtime hydra introspection url must not be empty")
	}
	if _, err := url.Parse(normalizedURL); err != nil {
		return nil, fmt.Errorf("parse trust runtime hydra introspection url: %w", err)
	}
	if strings.TrimSpace(clientID) == "" {
		return nil, fmt.Errorf("trust runtime hydra introspection client id must not be empty")
	}
	if strings.TrimSpace(clientSecret) == "" {
		return nil, fmt.Errorf("trust runtime hydra introspection client secret must not be empty")
	}
	if strings.TrimSpace(expectedClientID) == "" {
		return nil, fmt.Errorf("trust runtime hydra expected client id must not be empty")
	}
	if strings.TrimSpace(requiredScope) == "" {
		return nil, fmt.Errorf("trust runtime hydra required scope must not be empty")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}

	return &hydraIntrospectionAuthorizer{
		introspectionURL: normalizedURL,
		clientID:         strings.TrimSpace(clientID),
		clientSecret:     strings.TrimSpace(clientSecret),
		expectedClientID: strings.TrimSpace(expectedClientID),
		requiredScope:    strings.TrimSpace(requiredScope),
		httpClient:       httpClient,
	}, nil
}

func (a *hydraIntrospectionAuthorizer) Authorize(ctx context.Context, bearerToken string) (internalPrincipal, error) {
	trimmedToken := strings.TrimSpace(bearerToken)
	if trimmedToken == "" {
		return internalPrincipal{}, ErrInternalAuthUnauthenticated
	}

	form := url.Values{}
	form.Set("token", trimmedToken)
	form.Set("token_type_hint", "access_token")

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		a.introspectionURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return internalPrincipal{}, fmt.Errorf("build hydra introspection request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(a.clientID, a.clientSecret)

	response, err := a.httpClient.Do(request)
	if err != nil {
		return internalPrincipal{}, fmt.Errorf("%w: execute hydra introspection request: %v", ErrInternalAuthUnavailable, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return internalPrincipal{}, fmt.Errorf(
			"%w: hydra introspection endpoint returned %d: %s",
			ErrInternalAuthUnavailable,
			response.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	var payload hydraIntrospectionResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return internalPrincipal{}, fmt.Errorf("%w: decode hydra introspection response: %v", ErrInternalAuthUnavailable, err)
	}

	if !payload.Active {
		return internalPrincipal{}, ErrInternalAuthUnauthenticated
	}
	if strings.TrimSpace(payload.ClientID) != a.expectedClientID {
		return internalPrincipal{}, ErrInternalAuthForbidden
	}
	if !containsScope(payload.Scope, a.requiredScope) {
		return internalPrincipal{}, ErrInternalAuthForbidden
	}

	return internalPrincipal{ClientID: payload.ClientID}, nil
}

func containsScope(scopeSet string, expectedScope string) bool {
	for _, scope := range strings.Fields(strings.TrimSpace(scopeSet)) {
		if scope == expectedScope {
			return true
		}
	}

	return false
}
