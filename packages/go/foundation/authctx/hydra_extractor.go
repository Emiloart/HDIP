package authctx

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
	ErrUnauthenticated = errors.New("request unauthenticated")
	ErrAuthUnavailable = errors.New("auth provider unavailable")
)

type HydraIntrospectionConfig struct {
	IntrospectionURL string
	ClientID         string
	ClientSecret     string
	HTTPClient       *http.Client
}

type HydraIssuerOperatorExtractor struct {
	introspector *HydraIntrospectionExtractor
}

type HydraVerifierIntegratorExtractor struct {
	introspector *HydraIntrospectionExtractor
}

type HydraIntrospectionExtractor struct {
	introspectionURL string
	clientID         string
	clientSecret     string
	httpClient       *http.Client
}

type hydraIntrospectionResponse struct {
	Active   bool   `json:"active"`
	ClientID string `json:"client_id"`
	Scope    string `json:"scope"`
	TokenID  string `json:"jti"`
}

func NewHydraIssuerOperatorExtractor(cfg HydraIntrospectionConfig) (*HydraIssuerOperatorExtractor, error) {
	introspector, err := NewHydraIntrospectionExtractor(cfg)
	if err != nil {
		return nil, err
	}

	return &HydraIssuerOperatorExtractor{introspector: introspector}, nil
}

func NewHydraVerifierIntegratorExtractor(cfg HydraIntrospectionConfig) (*HydraVerifierIntegratorExtractor, error) {
	introspector, err := NewHydraIntrospectionExtractor(cfg)
	if err != nil {
		return nil, err
	}

	return &HydraVerifierIntegratorExtractor{introspector: introspector}, nil
}

func NewHydraIntrospectionExtractor(cfg HydraIntrospectionConfig) (*HydraIntrospectionExtractor, error) {
	normalizedURL := strings.TrimSpace(cfg.IntrospectionURL)
	if normalizedURL == "" {
		return nil, fmt.Errorf("hydra introspection url must not be empty")
	}
	parsedURL, err := url.Parse(normalizedURL)
	if err != nil {
		return nil, fmt.Errorf("parse hydra introspection url: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("hydra introspection url must use http or https")
	}
	if parsedURL.Host == "" {
		return nil, fmt.Errorf("hydra introspection url must include a host")
	}
	if strings.TrimSpace(cfg.ClientID) == "" {
		return nil, fmt.Errorf("hydra introspection client id must not be empty")
	}
	if strings.TrimSpace(cfg.ClientSecret) == "" {
		return nil, fmt.Errorf("hydra introspection client secret must not be empty")
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}

	return &HydraIntrospectionExtractor{
		introspectionURL: normalizedURL,
		clientID:         strings.TrimSpace(cfg.ClientID),
		clientSecret:     strings.TrimSpace(cfg.ClientSecret),
		httpClient:       httpClient,
	}, nil
}

func (e *HydraIssuerOperatorExtractor) IssuerOperatorFromRequest(r *http.Request) (Attribution, error) {
	attribution, err := e.introspector.attributionFromRequest(r, ActorTypeIssuerOperator)
	if err != nil {
		return Attribution{}, err
	}
	if err := attribution.ValidateFor(ActorTypeIssuerOperator); err != nil {
		return Attribution{}, err
	}

	return attribution, nil
}

func (e *HydraIssuerOperatorExtractor) Check(ctx context.Context) error {
	return e.introspector.Check(ctx)
}

func (e *HydraVerifierIntegratorExtractor) VerifierIntegratorFromRequest(r *http.Request) (Attribution, error) {
	attribution, err := e.introspector.attributionFromRequest(r, ActorTypeVerifierIntegrator)
	if err != nil {
		return Attribution{}, err
	}
	if err := attribution.ValidateFor(ActorTypeVerifierIntegrator); err != nil {
		return Attribution{}, err
	}

	return attribution, nil
}

func (e *HydraVerifierIntegratorExtractor) Check(ctx context.Context) error {
	return e.introspector.Check(ctx)
}

func (e *HydraIntrospectionExtractor) Check(ctx context.Context) error {
	_, err := e.introspectToken(ctx, "phase1-public-auth-readiness-probe")
	return err
}

func (e *HydraIntrospectionExtractor) attributionFromRequest(r *http.Request, actorType ActorType) (Attribution, error) {
	token, err := bearerTokenFromRequest(r)
	if err != nil {
		return Attribution{}, err
	}

	payload, err := e.introspectToken(r.Context(), token)
	if err != nil {
		return Attribution{}, err
	}
	if !payload.Active {
		return Attribution{}, ErrUnauthenticated
	}

	clientID := strings.TrimSpace(payload.ClientID)
	if clientID == "" {
		return Attribution{}, ErrUnauthenticated
	}

	return Attribution{
		PrincipalID:             clientID,
		OrganizationID:          clientID,
		ActorType:               actorType,
		Scopes:                  strings.Fields(strings.TrimSpace(payload.Scope)),
		AuthenticationReference: authenticationReference(payload),
	}, nil
}

func (e *HydraIntrospectionExtractor) introspectToken(ctx context.Context, token string) (hydraIntrospectionResponse, error) {
	form := url.Values{}
	form.Set("token", strings.TrimSpace(token))
	form.Set("token_type_hint", "access_token")

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		e.introspectionURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return hydraIntrospectionResponse{}, fmt.Errorf("build hydra introspection request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(e.clientID, e.clientSecret)

	response, err := e.httpClient.Do(request)
	if err != nil {
		return hydraIntrospectionResponse{}, fmt.Errorf("%w: execute hydra introspection request: %v", ErrAuthUnavailable, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return hydraIntrospectionResponse{}, fmt.Errorf(
			"%w: hydra introspection endpoint returned %d: %s",
			ErrAuthUnavailable,
			response.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	var payload hydraIntrospectionResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return hydraIntrospectionResponse{}, fmt.Errorf("%w: decode hydra introspection response: %v", ErrAuthUnavailable, err)
	}

	return payload, nil
}

func bearerTokenFromRequest(r *http.Request) (string, error) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return "", ErrUnauthenticated
	}

	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", ErrUnauthenticated
	}

	return strings.TrimSpace(parts[1]), nil
}

func authenticationReference(payload hydraIntrospectionResponse) string {
	if tokenID := strings.TrimSpace(payload.TokenID); tokenID != "" {
		return "hydra-token:" + tokenID
	}

	return "hydra-client:" + strings.TrimSpace(payload.ClientID)
}
