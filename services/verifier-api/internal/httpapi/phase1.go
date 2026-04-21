package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Emiloart/HDIP/packages/go/foundation/authctx"
	"github.com/Emiloart/HDIP/packages/go/foundation/httpx"
	phase1 "github.com/Emiloart/HDIP/services/verifier-api/internal/phase1"
)

const (
	verifierCreateScope        = "verifier.requests.create"
	verifierReadScope          = "verifier.results.read"
	placeholderVerificationID  = "verification_hdip_001"
	placeholderVerificationDID = "did:web:issuer.hdip.dev"
	placeholderArtifactHash    = "4262d0aacabb3bd709a5cd7abb52c0eb8be0d15d02f1e8e2e11c45bc5071502e"
)

var placeholderEvaluatedAt = time.Date(2026, time.April, 20, 9, 5, 0, 0, time.UTC)

type phase1VerifierHandler struct {
	verifierExtractor authctx.VerifierIntegratorExtractor
	issuers           phase1.IssuerRecordRepository
	requests          phase1.VerificationRequestRepository
	results           phase1.VerificationResultRepository
	audits            phase1.AuditRecordRepository
}

type verifierCredentialArtifactPayload struct {
	Kind      string `json:"kind"`
	MediaType string `json:"mediaType"`
	Value     string `json:"value"`
}

type verificationSubmissionRequestPayload struct {
	PolicyID           string                            `json:"policyId"`
	CredentialID       string                            `json:"credentialId,omitempty"`
	CredentialArtifact verifierCredentialArtifactPayload `json:"credentialArtifact"`
}

type verificationResultPayload struct {
	VerificationID   string    `json:"verificationId"`
	CredentialID     string    `json:"credentialId,omitempty"`
	IssuerID         string    `json:"issuerId"`
	Decision         string    `json:"decision"`
	ReasonCodes      []string  `json:"reasonCodes"`
	EvaluatedAt      time.Time `json:"evaluatedAt"`
	CredentialStatus string    `json:"credentialStatus"`
}

func newPhase1VerifierHandler() *phase1VerifierHandler {
	store := phase1.NewInMemoryStore()
	return &phase1VerifierHandler{
		verifierExtractor: authctx.HeaderVerifierIntegratorExtractor{},
		issuers:           store,
		requests:          store,
		results:           store,
		audits:            store,
	}
}

func (h *phase1VerifierHandler) createVerification(w http.ResponseWriter, r *http.Request) {
	attribution, ok := h.requireVerifierAttribution(w, r, verifierCreateScope)
	if !ok {
		return
	}

	var request verificationSubmissionRequestPayload
	if err := httpx.DecodeJSONBody(r, &request); err != nil {
		httpx.WriteError(w, r.Context(), http.StatusBadRequest, "invalid_request", "request body must match the Phase 1 verification contract")
		return
	}

	if err := request.validate(); err != nil {
		httpx.WriteError(w, r.Context(), http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	issuerRecord, err := h.issuers.GetIssuerRecord(r.Context(), placeholderVerificationDID)
	if err != nil {
		httpx.WriteError(w, r.Context(), http.StatusForbidden, "issuer_not_trusted", "issuer trust context is not available")
		return
	}

	verificationRequest := phase1.VerificationRequestRecord{
		VerificationID:            placeholderVerificationID,
		VerifierID:                attribution.OrganizationID,
		SubmittedCredentialDigest: placeholderArtifactHash,
		CredentialID:              request.CredentialID,
		PolicyID:                  request.PolicyID,
		RequestedAt:               placeholderEvaluatedAt,
		Actor:                     attribution,
		IdempotencyKey:            strings.TrimSpace(r.Header.Get("Idempotency-Key")),
	}

	if err := h.requests.CreateVerificationRequestRecord(r.Context(), verificationRequest); err != nil {
		httpx.WriteError(w, r.Context(), http.StatusInternalServerError, "persistence_error", "verification request could not be stored")
		return
	}

	verificationResult := phase1.VerificationResultRecord{
		VerificationID:   placeholderVerificationID,
		Decision:         phase1.VerificationDecisionAllow,
		ReasonCodes:      []string{"issuer_trusted", "credential_status_active", "template_match"},
		IssuerTrustState: issuerRecord.TrustState,
		CredentialStatus: phase1.CredentialStatusSnapshotActive,
		EvaluatedAt:      placeholderEvaluatedAt,
		ResponseVersion:  "2026.04",
	}

	if err := h.results.CreateVerificationResultRecord(r.Context(), verificationResult); err != nil {
		httpx.WriteError(w, r.Context(), http.StatusInternalServerError, "persistence_error", "verification result could not be stored")
		return
	}

	_ = h.audits.AppendAuditRecord(r.Context(), phase1.AuditRecord{
		AuditID:        verifierAuditIDForAction(r, "create"),
		Actor:          attribution,
		Action:         verifierCreateScope,
		ResourceType:   "verification",
		ResourceID:     verificationResult.VerificationID,
		RequestID:      httpx.RequestIDFromContext(r.Context()),
		IdempotencyKey: verificationRequest.IdempotencyKey,
		Outcome:        "succeeded",
		OccurredAt:     placeholderEvaluatedAt,
		ServiceName:    "verifier-api",
	})

	w.Header().Set("Location", "/v1/verifier/verifications/"+verificationResult.VerificationID)
	httpx.WriteJSON(w, http.StatusCreated, verificationResultPayload{
		VerificationID:   verificationResult.VerificationID,
		CredentialID:     verificationRequest.CredentialID,
		IssuerID:         issuerRecord.IssuerID,
		Decision:         string(verificationResult.Decision),
		ReasonCodes:      verificationResult.ReasonCodes,
		EvaluatedAt:      verificationResult.EvaluatedAt,
		CredentialStatus: string(verificationResult.CredentialStatus),
	})
}

func (h *phase1VerifierHandler) getVerification(w http.ResponseWriter, r *http.Request) {
	attribution, ok := h.requireVerifierAttribution(w, r, verifierReadScope)
	if !ok {
		return
	}

	requestRecord, err := h.requests.GetVerificationRequestRecord(r.Context(), r.PathValue("verificationId"))
	if err != nil {
		if errors.Is(err, phase1.ErrRecordNotFound) {
			httpx.WriteError(w, r.Context(), http.StatusNotFound, "verification_not_found", "verification record not found")
			return
		}

		httpx.WriteError(w, r.Context(), http.StatusInternalServerError, "persistence_error", "verification request could not be loaded")
		return
	}

	resultRecord, err := h.results.GetVerificationResultRecord(r.Context(), requestRecord.VerificationID)
	if err != nil {
		if errors.Is(err, phase1.ErrRecordNotFound) {
			httpx.WriteError(w, r.Context(), http.StatusNotFound, "verification_not_found", "verification result not found")
			return
		}

		httpx.WriteError(w, r.Context(), http.StatusInternalServerError, "persistence_error", "verification result could not be loaded")
		return
	}

	_ = h.audits.AppendAuditRecord(r.Context(), phase1.AuditRecord{
		AuditID:        verifierAuditIDForAction(r, "read"),
		Actor:          attribution,
		Action:         verifierReadScope,
		ResourceType:   "verification",
		ResourceID:     resultRecord.VerificationID,
		RequestID:      httpx.RequestIDFromContext(r.Context()),
		IdempotencyKey: strings.TrimSpace(r.Header.Get("Idempotency-Key")),
		Outcome:        "succeeded",
		OccurredAt:     placeholderEvaluatedAt,
		ServiceName:    "verifier-api",
	})

	httpx.WriteJSON(w, http.StatusOK, verificationResultPayload{
		VerificationID:   resultRecord.VerificationID,
		CredentialID:     requestRecord.CredentialID,
		IssuerID:         placeholderVerificationDID,
		Decision:         string(resultRecord.Decision),
		ReasonCodes:      resultRecord.ReasonCodes,
		EvaluatedAt:      resultRecord.EvaluatedAt,
		CredentialStatus: string(resultRecord.CredentialStatus),
	})
}

func (h *phase1VerifierHandler) requireVerifierAttribution(w http.ResponseWriter, r *http.Request, scope string) (authctx.Attribution, bool) {
	attribution, err := h.verifierExtractor.VerifierIntegratorFromRequest(r)
	if err != nil {
		httpx.WriteError(w, r.Context(), http.StatusUnauthorized, "unauthenticated", "authenticated verifier attribution is required")
		return authctx.Attribution{}, false
	}

	if err := authctx.RequireScope(attribution, scope); err != nil {
		httpx.WriteError(w, r.Context(), http.StatusForbidden, "insufficient_scope", "verifier scope is required for this action")
		return authctx.Attribution{}, false
	}

	return attribution, true
}

func (p verificationSubmissionRequestPayload) validate() error {
	switch {
	case strings.TrimSpace(p.PolicyID) == "":
		return errors.New("policyId must not be empty")
	case strings.TrimSpace(p.CredentialArtifact.Kind) == "":
		return errors.New("credentialArtifact.kind must not be empty")
	case strings.TrimSpace(p.CredentialArtifact.MediaType) == "":
		return errors.New("credentialArtifact.mediaType must not be empty")
	case strings.TrimSpace(p.CredentialArtifact.Value) == "":
		return errors.New("credentialArtifact.value must not be empty")
	default:
		return nil
	}
}

func verifierAuditIDForAction(r *http.Request, action string) string {
	return httpx.RequestIDFromContext(r.Context()) + ":" + action
}
