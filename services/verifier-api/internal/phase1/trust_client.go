package phase1

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

var ErrTrustRuntimeUnauthorized = errors.New("trust runtime request unauthorized")

type AccessTokenSource interface {
	Token(ctx context.Context) (string, error)
}

type TrustReadClient struct {
	tokenSource AccessTokenSource
	baseURL     string
	httpClient  *http.Client
}

type trustSnapshotPayload struct {
	IssuerID                  string   `json:"issuerId"`
	TrustState                string   `json:"trustState"`
	AllowedTemplateIDs        []string `json:"allowedTemplateIds"`
	VerificationKeyReferences []string `json:"verificationKeyReferences"`
}

func NewTrustReadClient(baseURL string, tokenSource AccessTokenSource, httpClient *http.Client) (*TrustReadClient, error) {
	normalizedBaseURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if normalizedBaseURL == "" {
		return nil, fmt.Errorf("trust registry base url must not be empty")
	}
	if tokenSource == nil {
		return nil, fmt.Errorf("trust runtime access token source is required")
	}

	if _, err := url.Parse(normalizedBaseURL); err != nil {
		return nil, fmt.Errorf("parse trust registry base url: %w", err)
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}

	return &TrustReadClient{
		tokenSource: tokenSource,
		baseURL:     normalizedBaseURL,
		httpClient:  httpClient,
	}, nil
}

func (c *TrustReadClient) GetIssuerTrustRecord(ctx context.Context, issuerID string) (IssuerTrustRecord, error) {
	bearerToken, err := c.tokenSource.Token(ctx)
	if err != nil {
		return IssuerTrustRecord{}, fmt.Errorf("acquire trust runtime access token: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"/internal/v1/phase1/issuers/"+url.PathEscape(strings.TrimSpace(issuerID))+"/trust",
		nil,
	)
	if err != nil {
		return IssuerTrustRecord{}, fmt.Errorf("build trust registry request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(bearerToken))

	response, err := c.httpClient.Do(request)
	if err != nil {
		return IssuerTrustRecord{}, fmt.Errorf("execute trust registry request: %w", err)
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return IssuerTrustRecord{}, ErrRecordNotFound
	case http.StatusUnauthorized, http.StatusForbidden:
		return IssuerTrustRecord{}, ErrTrustRuntimeUnauthorized
	default:
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return IssuerTrustRecord{}, fmt.Errorf("trust registry returned %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload trustSnapshotPayload
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return IssuerTrustRecord{}, fmt.Errorf("decode trust registry response: %w", err)
	}

	return IssuerTrustRecord{
		IssuerID:                  payload.IssuerID,
		TrustState:                payload.TrustState,
		AllowedTemplateIDs:        append([]string(nil), payload.AllowedTemplateIDs...),
		VerificationKeyReferences: append([]string(nil), payload.VerificationKeyReferences...),
	}, nil
}
