package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Emiloart/HDIP/packages/go/foundation/authctx"
	"github.com/Emiloart/HDIP/packages/go/foundation/httpx"
	"github.com/Emiloart/HDIP/services/verifier-api/internal/config"
	phase1 "github.com/Emiloart/HDIP/services/verifier-api/internal/phase1"
)

const (
	verifierCreateScope         = "verifier.requests.create"
	verifierReadScope           = "verifier.results.read"
	defaultPhase1IssuerID       = "did:web:issuer.hdip.dev"
	credentialArtifactKind      = "phase1_opaque_artifact"
	credentialArtifactMediaType = "application/vnd.hdip.phase1-opaque-artifact"
	credentialArtifactPrefix    = "opaque-artifact:v1:"
	responseVersion             = "2026.04"
)

var verificationBaseTime = time.Date(2026, time.April, 20, 9, 5, 0, 0, time.UTC)

type phase1VerifierHandler struct {
	verifierExtractor authctx.VerifierIntegratorExtractor
	store             *phase1.RuntimeStore
	trustRuntimeReady phase1ReadinessProbe
	trusts            phase1.TrustReadRepository
	credentials       phase1.CredentialRecordRepository
	requests          phase1.VerificationRequestRepository
	results           phase1.VerificationResultRepository
	audits            phase1.AuditRecordRepository
}

type phase1ReadinessProbe interface {
	Check(context.Context) error
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

type credentialArtifactEnvelope struct {
	CredentialID string `json:"credentialId"`
	IssuerID     string `json:"issuerId"`
	TemplateID   string `json:"templateId"`
	ExpiresAt    string `json:"expiresAt"`
}

func newPhase1VerifierHandler(cfg config.Config) (*phase1VerifierHandler, error) {
	store, err := phase1.OpenStore(phase1.StoreOptions{
		DatabaseDriver: cfg.Phase1DatabaseDriver,
		DatabaseURL:    cfg.Phase1DatabaseURL,
	})
	if err != nil {
		return nil, err
	}

	tokenSource, err := phase1.NewHydraClientCredentialsTokenSource(
		cfg.TrustRuntimeHydraTokenURL,
		cfg.TrustRuntimeHydraClientID,
		cfg.TrustRuntimeHydraClientSecret,
		cfg.TrustRuntimeHydraScope,
		&http.Client{Timeout: cfg.RequestTimeout},
	)
	if err != nil {
		_ = store.Close()
		return nil, err
	}

	trusts, err := phase1.NewTrustReadClient(cfg.TrustRegistryBaseURL, tokenSource, &http.Client{
		Timeout: cfg.RequestTimeout,
	})
	if err != nil {
		_ = store.Close()
		return nil, err
	}

	return newPhase1VerifierHandlerWithStoreAndTrustAndReadiness(store, trusts, tokenSource), nil
}

func newPhase1VerifierHandlerWithStore(store *phase1.RuntimeStore) *phase1VerifierHandler {
	return newPhase1VerifierHandlerWithStoreAndTrustAndReadiness(store, store, nil)
}

func newPhase1VerifierHandlerWithStoreAndTrust(
	store *phase1.RuntimeStore,
	trusts phase1.TrustReadRepository,
) *phase1VerifierHandler {
	return newPhase1VerifierHandlerWithStoreAndTrustAndReadiness(store, trusts, nil)
}

func newPhase1VerifierHandlerWithStoreAndTrustAndReadiness(
	store *phase1.RuntimeStore,
	trusts phase1.TrustReadRepository,
	trustRuntimeReady phase1ReadinessProbe,
) *phase1VerifierHandler {
	if trusts == nil {
		trusts = store
	}

	return &phase1VerifierHandler{
		verifierExtractor: authctx.HeaderVerifierIntegratorExtractor{},
		store:             store,
		trustRuntimeReady: trustRuntimeReady,
		trusts:            trusts,
		credentials:       store,
		requests:          store,
		results:           store,
		audits:            store,
	}
}

func (h *phase1VerifierHandler) readiness(ctx context.Context) error {
	if err := h.store.CheckReadiness(ctx, true); err != nil {
		return err
	}
	if h.trustRuntimeReady != nil {
		if err := h.trustRuntimeReady.Check(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (h *phase1VerifierHandler) runtimeMode() string {
	return h.store.RuntimeMode()
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

	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	requestFingerprint := verificationSubmissionFingerprint(request)
	reservationOpen := false
	if idempotencyKey != "" {
		if !h.reserveVerificationIdempotency(w, r, attribution, idempotencyKey, requestFingerprint) {
			return
		}
		reservationOpen = true
		defer func() {
			if reservationOpen {
				_ = h.store.ReleaseIdempotencyRecord(
					r.Context(),
					verifierCreateScope,
					attribution.OrganizationID,
					attribution.PrincipalID,
					string(attribution.ActorType),
					idempotencyKey,
				)
			}
		}()
	}

	verificationID, err := h.requests.NextVerificationID(r.Context())
	if err != nil {
		httpx.WriteError(w, r.Context(), http.StatusInternalServerError, "persistence_error", "verification identifier could not be allocated")
		return
	}

	evaluatedAt := evaluatedAtForVerificationID(verificationID)
	artifactEnvelope, err := parseCredentialArtifact(request.CredentialArtifact)
	if err != nil {
		httpx.WriteError(w, r.Context(), http.StatusBadRequest, "invalid_request", "credentialArtifact.value must match the deterministic Phase 1 opaque artifact format")
		return
	}

	submittedArtifactDigest := artifactDigest(request.CredentialArtifact.Value)
	verificationRequest := phase1.VerificationRequestRecord{
		VerificationID:            verificationID,
		VerifierID:                attribution.OrganizationID,
		SubmittedCredentialDigest: submittedArtifactDigest,
		CredentialID:              artifactEnvelope.CredentialID,
		PolicyID:                  request.PolicyID,
		RequestedAt:               evaluatedAt,
		Actor:                     attribution,
		IdempotencyKey:            idempotencyKey,
	}

	if err := h.requests.CreateVerificationRequestRecord(r.Context(), verificationRequest); err != nil {
		httpx.WriteError(w, r.Context(), http.StatusInternalServerError, "persistence_error", "verification request could not be stored")
		return
	}

	credentialRecord, resolvedByDigest, err := h.resolveCredentialRecord(r.Context(), verificationRequest, artifactEnvelope)
	if err != nil {
		_ = h.audits.AppendAuditRecord(r.Context(), phase1.AuditRecord{
			AuditID:        verifierAuditIDForAction(r, "create"),
			Actor:          attribution,
			Action:         verifierCreateScope,
			ResourceType:   "verification",
			ResourceID:     verificationRequest.VerificationID,
			RequestID:      httpx.RequestIDFromContext(r.Context()),
			IdempotencyKey: verificationRequest.IdempotencyKey,
			Outcome:        "failed",
			OccurredAt:     evaluatedAt,
			ServiceName:    "verifier-api",
		})
		errorPayload := httpx.ErrorEnvelope{
			Error: httpx.ErrorDetail{
				Code:      "credential_not_found",
				Message:   "submitted credential artifact did not resolve to a known credential",
				RequestID: httpx.RequestIDFromContext(r.Context()),
			},
		}
		if idempotencyKey != "" {
			if storeErr := h.completeVerificationIdempotencyResponse(
				r.Context(),
				attribution,
				idempotencyKey,
				requestFingerprint,
				http.StatusNotFound,
				verificationRequest.VerificationID,
				"",
				errorPayload,
				evaluatedAt,
			); storeErr != nil {
				httpx.WriteError(w, r.Context(), http.StatusInternalServerError, "persistence_error", "verification replay state could not be stored")
				return
			}
			reservationOpen = false
		}
		httpx.WriteError(w, r.Context(), http.StatusNotFound, "credential_not_found", "submitted credential artifact did not resolve to a known credential")
		return
	}

	verificationResult := h.evaluateVerification(
		r.Context(),
		request,
		artifactEnvelope,
		credentialRecord,
		resolvedByDigest,
		verificationRequest.VerificationID,
		evaluatedAt,
	)

	if err := h.results.CreateVerificationResultRecord(r.Context(), verificationResult); err != nil {
		httpx.WriteError(w, r.Context(), http.StatusInternalServerError, "persistence_error", "verification result could not be stored")
		return
	}

	response := verificationResultPayload{
		VerificationID:   verificationResult.VerificationID,
		CredentialID:     verificationRequest.CredentialID,
		IssuerID:         verificationResult.IssuerID,
		Decision:         string(verificationResult.Decision),
		ReasonCodes:      verificationResult.ReasonCodes,
		EvaluatedAt:      verificationResult.EvaluatedAt,
		CredentialStatus: string(verificationResult.CredentialStatus),
	}
	location := "/v1/verifier/verifications/" + verificationResult.VerificationID
	if idempotencyKey != "" {
		if err := h.completeVerificationIdempotencyResponse(
			r.Context(),
			attribution,
			idempotencyKey,
			requestFingerprint,
			http.StatusCreated,
			verificationResult.VerificationID,
			location,
			response,
			evaluatedAt,
		); err != nil {
			httpx.WriteError(w, r.Context(), http.StatusInternalServerError, "persistence_error", "verification replay state could not be stored")
			return
		}
		reservationOpen = false
	}

	_ = h.audits.AppendAuditRecord(r.Context(), phase1.AuditRecord{
		AuditID:        verifierAuditIDForAction(r, "create"),
		Actor:          attribution,
		Action:         verifierCreateScope,
		ResourceType:   "verification",
		ResourceID:     verificationResult.VerificationID,
		RequestID:      httpx.RequestIDFromContext(r.Context()),
		IdempotencyKey: verificationRequest.IdempotencyKey,
		Outcome:        auditOutcomeForDecision(verificationResult.Decision),
		OccurredAt:     evaluatedAt,
		ServiceName:    "verifier-api",
	})

	w.Header().Set("Location", location)
	httpx.WriteJSON(w, http.StatusCreated, response)
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

	if requestRecord.VerifierID != attribution.OrganizationID {
		httpx.WriteError(w, r.Context(), http.StatusNotFound, "verification_not_found", "verification record not found")
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
		OccurredAt:     time.Now().UTC(),
		ServiceName:    "verifier-api",
	})

	httpx.WriteJSON(w, http.StatusOK, verificationResultPayload{
		VerificationID:   resultRecord.VerificationID,
		CredentialID:     requestRecord.CredentialID,
		IssuerID:         resultRecord.IssuerID,
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
	case strings.TrimSpace(p.CredentialArtifact.Kind) != credentialArtifactKind:
		return errors.New("credentialArtifact.kind must match the deterministic Phase 1 opaque artifact kind")
	case strings.TrimSpace(p.CredentialArtifact.MediaType) != credentialArtifactMediaType:
		return errors.New("credentialArtifact.mediaType must match the deterministic Phase 1 opaque artifact media type")
	case strings.TrimSpace(p.CredentialArtifact.Value) == "":
		return errors.New("credentialArtifact.value must not be empty")
	default:
		return nil
	}
}

func verifierAuditIDForAction(r *http.Request, action string) string {
	return httpx.RequestIDFromContext(r.Context()) + ":" + action
}

func (h *phase1VerifierHandler) resolveCredentialRecord(
	ctx context.Context,
	requestRecord phase1.VerificationRequestRecord,
	artifactEnvelope credentialArtifactEnvelope,
) (phase1.CredentialRecord, bool, error) {
	credentialRecord, err := h.credentials.GetCredentialRecordByArtifactDigest(ctx, requestRecord.SubmittedCredentialDigest)
	if err == nil {
		return credentialRecord, true, nil
	}

	credentialRecord, err = h.credentials.GetCredentialRecord(ctx, artifactEnvelope.CredentialID)
	if err == nil {
		return credentialRecord, false, nil
	}

	return phase1.CredentialRecord{}, false, err
}

func (h *phase1VerifierHandler) evaluateVerification(
	ctx context.Context,
	request verificationSubmissionRequestPayload,
	artifactEnvelope credentialArtifactEnvelope,
	credentialRecord phase1.CredentialRecord,
	resolvedByDigest bool,
	verificationID string,
	evaluatedAt time.Time,
) phase1.VerificationResultRecord {
	credentialStatus := credentialStatusForEvaluation(credentialRecord, evaluatedAt)
	result := phase1.VerificationResultRecord{
		VerificationID:   verificationID,
		IssuerID:         artifactEnvelope.IssuerID,
		Decision:         phase1.VerificationDecisionDeny,
		ReasonCodes:      []string{"artifact_continuity_failed"},
		CredentialStatus: credentialStatus,
		EvaluatedAt:      evaluatedAt,
		ResponseVersion:  responseVersion,
	}

	if request.CredentialID != "" && strings.TrimSpace(request.CredentialID) != artifactEnvelope.CredentialID {
		result.ReasonCodes = []string{"credential_id_mismatch"}
		return result
	}

	if !resolvedByDigest ||
		artifactEnvelope.CredentialID != credentialRecord.CredentialID ||
		artifactEnvelope.IssuerID != credentialRecord.IssuerID ||
		artifactEnvelope.TemplateID != credentialRecord.TemplateID ||
		artifactEnvelope.ExpiresAt != credentialRecord.ExpiresAt.UTC().Format(time.RFC3339) {
		return result
	}

	result.IssuerID = credentialRecord.IssuerID
	trustRecord, err := h.trusts.GetIssuerTrustRecord(ctx, credentialRecord.IssuerID)
	if err != nil {
		if errors.Is(err, phase1.ErrTrustRuntimeUnavailable) {
			result.IssuerTrustState = "unavailable"
		} else {
			result.IssuerTrustState = "missing"
		}
		result.ReasonCodes = []string{"issuer_not_trusted"}
		return result
	}

	result.IssuerTrustState = trustRecord.TrustState
	switch strings.TrimSpace(trustRecord.TrustState) {
	case "active":
	case "suspended":
		result.ReasonCodes = []string{"issuer_suspended"}
		return result
	default:
		result.ReasonCodes = []string{"issuer_not_trusted"}
		return result
	}

	switch credentialStatus {
	case phase1.CredentialStatusSnapshotRevoked:
		result.ReasonCodes = []string{"credential_status_revoked"}
		return result
	case phase1.CredentialStatusSnapshotSuperseded:
		result.ReasonCodes = []string{"credential_status_superseded"}
		return result
	case phase1.CredentialStatusSnapshotExpired:
		result.ReasonCodes = []string{"credential_expired"}
		return result
	}

	if !isPolicyCompatible(request.PolicyID, credentialRecord.TemplateID) {
		result.ReasonCodes = []string{"template_mismatch"}
		return result
	}

	result.Decision = phase1.VerificationDecisionAllow
	result.ReasonCodes = []string{"issuer_trusted", "credential_status_active", "template_match"}
	return result
}

func parseCredentialArtifact(artifact verifierCredentialArtifactPayload) (credentialArtifactEnvelope, error) {
	rawValue := strings.TrimSpace(artifact.Value)
	if !strings.HasPrefix(rawValue, credentialArtifactPrefix) {
		return credentialArtifactEnvelope{}, errors.New("artifact prefix mismatch")
	}

	rawEnvelope, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(rawValue, credentialArtifactPrefix))
	if err != nil {
		return credentialArtifactEnvelope{}, err
	}

	var envelope credentialArtifactEnvelope
	if err := json.Unmarshal(rawEnvelope, &envelope); err != nil {
		return credentialArtifactEnvelope{}, err
	}

	switch {
	case strings.TrimSpace(envelope.CredentialID) == "":
		return credentialArtifactEnvelope{}, errors.New("artifact credentialId missing")
	case strings.TrimSpace(envelope.IssuerID) == "":
		return credentialArtifactEnvelope{}, errors.New("artifact issuerId missing")
	case strings.TrimSpace(envelope.TemplateID) == "":
		return credentialArtifactEnvelope{}, errors.New("artifact templateId missing")
	case !isRFC3339(envelope.ExpiresAt):
		return credentialArtifactEnvelope{}, errors.New("artifact expiresAt invalid")
	default:
		return envelope, nil
	}
}

func materializeCredentialArtifactValue(credentialID, issuerID, templateID string, expiresAt time.Time) string {
	envelope := credentialArtifactEnvelope{
		CredentialID: credentialID,
		IssuerID:     issuerID,
		TemplateID:   templateID,
		ExpiresAt:    expiresAt.UTC().Format(time.RFC3339),
	}

	raw, err := json.Marshal(envelope)
	if err != nil {
		panic(err)
	}

	return credentialArtifactPrefix + base64.RawURLEncoding.EncodeToString(raw)
}

func artifactDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func evaluatedAtForVerificationID(verificationID string) time.Time {
	parts := strings.Split(verificationID, "_")
	if len(parts) == 0 {
		return verificationBaseTime
	}

	sequence, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil || sequence <= 0 {
		return verificationBaseTime
	}

	return verificationBaseTime.Add(time.Duration(sequence-1) * time.Minute)
}

func credentialStatusForEvaluation(record phase1.CredentialRecord, evaluatedAt time.Time) phase1.CredentialStatusSnapshot {
	switch record.Status {
	case phase1.CredentialStatusSnapshotRevoked:
		return phase1.CredentialStatusSnapshotRevoked
	case phase1.CredentialStatusSnapshotSuperseded:
		return phase1.CredentialStatusSnapshotSuperseded
	}

	if !record.ExpiresAt.IsZero() && evaluatedAt.After(record.ExpiresAt) {
		return phase1.CredentialStatusSnapshotExpired
	}

	return phase1.CredentialStatusSnapshotActive
}

func isPolicyCompatible(policyID string, templateID string) bool {
	return strings.TrimSpace(policyID) == defaultPolicyID && strings.TrimSpace(templateID) == defaultTemplateID
}

func isRFC3339(value string) bool {
	_, err := time.Parse(time.RFC3339, value)
	return err == nil
}

func auditOutcomeForDecision(decision phase1.VerificationDecision) string {
	if decision == phase1.VerificationDecisionDeny {
		return "denied"
	}

	return "succeeded"
}

func (h *phase1VerifierHandler) reserveVerificationIdempotency(
	w http.ResponseWriter,
	r *http.Request,
	attribution authctx.Attribution,
	idempotencyKey string,
	requestFingerprint string,
) bool {
	return h.reserveIdempotencyRequest(
		w,
		r,
		attribution,
		verifierCreateScope,
		idempotencyKey,
		requestFingerprint,
		"verification",
	)
}

func (h *phase1VerifierHandler) reserveIdempotencyRequest(
	w http.ResponseWriter,
	r *http.Request,
	attribution authctx.Attribution,
	operation string,
	idempotencyKey string,
	requestFingerprint string,
	resourceType string,
) bool {
	result, err := h.store.ReserveIdempotencyRecord(
		r.Context(),
		phase1.IdempotencyRecord{
			Operation:            operation,
			CallerPrincipalID:    attribution.PrincipalID,
			CallerOrganizationID: attribution.OrganizationID,
			CallerActorType:      string(attribution.ActorType),
			IdempotencyKey:       idempotencyKey,
			RequestFingerprint:   requestFingerprint,
			ResourceType:         resourceType,
			CreatedAt:            time.Now().UTC(),
			UpdatedAt:            time.Now().UTC(),
		},
	)
	if err != nil {
		httpx.WriteError(w, r.Context(), http.StatusInternalServerError, "persistence_error", "idempotency state could not be loaded")
		return false
	}

	switch result.Outcome {
	case phase1.IdempotencyReservationReserved:
		return true
	case phase1.IdempotencyReservationReplay:
		record := result.Record
		_ = h.audits.AppendAuditRecord(r.Context(), phase1.AuditRecord{
			AuditID:        verifierAuditIDForAction(r, "idempotency-replay"),
			Actor:          attribution,
			Action:         operation,
			ResourceType:   resourceType,
			ResourceID:     record.ResourceID,
			RequestID:      httpx.RequestIDFromContext(r.Context()),
			IdempotencyKey: idempotencyKey,
			Outcome:        "replayed",
			OccurredAt:     time.Now().UTC(),
			ServiceName:    "verifier-api",
		})
		if record.Location != "" {
			w.Header().Set("Location", record.Location)
		}
		writeStoredJSON(w, record.ResponseStatusCode, record.ResponseBody)
		return false
	case phase1.IdempotencyReservationInProgress:
		record := result.Record
		_ = h.audits.AppendAuditRecord(r.Context(), phase1.AuditRecord{
			AuditID:        verifierAuditIDForAction(r, "idempotency-in-progress"),
			Actor:          attribution,
			Action:         operation,
			ResourceType:   resourceType,
			ResourceID:     record.ResourceID,
			RequestID:      httpx.RequestIDFromContext(r.Context()),
			IdempotencyKey: idempotencyKey,
			Outcome:        "failed",
			OccurredAt:     time.Now().UTC(),
			ServiceName:    "verifier-api",
		})
		httpx.WriteError(w, r.Context(), http.StatusConflict, "idempotency_in_progress", "Idempotency-Key is already reserved by an in-flight Phase 1 request")
		return false
	default:
		record := result.Record
		_ = h.audits.AppendAuditRecord(r.Context(), phase1.AuditRecord{
			AuditID:        verifierAuditIDForAction(r, "idempotency-conflict"),
			Actor:          attribution,
			Action:         operation,
			ResourceType:   resourceType,
			ResourceID:     record.ResourceID,
			RequestID:      httpx.RequestIDFromContext(r.Context()),
			IdempotencyKey: idempotencyKey,
			Outcome:        "failed",
			OccurredAt:     time.Now().UTC(),
			ServiceName:    "verifier-api",
		})
		httpx.WriteError(w, r.Context(), http.StatusConflict, "idempotency_conflict", "Idempotency-Key is already bound to a different Phase 1 request")
		return false
	}
}

func (h *phase1VerifierHandler) completeVerificationIdempotencyResponse(
	ctx context.Context,
	attribution authctx.Attribution,
	idempotencyKey string,
	requestFingerprint string,
	statusCode int,
	verificationID string,
	location string,
	response any,
	createdAt time.Time,
) error {
	rawResponse, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return h.store.CompleteIdempotencyRecord(ctx, phase1.IdempotencyRecord{
		Operation:            verifierCreateScope,
		CallerPrincipalID:    attribution.PrincipalID,
		CallerOrganizationID: attribution.OrganizationID,
		CallerActorType:      string(attribution.ActorType),
		IdempotencyKey:       idempotencyKey,
		RequestFingerprint:   requestFingerprint,
		State:                phase1.IdempotencyStateCompleted,
		ResponseStatusCode:   statusCode,
		ResourceType:         "verification",
		ResourceID:           verificationID,
		Location:             location,
		ResponseBody:         rawResponse,
		CreatedAt:            createdAt,
		UpdatedAt:            createdAt,
	})
}

func verificationSubmissionFingerprint(request verificationSubmissionRequestPayload) string {
	return fingerprintValue(struct {
		PolicyID                 string `json:"policyId"`
		CredentialID             string `json:"credentialId"`
		CredentialArtifactKind   string `json:"credentialArtifactKind"`
		CredentialArtifactDigest string `json:"credentialArtifactDigest"`
		CredentialArtifactType   string `json:"credentialArtifactMediaType"`
	}{
		PolicyID:                 strings.TrimSpace(request.PolicyID),
		CredentialID:             strings.TrimSpace(request.CredentialID),
		CredentialArtifactKind:   strings.TrimSpace(request.CredentialArtifact.Kind),
		CredentialArtifactDigest: artifactDigest(strings.TrimSpace(request.CredentialArtifact.Value)),
		CredentialArtifactType:   strings.TrimSpace(request.CredentialArtifact.MediaType),
	})
}

func fingerprintValue(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func writeStoredJSON(w http.ResponseWriter, statusCode int, payload json.RawMessage) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_, _ = w.Write(payload)
}

func defaultVerifierIssuerRecord() phase1.IssuerRecord {
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

func defaultVerifierCredentialRecord() phase1.CredentialRecord {
	expiresAt := time.Date(2027, time.April, 20, 9, 0, 0, 0, time.UTC)
	value := materializeCredentialArtifactValue("cred_hdip_passport_basic_001", defaultPhase1IssuerID, defaultTemplateID, expiresAt)
	return phase1.CredentialRecord{
		CredentialID:   "cred_hdip_passport_basic_001",
		IssuerID:       defaultPhase1IssuerID,
		TemplateID:     defaultTemplateID,
		ArtifactDigest: artifactDigest(value),
		ExpiresAt:      expiresAt,
		Status:         phase1.CredentialStatusSnapshotActive,
	}
}
