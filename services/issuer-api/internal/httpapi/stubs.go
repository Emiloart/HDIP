package httpapi

import (
	"fmt"

	"github.com/Emiloart/HDIP/services/issuer-api/internal/config"
)

const defaultTemplateID = "hdip-passport-basic"

type issuerProfileResponse struct {
	IssuerID                     string   `json:"issuerId"`
	DisplayName                  string   `json:"displayName"`
	Endpoint                     string   `json:"endpoint"`
	SupportedCredentialTemplates []string `json:"supportedCredentialTemplates"`
}

type credentialTemplateMetadataResponse struct {
	TemplateID      string   `json:"templateId"`
	DisplayName     string   `json:"displayName"`
	Version         string   `json:"version"`
	CredentialTypes []string `json:"credentialTypes"`
}

func stubIssuerProfile(cfg config.Config) issuerProfileResponse {
	return issuerProfileResponse{
		IssuerID:                     "did:web:issuer.hdip.dev",
		DisplayName:                  "HDIP Passport Issuer",
		Endpoint:                     fmt.Sprintf("http://%s", cfg.Address()),
		SupportedCredentialTemplates: []string{defaultTemplateID},
	}
}

func stubCredentialTemplate(templateID string) (credentialTemplateMetadataResponse, bool) {
	if templateID != defaultTemplateID {
		return credentialTemplateMetadataResponse{}, false
	}

	return credentialTemplateMetadataResponse{
		TemplateID:      defaultTemplateID,
		DisplayName:     "HDIP Passport Basic",
		Version:         "2026.04",
		CredentialTypes: []string{"HDIPPassportCredential", "KycCredential"},
	}, true
}
