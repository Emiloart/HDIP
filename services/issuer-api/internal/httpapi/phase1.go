package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Emiloart/HDIP/packages/go/foundation/authctx"
	"github.com/Emiloart/HDIP/packages/go/foundation/httpx"
	phase1 "github.com/Emiloart/HDIP/services/issuer-api/internal/phase1"
)

const (
	issuerIssueScope        = "issuer.credentials.issue"
	issuerReadScope         = "issuer.credentials.read"
	placeholderCredentialID = "cred_hdip_passport_basic_001"
	placeholderArtifactHash = "3a3d9fbf43cf1769b1485596ddc44f4d2d9df7f47f5f70b9771af35fb0dcb2ef"
)

var (
	placeholderIssuedAt = time.Date(2026, time.April, 20, 9, 0, 0, 0, time.UTC)
)

type phase1IssuerHandler struct {
	issuerExtractor authctx.IssuerOperatorExtractor
	issuers         phase1.IssuerRecordRepository
	credentials     phase1.CredentialRecordRepository
	audits          phase1.AuditRecordRepository
}

type issuanceRequestPayload struct {
	TemplateID       string                   `json:"templateId"`
	SubjectReference string                   `json:"subjectReference"`
	Claims           issuanceKYCClaimsPayload `json:"claims"`
}

type issuanceKYCClaimsPayload struct {
	FullLegalName      string    `json:"fullLegalName"`
	DateOfBirth        string    `json:"dateOfBirth"`
	CountryOfResidence string    `json:"countryOfResidence"`
	DocumentCountry    string    `json:"documentCountry"`
	KYCLevel           string    `json:"kycLevel"`
	VerifiedAt         time.Time `json:"verifiedAt"`
	ExpiresAt          time.Time `json:"expiresAt"`
}

type signedCredentialPayload struct {
	Format    string `json:"format"`
	MediaType string `json:"mediaType"`
	Value     string `json:"value"`
}

type issuanceResponsePayload struct {
	CredentialID     string                  `json:"credentialId"`
	IssuerID         string                  `json:"issuerId"`
	TemplateID       string                  `json:"templateId"`
	Status           string                  `json:"status"`
	IssuedAt         time.Time               `json:"issuedAt"`
	ExpiresAt        time.Time               `json:"expiresAt"`
	StatusReference  string                  `json:"statusReference"`
	SignedCredential signedCredentialPayload `json:"signedCredential"`
}

type credentialRecordPayload struct {
	CredentialID             string                   `json:"credentialId"`
	IssuerID                 string                   `json:"issuerId"`
	TemplateID               string                   `json:"templateId"`
	SubjectReference         string                   `json:"subjectReference"`
	Claims                   issuanceKYCClaimsPayload `json:"claims"`
	ArtifactDigest           string                   `json:"artifactDigest"`
	Status                   string                   `json:"status"`
	StatusReference          string                   `json:"statusReference"`
	IssuedAt                 time.Time                `json:"issuedAt"`
	ExpiresAt                time.Time                `json:"expiresAt"`
	StatusUpdatedAt          time.Time                `json:"statusUpdatedAt"`
	SupersededByCredentialID string                   `json:"supersededByCredentialId,omitempty"`
	SignedCredential         *signedCredentialPayload `json:"signedCredential,omitempty"`
	ArtifactReference        string                   `json:"artifactReference,omitempty"`
}

func newPhase1IssuerHandler() *phase1IssuerHandler {
	store := phase1.NewInMemoryStore()
	return &phase1IssuerHandler{
		issuerExtractor: authctx.HeaderIssuerOperatorExtractor{},
		issuers:         store,
		credentials:     store,
		audits:          store,
	}
}

func (h *phase1IssuerHandler) issueCredential(w http.ResponseWriter, r *http.Request) {
	attribution, ok := h.requireIssuerAttribution(w, r, issuerIssueScope)
	if !ok {
		return
	}

	var request issuanceRequestPayload
	if err := httpx.DecodeJSONBody(r, &request); err != nil {
		httpx.WriteError(w, r.Context(), http.StatusBadRequest, "invalid_request", "request body must match the Phase 1 issuance contract")
		return
	}

	if err := request.validate(); err != nil {
		httpx.WriteError(w, r.Context(), http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	issuerRecord, err := h.issuers.GetIssuerRecord(r.Context(), attribution.OrganizationID)
	if err != nil {
		httpx.WriteError(w, r.Context(), http.StatusForbidden, "issuer_not_trusted", "issuer trust context is not available")
		return
	}

	if !contains(issuerRecord.AllowedTemplateIDs, request.TemplateID) {
		httpx.WriteError(w, r.Context(), http.StatusBadRequest, "unsupported_template", "credential template is not allowed for issuer")
		return
	}

	record := phase1.CredentialRecord{
		CredentialID:     placeholderCredentialID,
		IssuerID:         issuerRecord.IssuerID,
		TemplateID:       request.TemplateID,
		SubjectReference: request.SubjectReference,
		Claims: phase1.KYCClaims{
			FullLegalName:      request.Claims.FullLegalName,
			DateOfBirth:        request.Claims.DateOfBirth,
			CountryOfResidence: request.Claims.CountryOfResidence,
			DocumentCountry:    request.Claims.DocumentCountry,
			KYCLevel:           request.Claims.KYCLevel,
			VerifiedAt:         request.Claims.VerifiedAt,
			ExpiresAt:          request.Claims.ExpiresAt,
		},
		ArtifactDigest: placeholderArtifactHash,
		SignedCredential: &phase1.SignedCredentialArtifact{
			Format:    "sd_jwt_vc",
			MediaType: "application/vc+sd-jwt",
			Value:     "eyJhbGciOiJFUzI1NiJ9.placeholder.credential",
		},
		Status:          phase1.CredentialStatusActive,
		StatusReference: statusReferenceForCredential(placeholderCredentialID),
		IssuedAt:        placeholderIssuedAt,
		ExpiresAt:       request.Claims.ExpiresAt,
		StatusUpdatedAt: placeholderIssuedAt,
	}

	if err := h.credentials.CreateCredentialRecord(r.Context(), record); err != nil {
		httpx.WriteError(w, r.Context(), http.StatusInternalServerError, "persistence_error", "credential record could not be stored")
		return
	}

	_ = h.audits.AppendAuditRecord(r.Context(), phase1.AuditRecord{
		AuditID:        auditIDForAction(r, "issue"),
		Actor:          attribution,
		Action:         issuerIssueScope,
		ResourceType:   "credential",
		ResourceID:     record.CredentialID,
		RequestID:      httpx.RequestIDFromContext(r.Context()),
		IdempotencyKey: strings.TrimSpace(r.Header.Get("Idempotency-Key")),
		Outcome:        "succeeded",
		OccurredAt:     placeholderIssuedAt,
		ServiceName:    "issuer-api",
	})

	w.Header().Set("Location", "/v1/issuer/credentials/"+record.CredentialID)
	httpx.WriteJSON(w, http.StatusCreated, issuanceResponsePayload{
		CredentialID:    record.CredentialID,
		IssuerID:        record.IssuerID,
		TemplateID:      record.TemplateID,
		Status:          string(record.Status),
		IssuedAt:        record.IssuedAt,
		ExpiresAt:       record.ExpiresAt,
		StatusReference: record.StatusReference,
		SignedCredential: signedCredentialPayload{
			Format:    record.SignedCredential.Format,
			MediaType: record.SignedCredential.MediaType,
			Value:     record.SignedCredential.Value,
		},
	})
}

func (h *phase1IssuerHandler) getCredential(w http.ResponseWriter, r *http.Request) {
	attribution, ok := h.requireIssuerAttribution(w, r, issuerReadScope)
	if !ok {
		return
	}

	record, err := h.credentials.GetCredentialRecord(r.Context(), r.PathValue("credentialId"))
	if err != nil {
		if errors.Is(err, phase1.ErrRecordNotFound) {
			httpx.WriteError(w, r.Context(), http.StatusNotFound, "credential_not_found", "credential record not found")
			return
		}

		httpx.WriteError(w, r.Context(), http.StatusInternalServerError, "persistence_error", "credential record could not be loaded")
		return
	}

	_ = h.audits.AppendAuditRecord(r.Context(), phase1.AuditRecord{
		AuditID:        auditIDForAction(r, "read"),
		Actor:          attribution,
		Action:         issuerReadScope,
		ResourceType:   "credential",
		ResourceID:     record.CredentialID,
		RequestID:      httpx.RequestIDFromContext(r.Context()),
		IdempotencyKey: strings.TrimSpace(r.Header.Get("Idempotency-Key")),
		Outcome:        "succeeded",
		OccurredAt:     placeholderIssuedAt,
		ServiceName:    "issuer-api",
	})

	httpx.WriteJSON(w, http.StatusOK, credentialRecordPayload{
		CredentialID:             record.CredentialID,
		IssuerID:                 record.IssuerID,
		TemplateID:               record.TemplateID,
		SubjectReference:         record.SubjectReference,
		Claims:                   claimsPayloadFromRecord(record.Claims),
		ArtifactDigest:           record.ArtifactDigest,
		Status:                   string(record.Status),
		StatusReference:          record.StatusReference,
		IssuedAt:                 record.IssuedAt,
		ExpiresAt:                record.ExpiresAt,
		StatusUpdatedAt:          record.StatusUpdatedAt,
		SupersededByCredentialID: record.SupersededByCredentialID,
		SignedCredential:         signedCredentialPayloadFromRecord(record.SignedCredential),
		ArtifactReference:        record.ArtifactReference,
	})
}

func (h *phase1IssuerHandler) requireIssuerAttribution(w http.ResponseWriter, r *http.Request, scope string) (authctx.Attribution, bool) {
	attribution, err := h.issuerExtractor.IssuerOperatorFromRequest(r)
	if err != nil {
		httpx.WriteError(w, r.Context(), http.StatusUnauthorized, "unauthenticated", "authenticated issuer attribution is required")
		return authctx.Attribution{}, false
	}

	if err := authctx.RequireScope(attribution, scope); err != nil {
		httpx.WriteError(w, r.Context(), http.StatusForbidden, "insufficient_scope", "issuer scope is required for this action")
		return authctx.Attribution{}, false
	}

	return attribution, true
}

func (p issuanceRequestPayload) validate() error {
	switch {
	case strings.TrimSpace(p.TemplateID) == "":
		return errors.New("templateId must not be empty")
	case strings.TrimSpace(p.SubjectReference) == "":
		return errors.New("subjectReference must not be empty")
	default:
		return p.Claims.validate()
	}
}

func (p issuanceKYCClaimsPayload) validate() error {
	switch {
	case strings.TrimSpace(p.FullLegalName) == "":
		return errors.New("claims.fullLegalName must not be empty")
	case strings.TrimSpace(p.DateOfBirth) == "":
		return errors.New("claims.dateOfBirth must not be empty")
	case strings.TrimSpace(p.CountryOfResidence) == "":
		return errors.New("claims.countryOfResidence must not be empty")
	case strings.TrimSpace(p.DocumentCountry) == "":
		return errors.New("claims.documentCountry must not be empty")
	case strings.TrimSpace(p.KYCLevel) == "":
		return errors.New("claims.kycLevel must not be empty")
	case p.VerifiedAt.IsZero():
		return errors.New("claims.verifiedAt must not be empty")
	case p.ExpiresAt.IsZero():
		return errors.New("claims.expiresAt must not be empty")
	default:
		return nil
	}
}

func statusReferenceForCredential(credentialID string) string {
	return "status:" + credentialID
}

func auditIDForAction(r *http.Request, action string) string {
	return httpx.RequestIDFromContext(r.Context()) + ":" + action
}

func claimsPayloadFromRecord(record phase1.KYCClaims) issuanceKYCClaimsPayload {
	return issuanceKYCClaimsPayload{
		FullLegalName:      record.FullLegalName,
		DateOfBirth:        record.DateOfBirth,
		CountryOfResidence: record.CountryOfResidence,
		DocumentCountry:    record.DocumentCountry,
		KYCLevel:           record.KYCLevel,
		VerifiedAt:         record.VerifiedAt,
		ExpiresAt:          record.ExpiresAt,
	}
}

func signedCredentialPayloadFromRecord(record *phase1.SignedCredentialArtifact) *signedCredentialPayload {
	if record == nil {
		return nil
	}

	return &signedCredentialPayload{
		Format:    record.Format,
		MediaType: record.MediaType,
		Value:     record.Value,
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}

	return false
}
