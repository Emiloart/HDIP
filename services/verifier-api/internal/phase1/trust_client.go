package phase1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TrustReadClient struct {
	baseURL    string
	httpClient *http.Client
}

type trustSnapshotPayload struct {
	IssuerID                  string   `json:"issuerId"`
	TrustState                string   `json:"trustState"`
	AllowedTemplateIDs        []string `json:"allowedTemplateIds"`
	VerificationKeyReferences []string `json:"verificationKeyReferences"`
}

func NewTrustReadClient(baseURL string, httpClient *http.Client) (*TrustReadClient, error) {
	normalizedBaseURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if normalizedBaseURL == "" {
		return nil, fmt.Errorf("trust registry base url must not be empty")
	}

	if _, err := url.Parse(normalizedBaseURL); err != nil {
		return nil, fmt.Errorf("parse trust registry base url: %w", err)
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}

	return &TrustReadClient{
		baseURL:    normalizedBaseURL,
		httpClient: httpClient,
	}, nil
}

func (c *TrustReadClient) GetIssuerTrustRecord(ctx context.Context, issuerID string) (IssuerTrustRecord, error) {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"/internal/v1/phase1/issuers/"+url.PathEscape(strings.TrimSpace(issuerID))+"/trust",
		nil,
	)
	if err != nil {
		return IssuerTrustRecord{}, fmt.Errorf("build trust registry request: %w", err)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return IssuerTrustRecord{}, fmt.Errorf("execute trust registry request: %w", err)
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return IssuerTrustRecord{}, ErrRecordNotFound
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
