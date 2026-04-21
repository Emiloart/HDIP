package httpapi

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Emiloart/HDIP/packages/go/foundation/authctx"
	"github.com/Emiloart/HDIP/packages/go/foundation/httpx"
	phase1 "github.com/Emiloart/HDIP/services/issuer-api/internal/phase1"
)

const (
	issuerIssueScope            = "issuer.credentials.issue"
	issuerReadScope             = "issuer.credentials.read"
	issuerStatusWriteScope      = "issuer.credentials.status.write"
	defaultPhase1IssuerID       = "did:web:issuer.hdip.dev"
	credentialArtifactKind      = "phase1_opaque_artifact"
	credentialArtifactMediaType = "application/vnd.hdip.phase1-opaque-artifact"
	credentialArtifactPrefix    = "opaque-artifact:v1:"
)

var isoCountryCodePattern = regexp.MustCompile(`^[A-Z]{2}$`)

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

type credentialArtifactPayload struct {
	Kind      string `json:"kind"`
	MediaType string `json:"mediaType"`
	Value     string `json:"value"`
}

type issuanceResponsePayload struct {
	CredentialID       string                    `json:"credentialId"`
	IssuerID           string                    `json:"issuerId"`
	TemplateID         string                    `json:"templateId"`
	Status             string                    `json:"status"`
	IssuedAt           time.Time                 `json:"issuedAt"`
	ExpiresAt          time.Time                 `json:"expiresAt"`
	StatusReference    string                    `json:"statusReference"`
	CredentialArtifact credentialArtifactPayload `json:"credentialArtifact"`
}

type credentialRecordPayload struct {
	CredentialID             string                     `json:"credentialId"`
	IssuerID                 string                     `json:"issuerId"`
	TemplateID               string                     `json:"templateId"`
	SubjectReference         string                     `json:"subjectReference"`
	Claims                   issuanceKYCClaimsPayload   `json:"claims"`
	ArtifactDigest           string                     `json:"artifactDigest"`
	Status                   string                     `json:"status"`
	StatusReference          string                     `json:"statusReference"`
	IssuedAt                 time.Time                  `json:"issuedAt"`
	ExpiresAt                time.Time                  `json:"expiresAt"`
	StatusUpdatedAt          time.Time                  `json:"statusUpdatedAt"`
	SupersededByCredentialID string                     `json:"supersededByCredentialId,omitempty"`
	CredentialArtifact       *credentialArtifactPayload `json:"credentialArtifact,omitempty"`
	ArtifactReference        string                     `json:"artifactReference,omitempty"`
}

type credentialStatusUpdateRequestPayload struct {
	Status                   string `json:"status"`
	SupersededByCredentialID string `json:"supersededByCredentialId,omitempty"`
}

type credentialStatusPayload struct {
	CredentialID             string    `json:"credentialId"`
	Status                   string    `json:"status"`
	StatusReference          string    `json:"statusReference"`
	StatusUpdatedAt          time.Time `json:"statusUpdatedAt"`
	ExpiresAt                time.Time `json:"expiresAt"`
	SupersededByCredentialID string    `json:"supersededByCredentialId,omitempty"`
}

func newPhase1IssuerHandler(runtimePath string) (*phase1IssuerHandler, error) {
	store, err := phase1.OpenRuntimeStore(runtimePath)
	if err != nil {
		return nil, err
	}

	if err := store.SeedIssuerRecord(defaultIssuerRecord()); err != nil {
		return nil, err
	}

	return newPhase1IssuerHandlerWithStore(store), nil
}

func newPhase1IssuerHandlerWithStore(store *phase1.RuntimeStore) *phase1IssuerHandler {
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

	credentialID, err := h.credentials.NextCredentialID(r.Context(), request.TemplateID)
	if err != nil {
		httpx.WriteError(w, r.Context(), http.StatusInternalServerError, "persistence_error", "credential identifier could not be allocated")
		return
	}

	issuedAt := request.Claims.VerifiedAt.UTC()
	expiresAt := request.Claims.ExpiresAt.UTC()
	credentialArtifact, err := materializeCredentialArtifact(credentialID, issuerRecord.IssuerID, request.TemplateID, expiresAt)
	if err != nil {
		httpx.WriteError(w, r.Context(), http.StatusInternalServerError, "artifact_materialization_failed", "credential artifact could not be materialized")
		return
	}

	record := phase1.CredentialRecord{
		CredentialID:     credentialID,
		IssuerID:         issuerRecord.IssuerID,
		TemplateID:       request.TemplateID,
		SubjectReference: request.SubjectReference,
		Claims: phase1.KYCClaims{
			FullLegalName:      request.Claims.FullLegalName,
			DateOfBirth:        request.Claims.DateOfBirth,
			CountryOfResidence: request.Claims.CountryOfResidence,
			DocumentCountry:    request.Claims.DocumentCountry,
			KYCLevel:           request.Claims.KYCLevel,
			VerifiedAt:         issuedAt,
			ExpiresAt:          expiresAt,
		},
		ArtifactDigest:     artifactDigest(credentialArtifact.Value),
		CredentialArtifact: &credentialArtifact,
		Status:             phase1.CredentialStatusActive,
		StatusReference:    statusReferenceForCredential(credentialID),
		IssuedAt:           issuedAt,
		ExpiresAt:          expiresAt,
		StatusUpdatedAt:    issuedAt,
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
		OccurredAt:     issuedAt,
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
		CredentialArtifact: credentialArtifactPayload{
			Kind:      record.CredentialArtifact.Kind,
			MediaType: record.CredentialArtifact.MediaType,
			Value:     record.CredentialArtifact.Value,
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

	if record.IssuerID != attribution.OrganizationID {
		httpx.WriteError(w, r.Context(), http.StatusNotFound, "credential_not_found", "credential record not found")
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
		OccurredAt:     time.Now().UTC(),
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
		CredentialArtifact:       credentialArtifactPayloadFromRecord(record.CredentialArtifact),
		ArtifactReference:        record.ArtifactReference,
	})
}

func (h *phase1IssuerHandler) updateCredentialStatus(w http.ResponseWriter, r *http.Request) {
	attribution, ok := h.requireIssuerAttribution(w, r, issuerStatusWriteScope)
	if !ok {
		return
	}

	var request credentialStatusUpdateRequestPayload
	if err := httpx.DecodeJSONBody(r, &request); err != nil {
		httpx.WriteError(w, r.Context(), http.StatusBadRequest, "invalid_request", "request body must match the Phase 1 credential status contract")
		return
	}

	if err := request.validate(); err != nil {
		httpx.WriteError(w, r.Context(), http.StatusBadRequest, "invalid_request", err.Error())
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

	if record.IssuerID != attribution.OrganizationID {
		httpx.WriteError(w, r.Context(), http.StatusNotFound, "credential_not_found", "credential record not found")
		return
	}

	if record.Status != phase1.CredentialStatusActive {
		httpx.WriteError(w, r.Context(), http.StatusConflict, "invalid_status_transition", "credential status transition is not allowed from the current status")
		return
	}

	nextStatus := phase1.CredentialStatus(request.Status)
	statusUpdatedAt := time.Now().UTC()
	if err := h.credentials.UpdateCredentialStatus(
		r.Context(),
		record.CredentialID,
		nextStatus,
		statusUpdatedAt,
		strings.TrimSpace(request.SupersededByCredentialID),
	); err != nil {
		if errors.Is(err, phase1.ErrRecordNotFound) {
			httpx.WriteError(w, r.Context(), http.StatusNotFound, "credential_not_found", "credential record not found")
			return
		}

		httpx.WriteError(w, r.Context(), http.StatusInternalServerError, "persistence_error", "credential status could not be updated")
		return
	}

	record.Status = nextStatus
	record.StatusUpdatedAt = statusUpdatedAt
	record.SupersededByCredentialID = strings.TrimSpace(request.SupersededByCredentialID)

	_ = h.audits.AppendAuditRecord(r.Context(), phase1.AuditRecord{
		AuditID:        auditIDForAction(r, "status"),
		Actor:          attribution,
		Action:         issuerStatusWriteScope,
		ResourceType:   "credential",
		ResourceID:     record.CredentialID,
		RequestID:      httpx.RequestIDFromContext(r.Context()),
		IdempotencyKey: strings.TrimSpace(r.Header.Get("Idempotency-Key")),
		Outcome:        "succeeded",
		OccurredAt:     statusUpdatedAt,
		ServiceName:    "issuer-api",
	})

	httpx.WriteJSON(w, http.StatusOK, credentialStatusPayloadFromRecord(record))
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
	case !isISODate(p.DateOfBirth):
		return errors.New("claims.dateOfBirth must be an ISO date")
	case strings.TrimSpace(p.CountryOfResidence) == "":
		return errors.New("claims.countryOfResidence must not be empty")
	case !isoCountryCodePattern.MatchString(p.CountryOfResidence):
		return errors.New("claims.countryOfResidence must be a two-letter ISO country code")
	case strings.TrimSpace(p.DocumentCountry) == "":
		return errors.New("claims.documentCountry must not be empty")
	case !isoCountryCodePattern.MatchString(p.DocumentCountry):
		return errors.New("claims.documentCountry must be a two-letter ISO country code")
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

func (p credentialStatusUpdateRequestPayload) validate() error {
	status := strings.TrimSpace(p.Status)
	supersededByCredentialID := strings.TrimSpace(p.SupersededByCredentialID)

	switch status {
	case string(phase1.CredentialStatusRevoked):
		if supersededByCredentialID != "" {
			return errors.New("supersededByCredentialId must be empty when status is revoked")
		}
	case string(phase1.CredentialStatusSuperseded):
		if supersededByCredentialID == "" {
			return errors.New("supersededByCredentialId must not be empty when status is superseded")
		}
	case string(phase1.CredentialStatusActive):
		return errors.New("status must be a terminal Phase 1 status transition")
	default:
		return errors.New("status must be one of revoked or superseded")
	}

	return nil
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

func credentialArtifactPayloadFromRecord(record *phase1.CredentialArtifact) *credentialArtifactPayload {
	if record == nil {
		return nil
	}

	return &credentialArtifactPayload{
		Kind:      record.Kind,
		MediaType: record.MediaType,
		Value:     record.Value,
	}
}

func credentialStatusPayloadFromRecord(record phase1.CredentialRecord) credentialStatusPayload {
	return credentialStatusPayload{
		CredentialID:             record.CredentialID,
		Status:                   string(record.Status),
		StatusReference:          record.StatusReference,
		StatusUpdatedAt:          record.StatusUpdatedAt,
		ExpiresAt:                record.ExpiresAt,
		SupersededByCredentialID: record.SupersededByCredentialID,
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

type credentialArtifactEnvelope struct {
	CredentialID string `json:"credentialId"`
	IssuerID     string `json:"issuerId"`
	TemplateID   string `json:"templateId"`
	ExpiresAt    string `json:"expiresAt"`
}

func materializeCredentialArtifact(
	credentialID string,
	issuerID string,
	templateID string,
	expiresAt time.Time,
) (phase1.CredentialArtifact, error) {
	envelope := credentialArtifactEnvelope{
		CredentialID: credentialID,
		IssuerID:     issuerID,
		TemplateID:   templateID,
		ExpiresAt:    expiresAt.UTC().Format(time.RFC3339),
	}

	raw, err := json.Marshal(envelope)
	if err != nil {
		return phase1.CredentialArtifact{}, err
	}

	return phase1.CredentialArtifact{
		Kind:      credentialArtifactKind,
		MediaType: credentialArtifactMediaType,
		Value:     credentialArtifactPrefix + base64.RawURLEncoding.EncodeToString(raw),
	}, nil
}

func artifactDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func defaultIssuerRecord() phase1.IssuerRecord {
	now := time.Date(2026, time.April, 20, 9, 0, 0, 0, time.UTC)
	return phase1.IssuerRecord{
		IssuerID:                  defaultPhase1IssuerID,
		DisplayName:               "HDIP Passport Issuer",
		TrustState:                "active",
		AllowedTemplateIDs:        []string{defaultTemplateID},
		VerificationKeyReferences: []string{"key:issuer.hdip.dev:2026-04"},
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}
}

func isISODate(value string) bool {
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}
