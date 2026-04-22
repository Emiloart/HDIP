package phase1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type HydraClientCredentialsTokenSource struct {
	tokenURL     string
	clientID     string
	clientSecret string
	scope        string
	httpClient   *http.Client

	mu           sync.Mutex
	cachedToken  string
	cachedExpiry time.Time
}

type hydraTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

func NewHydraClientCredentialsTokenSource(
	tokenURL string,
	clientID string,
	clientSecret string,
	scope string,
	httpClient *http.Client,
) (*HydraClientCredentialsTokenSource, error) {
	normalizedTokenURL := strings.TrimSpace(tokenURL)
	if normalizedTokenURL == "" {
		return nil, fmt.Errorf("trust runtime hydra token url must not be empty")
	}
	if _, err := url.Parse(normalizedTokenURL); err != nil {
		return nil, fmt.Errorf("parse trust runtime hydra token url: %w", err)
	}
	if strings.TrimSpace(clientID) == "" {
		return nil, fmt.Errorf("trust runtime hydra client id must not be empty")
	}
	if strings.TrimSpace(clientSecret) == "" {
		return nil, fmt.Errorf("trust runtime hydra client secret must not be empty")
	}
	if strings.TrimSpace(scope) == "" {
		return nil, fmt.Errorf("trust runtime hydra scope must not be empty")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}

	return &HydraClientCredentialsTokenSource{
		tokenURL:     normalizedTokenURL,
		clientID:     strings.TrimSpace(clientID),
		clientSecret: strings.TrimSpace(clientSecret),
		scope:        strings.TrimSpace(scope),
		httpClient:   httpClient,
	}, nil
}

func (s *HydraClientCredentialsTokenSource) Token(ctx context.Context) (string, error) {
	if token := s.cachedAccessToken(); token != "" {
		return token, nil
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("scope", s.scope)

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		s.tokenURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("build hydra token request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(s.clientID, s.clientSecret)

	response, err := s.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("execute hydra token request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return "", fmt.Errorf("hydra token endpoint returned %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload hydraTokenResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode hydra token response: %w", err)
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return "", fmt.Errorf("hydra token response missing access_token")
	}
	if tokenType := strings.TrimSpace(payload.TokenType); tokenType != "" && !strings.EqualFold(tokenType, "bearer") {
		return "", fmt.Errorf("hydra token response used unsupported token_type %q", tokenType)
	}

	s.cacheAccessToken(payload.AccessToken, payload.ExpiresIn)
	return strings.TrimSpace(payload.AccessToken), nil
}

func (s *HydraClientCredentialsTokenSource) cachedAccessToken() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if strings.TrimSpace(s.cachedToken) == "" {
		return ""
	}
	if !s.cachedExpiry.IsZero() && time.Now().UTC().After(s.cachedExpiry) {
		s.cachedToken = ""
		s.cachedExpiry = time.Time{}
		return ""
	}

	return s.cachedToken
}

func (s *HydraClientCredentialsTokenSource) cacheAccessToken(token string, expiresInSeconds int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cachedToken = strings.TrimSpace(token)

	now := time.Now().UTC()
	switch {
	case expiresInSeconds <= 0:
		s.cachedExpiry = now
	case expiresInSeconds <= 30:
		s.cachedExpiry = now.Add(time.Duration(expiresInSeconds) * time.Second)
	default:
		s.cachedExpiry = now.Add(time.Duration(expiresInSeconds-30) * time.Second)
	}
}
